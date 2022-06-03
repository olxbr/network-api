package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/olxbr/network-api/pkg/provider/aws"
)

func main() {
	lambda.Start(aws.LambdaHandler)
}
