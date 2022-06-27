package cli

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/olxbr/network-api/pkg/client"
)

func SetupClientContext(ctx context.Context, cfg *Config) (context.Context, error) {
	path := filepath.Join(os.Getenv("HOME"), configPathDefault)

	auth, err := client.NewOAuth2Authorizer(&client.OAuth2AuthorizerOptions{
		ClientID: cfg.ClientID,
		Issuer:   cfg.IssuerURL,
		Scopes:   cfg.Scopes,
		TokenDir: path,
	})
	if err != nil {
		log.Printf("error: %+v", err)
		return nil, err
	}
	defer auth.Close()

	t, err := auth.GetToken(ctx)
	if err != nil {
		log.Printf("error: %+v", err)
		return nil, err
	}

	httpClient := auth.NewClient(ctx, t)
	ctx = client.WithNewClient(ctx, &client.ClientOptions{
		Endpoint: cfg.Endpoint,
		Client:   httpClient,
	})
	return ctx, nil
}
