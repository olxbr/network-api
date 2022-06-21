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
			httpClient := auth.NewClient(ctx, t)

			cmd.SetContext(client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: cfg.Endpoint,
				Client:   httpClient,
			}))
			return nil
		},
	}

	networkCmd.AddCommand(networkAddCmd())
	networkCmd.AddCommand(networkRemoveCmd)
	networkCmd.AddCommand(networkListCmd)

	return networkCmd
}

func networkAddCmd() *cobra.Command {
	req := &types.NetworkRequest{}
	var AttachTGW bool
	var PrivateSubnet bool
	var PublicSubnet bool
	c := &cobra.Command{
		Use:   "add",
		Short: "Creates a new network",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return
			}

			req.AttachTGW = types.Bool(AttachTGW)
			req.PrivateSubnet = types.Bool(PrivateSubnet)
			req.PublicSubnet = types.Bool(PublicSubnet)

			n, err := cli.CreateNetwork(ctx, req)
			if err != nil {
				log.Printf("error creating network: %+v", err)
				return
			}

			log.Printf("Network: %+v", n)
		},
	}

	f := c.Flags()
	f.StringVar(&req.Provider, "provider", "", "Provider")
	f.StringVar(&req.Account, "account", "", "Account")
	f.StringVar(&req.Region, "region", "", "Region")
	f.StringVarP(&req.Environment, "environment", "e", "", "Environment")
	f.IntVar(&req.SubnetSize, "subnet", 0, "subnet")
	f.BoolVar(&AttachTGW, "transit-gateway", true, "Attach transit gateway")
	f.BoolVar(&PrivateSubnet, "private", true, "Private subnet")
	f.BoolVar(&PublicSubnet, "public", true, "Public subnet")

	return c
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
