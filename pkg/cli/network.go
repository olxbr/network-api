package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/olxbr/network-api/pkg/client"
	"github.com/spf13/cobra"
)

func newNetworkCommand() *cobra.Command {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Network operations",
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
			fmt.Printf("Token: %s\n", t.AccessToken)
			fmt.Printf("RefreshToken: %s\n", t.RefreshToken)
			httpClient := auth.NewClient(ctx, t)

			cmd.SetContext(client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: cfg.Endpoint,
				Client:   httpClient,
			}))
			return nil
		},
	}

	networkCmd.AddCommand(networkAddCmd)
	networkCmd.AddCommand(networkRemoveCmd)
	networkCmd.AddCommand(networkListCmd)

	return networkCmd
}

var networkAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Creates a new network",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var networkRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Configure remote endpoint",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "Configure remote endpoint",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		fmt.Printf("%+v\n", ctx)
		cli, ok := client.ClientFromContext(ctx)
		if !ok {
			log.Printf("error retriving client")
			return
		}
		ns, err := cli.ListNetworks(ctx)
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Provider", "Account", "Region", "Environment", "CIDR", "VpcID", "Info"})

		for _, n := range ns {
			table.Append([]string{
				n.ID.String(),
				n.Provider,
				n.Account,
				n.Region,
				n.Environment,
				n.CIDR,
				n.VpcID,
				n.Info,
			})
		}

		table.Render()
	},
}
