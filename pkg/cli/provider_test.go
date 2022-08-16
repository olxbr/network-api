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

func TestProviderAddCommand(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		prepare func(w http.ResponseWriter, r *http.Request)
		assert  func(t *testing.T, out string, e error)
	}{
		{
			name:    "without name flag",
			flags:   []string{},
			prepare: func(w http.ResponseWriter, r *http.Request) {},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, e.Error(), `accepts 1 arg(s), received 0`)
			},
		},
		{
			name:    "without required flags",
			flags:   []string{"provider-01"},
			prepare: func(w http.ResponseWriter, r *http.Request) {},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, e.Error(), `required flag(s) "token", "url" not set`)
			},
		},
		{
			name: "add provider",
			flags: []string{
				"provider-01",
				"--url", "http://provider-01",
				"--token", "provider-01-token",
			},
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(&types.Provider{
					Name:       "provider-01",
					WebhookURL: "http://provider-01",
				})
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "provider-01")
				assert.Contains(t, out, "http://provider-01")
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
			cmd := providerAddCmd()
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

func TestProviderListCommand(t *testing.T) {
	id01 := types.NewUUID()
	id02 := types.NewUUID()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		nr := &types.ProviderListResponse{
			Items: []*types.Provider{
				{
					ID:         id01,
					Name:       "provider-01",
					WebhookURL: "http://provider-01",
				},
				{
					ID:         id02,
					Name:       "provider-02",
					WebhookURL: "http://provider-02",
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
	cmd := providerListCmd
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
	assert.Contains(t, result, "provider-01")
	assert.Contains(t, result, "http://provider-01")
	assert.Contains(t, result, "provider-02")
	assert.Contains(t, result, "http://provider-02")
	log.SetOutput(os.Stderr)
	cmd.SetOut(os.Stdout)
}

func TestProviderUpdateCommand(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		prepare func(w http.ResponseWriter, r *http.Request)
		assert  func(t *testing.T, out string, e error)
	}{
		{
			name: "update provider",
			flags: []string{
				"provider-01",
				"--url", "http://provider-01-new",
			},
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				_ = json.NewEncoder(w).Encode(&types.Provider{
					Name:       "provider-01",
					WebhookURL: "http://provider-01-new",
				})
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "Updated provider:")
				assert.Contains(t, out, "provider-01")
				assert.Contains(t, out, "http://provider-01-new")
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
			cmd := providerUpdateCmd()
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

func TestProviderRemoveCommand(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		prepare func(w http.ResponseWriter, r *http.Request)
		assert  func(t *testing.T, out string, e error)
	}{
		{
			name:  "delete provider",
			flags: []string{"provider-01"},
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "Provider removed: provider-01")
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
			cmd := providerRemoveCmd
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
