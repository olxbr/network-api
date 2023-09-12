package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/olxbr/network-api/pkg/client"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPoolAddCommand(t *testing.T) {
	uuid := types.NewUUID()

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
				assert.Contains(t, e.Error(), `accepts 1 arg(s), received 0`)
			},
		},
		{
			name: "add pool without subnet mask or maxip",
			flags: []string{
				"pool-01",
				"--region", "us-east-1",
				"--subnet-ip", "10.2.0.0",
			},
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(&types.Pool{
					ID:       uuid,
					Name:     "pool-01",
					Region:   "us-east-1",
					SubnetIP: "10.2.0.0",
				})
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "A subnet mask or a maximum IP address are required")
			},
		},
		{
			name: "add pool",
			flags: []string{
				"pool-01",
				"--region", "us-east-1",
				"--subnet-ip", "10.2.0.0",
				"--subnet-mask", "16",
			},
			prepare: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(&types.Pool{
					ID:       uuid,
					Name:     "pool-01",
					Region:   "us-east-1",
					SubnetIP: "10.2.0.0",
				})
			},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, uuid.String())
				assert.Contains(t, out, "pool-01")
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
			cmd := poolAddCmd()
			var b bytes.Buffer
			cmd.SetOut(&b)
			log.SetOutput(&b)
			cmd.SetArgs(tt.flags)
			e := cmd.ExecuteContext(ctx)
			out, err := io.ReadAll(&b)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, string(out), e)
			log.SetOutput(os.Stderr)
			cmd.SetOut(os.Stdout)
		})
	}
}

func TestPoolListCommand(t *testing.T) {
	id01 := types.NewUUID()
	id02 := types.NewUUID()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		nr := &types.PoolListResponse{
			Items: []*types.Pool{
				{
					ID:       id01,
					Name:     "pool-01",
					Region:   "us-east-1",
					SubnetIP: "10.2.0.0",
				},
				{
					ID:       id02,
					Name:     "pool-02",
					Region:   "us-east-1",
					SubnetIP: "10.240.0.0",
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
	cmd := poolListCmd
	var b bytes.Buffer
	cmd.SetOut(&b)
	log.SetOutput(&b)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(&b)
	if err != nil {
		t.Fatal(err)
	}
	result := string(out)
	assert.Contains(t, result, id01.String())
	assert.Contains(t, result, "pool-01")
	assert.Contains(t, result, id02.String())
	assert.Contains(t, result, "pool-02")
	log.SetOutput(os.Stderr)
	cmd.SetOut(os.Stdout)
}
