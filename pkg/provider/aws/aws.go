package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	ktypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/go-playground/validator/v10"
	"github.com/olxbr/network-api/pkg/types"
)

var validate *validator.Validate

func apiGatewayError(err error, code int) events.APIGatewayProxyResponse {
	response := events.APIGatewayProxyResponse{}
	j, _ := json.Marshal(types.NewSingleErrorResponse(err.Error()))
	response.Body = string(j)
	response.StatusCode = code
	return response
}

func apiGatewayResponse(i interface{}, code int) events.APIGatewayProxyResponse {
	response := events.APIGatewayProxyResponse{}
	j, _ := json.Marshal(i)
	response.Body = string(j)
	response.StatusCode = code
	return response
}

func validateKMSToken(ctx context.Context, key, token string) bool {
	parts := strings.Split(token, " ")
	if len(parts) != 2 {
		return false
	}

	signedMessage := strings.Split(parts[1], ".")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	cli := kms.NewFromConfig(cfg)

	vo, err := cli.Verify(ctx, &kms.VerifyInput{
		KeyId:            aws.String(key),
		Message:          []byte(signedMessage[0]),
		Signature:        []byte(signedMessage[1]),
		SigningAlgorithm: ktypes.SigningAlgorithmSpecRsassaPssSha256,
	})
	if err != nil {
		return false
	}

	return vo.SignatureValid
}

func LambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	validate = validator.New()

	var auth string
	var ok bool
	if auth, ok = request.Headers["authorization"]; !ok {
		e := fmt.Errorf("unauthorized")
		return apiGatewayError(e, 401), nil
	}
	if !validateKMSToken(ctx, os.Getenv("SIGNING_KEY"), auth) {
		e := fmt.Errorf("unauthorized")
		return apiGatewayError(e, 401), nil
	}

	if len(request.Body) == 0 {
		e := fmt.Errorf("empty body")
		return apiGatewayError(e, 400), nil
	}

	dec64 := base64.NewDecoder(base64.StdEncoding, strings.NewReader(request.Body))
	d, err := ioutil.ReadAll(dec64)
	if err != nil {
		e := fmt.Errorf("error decoding base64: %s", err)
		return apiGatewayError(e, 400), nil
	}

	webhook := &types.ProviderWebhook{}
	err = json.Unmarshal(d, webhook)
	if err != nil {
		return apiGatewayError(err, 400), nil
	}

	err = validate.Struct(webhook)
	if err != nil {
		return apiGatewayError(err, http.StatusBadRequest), nil
	}

	cfg := AssumeRole(webhook.Account, webhook.Region, os.Getenv("TRUST_ROLE"))

	switch webhook.Event {
	case types.CreateNetwork:
		resp, err := CreateNetwork(ctx, cfg, webhook)
		if err != nil {
			return apiGatewayError(err, 500), err
		}
		return apiGatewayResponse(&types.ProviderWebhookResponse{
			StatusCode: 200,
			ID:         resp,
		}, 200), nil
	}

	return apiGatewayResponse("{\"message\": \"success\"}", 200), nil
}

func CreateNetwork(ctx context.Context, cfg aws.Config, pw *types.ProviderWebhook) (string, error) {
	cli := cloudformation.NewFromConfig(cfg)

	templates := os.Getenv("CF_REPOSITORY")
	log.Printf("Templates: %s", templates)

	d := NewDeployer(cli)

	stackName := fmt.Sprintf("network-%s", pw.NetworkID)
	cs, err := d.CreateStack(ctx, &DeployerInput{
		StackName:   stackName,
		NetworkID:   pw.NetworkID,
		Params:      BuildParameters(pw),
		TemplateURL: fmt.Sprintf("%s/%s", templates, "template.yaml"),
	})
	if err != nil {
		log.Printf("error creating stack: %+v", err)
		return "", err
	}
	return cs.ID, nil
}

func CreateNetworkWithChangeSet(ctx context.Context, cfg aws.Config, pw *types.ProviderWebhook) (string, error) {
	cli := cloudformation.NewFromConfig(cfg)

	templates := os.Getenv("CF_REPOSITORY")
	log.Printf("Templates: %s", templates)

	d := NewDeployer(cli)

	stackName := fmt.Sprintf("network-%s", pw.NetworkID)
	csr, err := d.CreateChangeSet(ctx, &DeployerInput{
		StackName:   stackName,
		Params:      BuildParameters(pw),
		TemplateURL: fmt.Sprintf("%s/%s", templates, "template.yaml"),
	})
	if err != nil {
		log.Printf("error creating changeset: %+v", err)
		return "", err
	}

	log.Printf("changeSet created: %s", csr.ID)
	err = d.WaitChangeSet(ctx, stackName, csr.ID)
	if err != nil {
		log.Printf("error waiting for changeset: %+v", err)
		return "", err
	}

	err = d.ExecuteChangeSet(ctx, stackName, csr.ID)
	if err != nil {
		log.Printf("error executing for changeset: %+v", err)
		return "", err
	}
	return "", nil
}

func BuildParameters(pw *types.ProviderWebhook) []cftypes.Parameter {
	params := []cftypes.Parameter{
		{
			ParameterKey:   aws.String("VPCName"),
			ParameterValue: aws.String(pw.NetworkID),
		},
		{
			ParameterKey:   aws.String("VPCCidr"),
			ParameterValue: aws.String(pw.CIDR),
		},
		{
			ParameterKey:   aws.String("Environment"),
			ParameterValue: aws.String(pw.Environment),
		},
	}

	private := 0
	public := 0
	tgw := 0
	for _, s := range pw.Subnets {
		var key string
		switch s.Type {
		case types.Private:
			key = fmt.Sprintf("PrivateSubnet%dCidr", private)
			private++
		case types.Public:
			key = fmt.Sprintf("PublicSubnet%dCidr", public)
			public++
		case types.TransitGateway:
			key = fmt.Sprintf("TGWSubnet%dCidr", tgw)
			tgw++
		}
		p := cftypes.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(s.CIDR),
		}
		params = append(params, p)
	}
	return params
}
