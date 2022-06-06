package provider

import (
	"context"
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
