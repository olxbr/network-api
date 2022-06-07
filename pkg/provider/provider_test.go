package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	fakeDb "github.com/olxbr/network-api/pkg/db/fake"
	fakeSecrets "github.com/olxbr/network-api/pkg/secret/fake"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCanGetProviderClient(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		prepare  func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets)
		assert   func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets, pc *ProviderClient, err error)
	}{
		{
			name:     "valid provider",
			provider: "aws",
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {
				db.On("GetProvider", mock.Anything, "aws").Return(&types.Provider{
					Name:       "aws",
					WebhookURL: "https://example.com/webhook",
				}, nil)
				s.On("GetAPIToken", mock.Anything, "aws").Return("token", nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets, pc *ProviderClient, err error) {
				db.AssertExpectations(t)
				s.AssertExpectations(t)
				assert.NoError(t, err)
				assert.Equal(t, "https://example.com/webhook", pc.url)
				assert.Equal(t, "token", pc.auth)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			s := &fakeSecrets.Secrets{}
			pm := New(db, s)

			tt.prepare(t, db, s)

			c, err := pm.GetClient(context.Background(), tt.provider)
			tt.assert(t, db, s, c, err)
		})
	}
}

func TestProviderClientCanCreateNetwork(t *testing.T) {
	tests := []struct {
		name    string
		network *types.Network
		handler func(t *testing.T) http.Handler
		assert  func(t *testing.T, pwr *types.ProviderWebhookResponse, err error)
	}{
		{
			name: "valid network",
			network: &types.Network{
				ID:            types.NewUUID(),
				Account:       "123456789012",
				Region:        "us-east-1",
				Environment:   "prod",
				CIDR:          "10.10.0.0/20",
				PrivateSubnet: true,
				PublicSubnet:  true,
				AttachTGW:     true,
			},
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Authorization")

					pw := &types.ProviderWebhook{}
					err := json.NewDecoder(r.Body).Decode(pw)
					assert.NoError(t, err)

					assert.Equal(t, types.CreateNetwork, pw.Event)
					assert.Equal(t, "10.10.0.0/20", pw.CIDR)
					assert.Equal(t, "123456789012", pw.Account)
					assert.Equal(t, "us-east-1", pw.Region)
					assert.Equal(t, "prod", pw.Environment)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					err = json.NewEncoder(w).Encode(&types.ProviderWebhookResponse{
						StatusCode: http.StatusOK,
						ID:         "id",
					})
					assert.NoError(t, err)
				})
			},
			assert: func(t *testing.T, pwr *types.ProviderWebhookResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, pwr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler(t))
			defer server.Close()

			p := &ProviderClient{
				cli:  http.Client{},
				auth: "token",
				url:  server.URL,
			}

			resp, err := p.CreateNetwork(context.Background(), tt.network)

			tt.assert(t, resp, err)
		})
	}

}
