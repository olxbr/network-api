package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/olxbr/network-api/pkg/api"
	"github.com/olxbr/network-api/pkg/db"
	"github.com/olxbr/network-api/pkg/secret"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	secretsClient := secretsmanager.NewFromConfig(cfg)

	secretsArn := os.Getenv("SecretsARN")

	d := db.New(dynamoClient)
	s := secret.New(secretsClient, secretsArn)
	mux := api.New(d, s)
	api.NewValidator()
	lambda.Start(httpadapter.New(mux.GetHandler()).ProxyWithContext)
}
