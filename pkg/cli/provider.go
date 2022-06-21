package cli

import (
	"log"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/olxbr/network-api/pkg/client"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/spf13/cobra"
)

func newProviderCommand() *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Provider operations",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig()
			if err != nil {
				log.Printf("Is your config file correctly created?")
				return err
			}
			ctx := cmd.Context()
			cmd.SetContext(WithConfig(ctx, cfg))
			path := filepath.Join(os.Getenv("HOME"), configPathDefault)
			auth, err := client.NewOAuth2Authorizer(&client.OAuth2AuthorizerOptions{
				ClientID: cfg.ClientID,
				Issuer:   cfg.IssuerURL,
				Scopes:   cfg.Scopes,
				TokenDir: path,
			})
			if err != nil {
				log.Printf("error: %+v", err)
				return err
			}
			defer auth.Close()
			t, err := auth.GetToken(ctx)
			if err != nil {
				log.Printf("error: %+v", err)
				return err
			}
			httpClient := auth.NewClient(ctx, t)

			cmd.SetContext(client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: cfg.Endpoint,
				Client:   httpClient,
			}))
			return nil
		},
	}

	providerCmd.AddCommand(providerAddCmd())
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerListCmd)

	return providerCmd
}

func providerAddCmd() *cobra.Command {
	req := &types.ProviderRequest{}

	c := &cobra.Command{
		Use:   "add",
		Short: "Adds a new provider",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return
			}

			n, err := cli.CreateProvider(ctx, req)
			if err != nil {
				log.Printf("error creating network: %+v", err)
				return
			}

			log.Printf("Network: %+v", n)
		},
	}

	f := c.Flags()
	f.StringVar(&req.Name, "name", "", "Name")
	f.StringVar(&req.WebhookURL, "url", "", "Webhook URL")
	f.StringVar(&req.APIToken, "token", "", "API Token")

	return c
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Configure remote endpoint",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Configure remote endpoint",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cli, ok := client.ClientFromContext(ctx)
		if !ok {
			log.Printf("error retriving client")
			return
		}
		ps, err := cli.ListProviders(ctx)
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Provider", "URL"})

		for _, p := range ps.Items {
			table.Append([]string{
				p.Name,
				p.WebhookURL,
			})
		}

		table.Render()
	},
}
