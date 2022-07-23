package cli

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olxbr/network-api/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestNewNetworkCommand(t *testing.T) {
	cmd := newNetworkCommand()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(out), "Network operations")
}

func TestNetworkAddCommand(t *testing.T) {
	tests := []struct {
		name   string
		flags  []string
		assert func(t *testing.T, out string, e error)
	}{
		{
			name:  "without flags",
			flags: []string{},
			assert: func(t *testing.T, out string, e error) {
				assert.Contains(t, out, "Usage")
				assert.Contains(t, out, "add [flags]")
			},
		},
		{
			name:  "legacy without cidr",
			flags: []string{"--legacy"},
			assert: func(t *testing.T, out string, e error) {
				MissingCIDR := errors.New("missing CIDR with flags --reserved or --legacy")
				assert.Equal(t, e, MissingCIDR)
			},
		},
		{
			name:  "reserved without cidr",
			flags: []string{"--reserved"},
			assert: func(t *testing.T, out string, e error) {
				MissingCIDR := errors.New("missing CIDR with flags --reserved or --legacy")
				assert.Equal(t, e, MissingCIDR)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			defer s.Close()
			ctx := context.Background()
			ctx = client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: s.URL,
				Client:   &http.Client{},
			})
			cmd := networkAddCmd()
			cmd.SetContext(ctx)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetArgs(tt.flags)
			e := cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, string(out), e)
		})
	}
}

func TestNetworkInfoCommand(t *testing.T) {
	tests := []struct {
		name   string
		flags  []string
		assert func(t *testing.T, out string, e error)
	}{
		{
			name:  "without network-id flag",
			flags: []string{},
			assert: func(t *testing.T, out string, e error) {
				MissingNetworkID := errors.New("missing Network ID")
				assert.Equal(t, e, MissingNetworkID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			defer s.Close()
			ctx := context.Background()
			ctx = client.WithNewClient(ctx, &client.ClientOptions{
				Endpoint: s.URL,
				Client:   &http.Client{},
			})
			cmd := networkInfoCmd()
			cmd.SetContext(ctx)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetArgs(tt.flags)
			e := cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, string(out), e)
		})
	}
}
