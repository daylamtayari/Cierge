package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

var decrypter KMSDecrypter

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("failed to load aws config: " + err.Error())
	}

	kmsClient := kms.NewFromConfig(cfg)
	decrypter = KMSDecrypter{
		client: kmsClient,
	}
}

func main() {
	lambda.Start(HandleRequest)
}
