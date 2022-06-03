package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func AssumeRole(account, region, roleName string) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	cli := sts.NewFromConfig(cfg)

	roleARN := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, roleName)

	cfg.Region = region
	cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(cli, roleARN))
	return cfg
}
