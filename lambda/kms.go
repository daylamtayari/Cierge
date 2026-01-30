package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type KMSDecrypter struct {
	client *kms.Client
}

func (k *KMSDecrypter) Decrypt(ctx context.Context, encrypted []byte) (string, error) {
	decrypted, err := k.client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: encrypted,
	})

	if err != nil {
		return "", err
	}
	return string(decrypted.Plaintext), nil
}
