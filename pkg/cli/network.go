package cli

import (
	"fmt"
	"log"
	"os"

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
			ctx, err = SetupClientContext(WithConfig(ctx, cfg), cfg)
			if err != nil {
				return err
			}
			cmd.SetContext(ctx)
			return nil
		},
	}

	networkCmd.AddCommand(networkAddCmd())
	networkCmd.AddCommand(networkRemoveCmd)
	networkCmd.AddCommand(networkListCmd)
	networkCmd.AddCommand(networkInfoCmd())

	return networkCmd
}

func networkAddCmd() *cobra.Command {
	req := &types.NetworkRequest{}

	var AttachTGW bool
	var PrivateSubnet bool
	var PublicSubnet bool
	var Legacy bool
	var Reserved bool
	var CIDR string
	var SubnetSize int

	c := &cobra.Command{
		Use:   "add",
		Short: "Creates a new network",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return nil
			}

			if Reserved || Legacy {
				if CIDR == "" {
					return fmt.Errorf("missing CIDR with flags --reserved or --legacy")
				}
				req.CIDR = CIDR
			} else {
				req.SubnetSize = SubnetSize
			}

			req.AttachTGW = types.Bool(AttachTGW)
			req.PrivateSubnet = types.Bool(PrivateSubnet)
			req.PublicSubnet = types.Bool(PublicSubnet)

			n, err := cli.CreateNetwork(ctx, req)
			if err != nil {
				log.Printf("error creating network: %+v", err)
				return err
			}

			log.Printf("Network: %+v", n)
			return nil
		},
	}

	f := c.Flags()
	f.StringVar(&req.Provider, "provider", "", "Provider")
	f.StringVar(&req.Account, "account", "", "Account")
	f.StringVar(&req.Region, "region", "", "Region")
	f.StringVarP(&req.Environment, "environment", "e", "", "Environment")
	f.IntVar(&SubnetSize, "subnet-size", 0, "subnet")

	f.StringVar(&req.Info, "info", "", "Extra information about the VPC")

	f.BoolVar(&AttachTGW, "transit-gateway", true, "Attach transit gateway")
	f.BoolVar(&PrivateSubnet, "private", true, "Private subnet")
	f.BoolVar(&PublicSubnet, "public", true, "Public subnet")

	f.BoolVar(&Legacy, "legacy", false, "Legacy network - requires CIDR")
	f.BoolVar(&Reserved, "reserved", false, "Reserverd network - requires CIDR")
	f.StringVar(&CIDR, "cidr", "", "CIDR")

	return c
}

var networkRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a network",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func networkInfoCmd() *cobra.Command {
	var ID string
	// var VpcID string

	c := &cobra.Command{
		Use:   "info",
		Short: "Show network details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return nil
			}

			if ID == "" {
				return fmt.Errorf("missing Network ID")
			}

			n, err := cli.DetailNetwork(ctx, ID)
			if err != nil {
				log.Printf("Error: %s", err)
				return nil
			}

			fmt.Printf("Network: %s", n.ID.String())
			return nil
		},
	}

	f := c.Flags()
	f.StringVar(&ID, "network-id", "", "Network ID")
	// f.StringVar(&VpcID, "--vpc-id", "", "VPC ID")

	return c
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List networks",
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
