package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Endpoint  string   `yaml:"endpoint,omitempty"`
	IssuerURL string   `yaml:"issuerURL,omitempty"`
	ClientID  string   `yaml:"clientID,omitempty"`
	Scopes    []string `yaml:"scopes,omitempty"`
}

var (
	configPathDefault = ".network-api"
	configFileDefault = configPathDefault + "/config"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	contextKeyConfig = contextKey("config")
)

func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKeyConfig, cfg)
}

func ConfigFromContext(ctx context.Context) (*Config, bool) {
	c, ok := ctx.Value(contextKeyConfig).(*Config)
	return c, ok
}

func LoadConfig() (*Config, error) {
	filename := filepath.Join(os.Getenv("HOME"), configFileDefault)

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", filename, err)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	var c Config
	if err := d.Decode(&c); err != nil {
		return nil, fmt.Errorf("invalid yaml file %s: %w", filename, err)
	}

	return &c, nil
}

func StatConfig(required bool) bool {
	filename := filepath.Join(os.Getenv("HOME"), configPathDefault)
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if required {
		return !required
	}
	if os.IsNotExist(err) {
		err := os.Mkdir(filename, 0750)
		if err != nil {
			log.Printf("error creating directory %s: %v", filename, err)
			return false
		}
		f, err := os.OpenFile(
			filepath.Join(os.Getenv("HOME"), configFileDefault),
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			0600,
		)
		if err != nil {
			log.Printf("error creating file %s: %v", filename, err)
			return false
		}
		defer f.Close()
		_, err = f.Write([]byte("\n"))
		if err != nil {
			log.Printf("error creating file %s: %v", filename, err)
			return false
		}
	}
	return true
}

func (c *Config) Save() error {
	filename := filepath.Join(os.Getenv("HOME"), configFileDefault)

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", filename, err)
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	if err := e.Encode(c); err != nil {
		return fmt.Errorf("error saving config %s: %w", filename, err)
	}
	return nil
}

func newConfigCommand() *cobra.Command {
	a := Config{}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure network-api cli",
		Run: func(cmd *cobra.Command, args []string) {
			if !StatConfig(false) {
				log.Printf("Something went wrong.")
				return
			}

			var err error
			var cfg *Config
			cfg, err = LoadConfig()
			if err != nil {
				cfg = &Config{}
			}

			if a.Endpoint != "" {
				cfg.Endpoint = a.Endpoint
			}

			if a.IssuerURL != "" {
				cfg.IssuerURL = a.IssuerURL
			}

			if a.ClientID != "" {
				cfg.ClientID = a.ClientID
			}

			if len(a.Scopes) > 0 {
				cfg.Scopes = a.Scopes
			}

			err = cfg.Save()
			if err != nil {
				log.Printf("Failed saving config: %+v", err)
			}
		},
	}

	f := configCmd.Flags()
	f.StringVar(&a.Endpoint, "endpoint", "", "Configure API Endpoint")
	f.StringVar(&a.IssuerURL, "oidc-issuer", "", "Configure OIDC Issuer URL")
	f.StringVar(&a.ClientID, "oidc-client-id", "", "Configure OIDC Client ID")
	f.StringSliceVar(&a.Scopes, "oidc-scopes", []string{}, "Configure OIDC Scopes")

	configCmd.AddCommand(configShowCmd)

	return configCmd
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Configure remote endpoint",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			return
		}
		fmt.Printf("Config: %+v\n", cfg)
	},
}
