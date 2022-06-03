package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

// Inspired by: https://github.com/aws/aws-cli/blob/develop/awscli/customizations/cloudformation/deployer.py

type Deployer struct {
	Client *cloudformation.Client
}

type DeployerInput struct {
	StackName   string
	NetworkID   string
	Params      []cftypes.Parameter
	TemplateURL string
}

type ChangeSetResult struct {
	ID   string
	Type cftypes.ChangeSetType
}

type StackResult struct {
	ID string
}

func NewDeployer(cli *cloudformation.Client) *Deployer {
	return &Deployer{
		Client: cli,
	}
}

func (d *Deployer) HasStack(ctx context.Context, name string) (bool, error) {
	do, err := d.Client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if strings.Contains(ae.ErrorMessage(), name+" does not exist") {
				return false, nil
			}
		}
		return false, err
	}
	if len(do.Stacks) != 0 {
		return false, nil
	}
	return do.Stacks[0].StackStatus != cftypes.StackStatusReviewInProgress, nil
}

func (d *Deployer) CreateStack(ctx context.Context, input *DeployerInput) (*StackResult, error) {
	cs, err := d.Client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:   aws.String(input.StackName),
		Parameters:  input.Params,
		TemplateURL: aws.String(input.TemplateURL),
		Tags: []cftypes.Tag{
			{Key: aws.String("network-api-managed"), Value: aws.String("true")},
			{Key: aws.String("network-id"), Value: aws.String(input.NetworkID)},
		},
	})
	if err != nil {
		return nil, err
	}
	return &StackResult{
		ID: aws.ToString(cs.StackId),
	}, nil
}

func (d *Deployer) CreateChangeSet(ctx context.Context, input *DeployerInput) (*ChangeSetResult, error) {
	changeSetType := cftypes.ChangeSetTypeUpdate
	hasStack, err := d.HasStack(ctx, input.StackName)
	if err != nil {
		return nil, err
	}
	if !hasStack {
		changeSetType = cftypes.ChangeSetTypeCreate
	}

	changeSetInput := &cloudformation.CreateChangeSetInput{
		ChangeSetName: aws.String(input.StackName),
		StackName:     aws.String(input.StackName),
		ChangeSetType: changeSetType,
		Parameters:    input.Params,
		TemplateURL:   aws.String(input.TemplateURL),
	}

	cso, err := d.Client.CreateChangeSet(ctx, changeSetInput)
	if err != nil {
		return nil, err
	}

	return &ChangeSetResult{
		ID:   aws.ToString(cso.Id),
		Type: changeSetType,
	}, nil
}

func (d *Deployer) CheckChangeSet(ctx context.Context, stackName, ID string) error {
	out, err := d.Client.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{
		StackName:     aws.String(stackName),
		ChangeSetName: aws.String(ID),
	})
	if err != nil {
		return err
	}

	switch out.Status {
	case cftypes.ChangeSetStatusCreateComplete:
		return nil
	case cftypes.ChangeSetStatusFailed:
		return fmt.Errorf("failed to create change set: %s", aws.ToString(out.StatusReason))
	}

	return nil
}

func (d *Deployer) WaitChangeSet(ctx context.Context, stackName, ID string) error {
	waiter := cloudformation.NewChangeSetCreateCompleteWaiter(d.Client, func(csccwo *cloudformation.ChangeSetCreateCompleteWaiterOptions) {

	})
	return waiter.Wait(ctx, &cloudformation.DescribeChangeSetInput{
		StackName:     aws.String(stackName),
		ChangeSetName: aws.String(ID),
	}, 1*time.Minute)
}

func (d *Deployer) ExecuteChangeSet(ctx context.Context, stackName, ID string) error {
	_, err := d.Client.ExecuteChangeSet(ctx, &cloudformation.ExecuteChangeSetInput{
		StackName:     aws.String(stackName),
		ChangeSetName: aws.String(ID),
	})
	return err
}

func (d *Deployer) WaitExecute(ctx context.Context, stackName string, t cftypes.ChangeSetType) error {
	if t == cftypes.ChangeSetTypeCreate {
		waiter := cloudformation.NewStackCreateCompleteWaiter(d.Client)
		return waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}, 1*time.Minute)
	} else if t == cftypes.ChangeSetTypeUpdate {
		waiter := cloudformation.NewStackUpdateCompleteWaiter(d.Client)
		return waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}, 1*time.Minute)
	}

	return fmt.Errorf("unsupported change set type: %s", t)
}
