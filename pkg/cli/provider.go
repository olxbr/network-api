package cli

import (
	"log"
	"os"

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
			ctx, err = SetupClientContext(WithConfig(ctx, cfg), cfg)
			if err != nil {
				return err
			}
			cmd.SetContext(ctx)
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
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return
			}

			req.Name = args[0]

			n, err := cli.CreateProvider(ctx, req)
			if err != nil {
				log.Printf("error creating network: %+v", err)
				return
			}

			log.Printf("Network: %+v", n)
		},
	}

	f := c.Flags()
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