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

func newPoolCommand() *cobra.Command {
	poolCmd := &cobra.Command{
		Use:   "pool",
		Short: "Pool operations",
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

	poolCmd.AddCommand(poolAddCmd())
	poolCmd.AddCommand(poolRemoveCmd)
	poolCmd.AddCommand(poolListCmd)

	return poolCmd
}

func poolAddCmd() *cobra.Command {
	req := &types.PoolRequest{}
	var subnetMask int
	var subnetMaxIP string
	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Adds a new pool",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			cli, ok := client.ClientFromContext(ctx)
			if !ok {
				log.Printf("error retriving client")
				return
			}

			req.Name = args[0]

			if subnetMask != -1 {
				req.SubnetMask = types.Int(subnetMask)
			} else if subnetMaxIP != "" {
				req.SubnetMaxIP = types.String(subnetMaxIP)
			} else {
				log.Printf("A subnet mask or a maximum IP address are required")
				return
			}

			p, err := cli.CreatePool(ctx, req)
			if err != nil {
				log.Printf("error creating pool: %+v", err)
				return
			}

			log.Printf("Pool: %+v", p)
		},
	}

	f := c.Flags()
	f.StringVar(&req.Region, "region", "", "Region")
	f.StringVar(&req.SubnetIP, "subnet-ip", "", "Subnet IP Address")
	f.IntVar(&subnetMask, "subnet-mask", -1, "Subnet Mask")
	f.StringVar(&subnetMaxIP, "subnet-maxip", "", "Subnet Maximum IP Address")

	c.MarkFlagsMutuallyExclusive("subnet-mask", "subnet-maxip")
	c.MarkFlagRequired("region")
	c.MarkFlagRequired("subnet-ip")

	return c
}

var poolRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a pool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available IP pools",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cli, ok := client.ClientFromContext(ctx)
		if !ok {
			log.Printf("error retriving client")
			return
		}
		ps, err := cli.ListPools(ctx)
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Region", "Range"})

		for _, p := range ps.Items {
			var r string
			if p.SubnetMask != nil {
				r = fmt.Sprintf("%s/%d", p.SubnetIP, types.ToInt(p.SubnetMask))
			} else if p.SubnetMaxIP != nil {
				r = fmt.Sprintf("%s - %s", p.SubnetIP, types.ToString(p.SubnetMaxIP))
			}
			table.Append([]string{
				p.Name,
				p.Region,
				r,
			})
		}

		table.Render()
	},
}
