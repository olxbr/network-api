package secret

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type secrets struct {
	Client    *secretsmanager.Client
	SecretArn string
}

type Secrets interface {
	PutAPIToken(ctx context.Context, provider, token string) error
	GetAPIToken(ctx context.Context, provider string) (string, error)
}

type SecretStorage map[string]string

func New(cli *secretsmanager.Client, arn string) Secrets {
	return &secrets{
		Client:    cli,
		SecretArn: arn,
	}
}

func (s *secrets) PutAPIToken(ctx context.Context, provider, token string) error {
	so, err := s.Client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.SecretArn),
	})
	if err != nil {
		return err
	}

	secretStr := aws.ToString(so.SecretString)
	storage := SecretStorage{}
	err = json.Unmarshal([]byte(secretStr), &storage)
	if err != nil {
		return err
	}

	storage[provider] = token

	secretByte, err := json.Marshal(storage)
	if err != nil {
		return err
	}

	_, err = s.Client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(s.SecretArn),
		SecretString: aws.String(string(secretByte)),
	})
	return err
}

func (s *secrets) GetAPIToken(ctx context.Context, provider string) (string, error) {
	so, err := s.Client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.SecretArn),
	})
	if err != nil {
		return "", err
	}

	secretStr := aws.ToString(so.SecretString)
	storage := SecretStorage{}
	err = json.Unmarshal([]byte(secretStr), &storage)
	if err != nil {
		return "", err
	}

	if token, ok := storage[provider]; ok {
		return token, nil
	}

	return "", nil
}
