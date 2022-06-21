package cli

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	//flags
	dryRun bool
)

type Runner struct {
	root *cobra.Command
}

func newRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:           "network-cli",
		Short:         "network-cli a command line interface for NetworkAPI",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !StatConfig(true) {
				log.Printf("Please configure using config option")
				return
			}

			cfg, err := LoadConfig()
			if err != nil {
				log.Printf("Is your config file correctly created?")
			}

			cmd.SetContext(WithConfig(cmd.Context(), cfg))
		},
	}
	f := c.PersistentFlags()
	f.BoolVar(&dryRun, "dry", false, "Dry run")
	return c
}

func NewRunner() *Runner {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newNetworkCommand())
	rootCmd.AddCommand(newProviderCommand())
	rootCmd.AddCommand(newConfigCommand())
	return &Runner{
		root: rootCmd,
	}
}

func (r *Runner) Run() {
	rand.Seed(time.Now().UnixNano())
	err := run(r.root)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Program aborted: %v\n", err)
		os.Exit(1)
	}
}

func run(command *cobra.Command) error {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	return command.ExecuteContext(ctx)
}
