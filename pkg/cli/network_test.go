package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/olxbr/network-api/pkg/client"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNetworkAddCommand(t *testing.T) {
	uuid := types.NewUUID()
	params := []string{
		"--pool-id", "PoolID",
		"--account", "TestAccount",
		"--provider", "TestProvider",
		"--environment", "TestEnv",
		"--subnet-size", "16",
	}

	tests := []struct {
		name    string
		flags   []string
		prepare func(w http.ResponseWriter, r *http.Request)
		assert  func(t *testing.T, out string, e error)
	}{
		{
			name:    "without required flags",
			flags:   []string{},
			prepare: func(w http.ResponseWriter, r *http.Request) {},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, e.Error(), `"account", "environment", "pool-id", "provider" not set`)
			},
		},
		{
			name:    "legacy without cidr",
			flags:   append(params, "--legacy"),
			prepare: func(w http.ResponseWriter, r *http.Request) {},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "missing CIDR with flags --reserved or --legacy")
			},
		},
		{
			name:    "reserved without cidr",
			flags:   append(params, "--reserved"),
			prepare: func(w http.ResponseWriter, r *http.Request) {},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "missing CIDR with flags --reserved or --legacy")
			},
		},
		{
			name:  "add network",
			flags: params,
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(&types.NetworkResponse{
					Network: &types.Network{
						ID:          uuid,
						Provider:    "TestProvider",
						Region:      "us-east-1",
						Account:     "TestAccount",
						Environment: "TestEnv",
						CIDR:        "10.2.0.0",
					},
					Webhook: &types.ProviderWebhookResponse{},
				})
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, uuid.String())
				assert.Contains(t, out, "TestAccount")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(tt.prepare))
			defer s.Close()
			ctx := context.TODO()
			ctx = client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: s.URL,
				Client:   &http.Client{},
			})
			cmd := networkAddCmd()
			var b bytes.Buffer
			cmd.SetOut(&b)
			log.SetOutput(&b)
			cmd.SetArgs(tt.flags)
			e := cmd.ExecuteContext(ctx)
			out, err := ioutil.ReadAll(&b)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, string(out), e)
			log.SetOutput(os.Stderr)
			cmd.SetOut(os.Stdout)
		})
	}
}

func TestNetworkInfoCommand(t *testing.T) {
	uuid := types.NewUUID()

	tests := []struct {
		name    string
		flags   []string
		prepare func(w http.ResponseWriter, r *http.Request)
		assert  func(t *testing.T, out string, e error)
	}{
		{
			name:    "without network id",
			flags:   []string{},
			prepare: func(w http.ResponseWriter, r *http.Request) {},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, e.Error(), "accepts 1 arg(s), received 0")
			},
		},
		{
			name:  "valid network-id",
			flags: []string{uuid.String()},
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				_ = json.NewEncoder(w).Encode(&types.Network{
					ID:          uuid,
					Provider:    "TestProvider",
					Region:      "us-east-1",
					Account:     "TestAccount",
					Environment: "TestEnv",
					CIDR:        "10.2.0.0",
				})
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, uuid.String())
				assert.Contains(t, out, "TestAccount")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(tt.prepare))
			defer s.Close()
			ctx := context.TODO()
			ctx = client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: s.URL,
				Client:   &http.Client{},
			})
			cmd := networkInfoCmd
			var b bytes.Buffer
			cmd.SetOut(&b)
			log.SetOutput(&b)
			cmd.SetArgs(tt.flags)
			e := cmd.ExecuteContext(ctx)
			out, err := ioutil.ReadAll(&b)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, string(out), e)
			log.SetOutput(os.Stderr)
			cmd.SetOut(os.Stdout)
		})
	}
}

func TestNetworkListCommand(t *testing.T) {
	id01 := types.NewUUID()
	id02 := types.NewUUID()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		nr := &types.NetworkListResponse{
			Items: []*types.Network{
				{
					ID:          id01,
					Provider:    "TestProvider",
					Region:      "us-east-1",
					Account:     "TestAccount01",
					Environment: "TestEnv",
					CIDR:        "10.2.0.0",
				},
				{
					ID:          id02,
					Provider:    "TestProvider",
					Region:      "us-east-1",
					Account:     "TestAccount02",
					Environment: "TestEnv",
					CIDR:        "10.2.0.0",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(nr)
	}))
	defer s.Close()
	ctx := context.TODO()
	ctx = client.WithNewClient(ctx, &client.ClientOptions{
		Endpoint: s.URL,
		Client:   &http.Client{},
	})
	cmd := networkListCmd
	var b bytes.Buffer
	cmd.SetOut(&b)
	log.SetOutput(&b)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(&b)
	if err != nil {
		t.Fatal(err)
	}
	result := string(out)
	assert.Contains(t, result, id01.String())
	assert.Contains(t, result, "TestAccount01")
	assert.Contains(t, result, id02.String())
	assert.Contains(t, result, "TestAccount02")
	log.SetOutput(os.Stderr)
	cmd.SetOut(os.Stdout)
}
