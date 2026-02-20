package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/daylamtayari/cierge/reservation"
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
	lambda.Start(handle)
}

func handle(ctx context.Context, event reservation.Event) error {
	output := reservation.Handle(ctx, event, &decrypter)

	marshalledOutput, _ := json.Marshal(output)
	fmt.Println(string(marshalledOutput))

	return nil
}
