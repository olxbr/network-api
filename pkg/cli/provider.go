package cli

import (
	"io"
	"log"

	"github.com/olekukonko/tablewriter"
	"github.com/olxbr/network-api/pkg/client"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/spf13/cobra"
)

func renderProviders(w io.Writer, ps *types.ProviderListResponse) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Provider", "URL"})
	for _, p := range ps.Items {
		table.Append([]string{
			p.Name,
			p.WebhookURL,
		})
	}
	table.Render()
}

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
			ctx, err = SetupClientContext(WithConfig(ctx, cfg), cfg)
			if err != nil {
				return err
			}
			cmd.SetContext(ctx)
			return nil
		},
	}

	providerCmd.AddCommand(providerAddCmd())
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerUpdateCmd())
	providerCmd.AddCommand(providerRemoveCmd)

	return providerCmd
}

func providerAddCmd() *cobra.Command {
	req := &types.ProviderRequest{}

	c := &cobra.Command{
		Use:   "add",
		Short: "Adds a new provider",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return
			}

			req.Name = args[0]

			p, err := cli.CreateProvider(ctx, req)
			if err != nil {
				log.Printf("error creating provider: %+v", err)
				return
			}

			log.Println("Provider:")
			renderProviders(cmd.OutOrStdout(), &types.ProviderListResponse{
				Items: []*types.Provider{p},
			})
		},
	}

	f := c.Flags()
	f.StringVar(&req.WebhookURL, "url", "", "Webhook URL")
	f.StringVar(&req.APIToken, "token", "", "API Token")

	return c
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available providers",
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
		renderProviders(cmd.OutOrStdout(), ps)
	},
}

func providerUpdateCmd() *cobra.Command {
	req := &types.ProviderUpdateRequest{}
	var WebhookURL string
	var APIToken string

	c := &cobra.Command{
		Use:   "update",
		Short: "Updates a provider",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return
			}

			name := args[0]
			req.WebhookURL = types.String(WebhookURL)
			req.APIToken = types.String(APIToken)

			p, err := cli.UpdateProvider(ctx, name, req)
			if err != nil {
				log.Printf("Error: %s", err)
				return
			}

			log.Println("Updated provider:")
			renderProviders(cmd.OutOrStdout(), &types.ProviderListResponse{
				Items: []*types.Provider{p},
			})
		},
	}

	f := c.Flags()
	f.StringVar(&WebhookURL, "url", "", "Webhook URL")
	f.StringVar(&APIToken, "token", "", "API Token")

	return c
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a provider",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cli, ok := client.ClientFromContext(ctx)
		if !ok {
			log.Printf("error retriving client")
			return
		}

		name := args[0]

		err := cli.DeleteProvider(ctx, name)
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}

		log.Printf("Provider removed: %s", name)
	},
}
