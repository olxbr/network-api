package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	fakeDb "github.com/olxbr/network-api/pkg/db/fake"
	fakeSecrets "github.com/olxbr/network-api/pkg/secret/fake"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCanListProviders(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "empty response",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanProviders", mock.Anything).Return([]*types.Provider{}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "{\"items\":[]}\n", w.Body.String())
			},
		},
		{
			name: "fail to read from database",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanProviders", mock.Anything).Return(nil, fmt.Errorf("error"))
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, "{\"errors\":{\"_all\":\"error\"}}\n", w.Body.String())
			},
		},
		{
			name: "providers",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanProviders", mock.Anything).Return([]*types.Provider{
					{
						ID:         types.NewUUID(),
						Name:       "aws",
						WebhookURL: "https://aws-napi.provider",
						APIToken:   "awsapitoken",
					},
					{
						ID:         types.NewUUID(),
						Name:       "gcp",
						WebhookURL: "https://gcp-napi.provider",
						APIToken:   "gcpapitoken",
					},
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.ProviderListResponse{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.Equal(t, 2, len(n.Items))
				assert.Equal(t, "aws", n.Items[0].Name)
				assert.Equal(t, "gcp", n.Items[1].Name)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.ListProviders(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanCreateProvider(t *testing.T) {
	tests := []struct {
		name    string
		payload interface{}
		prepare func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name:    "invalid payload",
			payload: "invalid",
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				assert.Equal(t, "{\"errors\":{\"_all\":\"json: cannot unmarshal string into Go value of type types.ProviderRequest\"}}\n", w.Body.String())
			},
		},
		{
			name: "missing payload data",
			payload: types.ProviderRequest{
				Name:       "",
				WebhookURL: "",
				APIToken:   "",
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'ProviderRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"))
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'ProviderRequest.WebhookURL' Error:Field validation for 'WebhookURL' failed on the 'required' tag"))
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'ProviderRequest.APIToken' Error:Field validation for 'APIToken' failed on the 'required' tag"))
			},
		},
		{
			name: "valid payload data",
			payload: types.ProviderRequest{
				Name:       "aws",
				WebhookURL: "https://aws-napi.provider",
				APIToken:   "awsapitoken",
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {
				s.On("PutAPIToken", mock.Anything, "aws", "awsapitoken").Return(nil)
				db.On("PutProvider", mock.Anything, mock.MatchedBy(func(n *types.Provider) bool {
					return (n.Name == "aws" &&
						n.WebhookURL == "https://aws-napi.provider" &&
						n.APIToken == "")
				})).Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusCreated, w.Code)
				n := &types.Provider{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.Equal(t, "aws", n.Name)
				assert.Equal(t, "https://aws-napi.provider", n.WebhookURL)
				assert.Equal(t, "", n.APIToken)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			s := &fakeSecrets.Secrets{}
			tt.prepare(t, db, s)

			payload := &bytes.Buffer{}
			err := json.NewEncoder(payload).Encode(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", payload)
			w := httptest.NewRecorder()
			api := New(db, s)

			api.CreateProvider(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanDetailProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		prepare      func(t *testing.T, db *fakeDb.Database)
		assert       func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name:         "valid provider name",
			providerName: "aws",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("GetProvider", mock.Anything, "aws").Return(&types.Provider{
					ID:         types.NewUUID(),
					Name:       "aws",
					WebhookURL: "https://aws-napi.provider",
					APIToken:   "",
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Provider{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "aws", n.Name)
				assert.Equal(t, "https://aws-napi.provider", n.WebhookURL)
				assert.Equal(t, "", n.APIToken)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"name": tt.providerName})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.DetailProvider(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanUpdateProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		payload      interface{}
		prepare      func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets)
		assert       func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name:         "valid update",
			providerName: "aws",
			payload: &types.ProviderUpdateRequest{
				WebhookURL: types.String("https://aws-net-api.com"),
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {
				db.On("GetProvider", mock.Anything, "aws").Return(&types.Provider{
					ID:         types.NewUUID(),
					Name:       "aws",
					WebhookURL: "https://aws-napi.provider",
					APIToken:   "",
				}, nil)
				s.On("PutAPIToken", mock.Anything, "aws", "awsapitoken").Return(nil)
				db.On("PutProvider", mock.Anything, mock.MatchedBy(func(n *types.Provider) bool {
					return (n.Name == "aws" &&
						n.WebhookURL == "https://aws-net-api.com" &&
						n.APIToken == "")
				})).Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Provider{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "aws", n.Name)
				assert.Equal(t, "https://aws-net-api.com", n.WebhookURL)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			s := &fakeSecrets.Secrets{}
			tt.prepare(t, db, s)

			payload := &bytes.Buffer{}
			err := json.NewEncoder(payload).Encode(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", payload)
			req = mux.SetURLVars(req, map[string]string{"name": tt.providerName})
			w := httptest.NewRecorder()
			api := New(db, s)

			api.UpdateProvider(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanDeleteProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		prepare      func(t *testing.T, db *fakeDb.Database)
		assert       func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name:         "valid delete",
			providerName: "aws",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("GetProvider", mock.Anything, "aws").Return(&types.Provider{
					ID:         types.NewUUID(),
					Name:       "aws",
					WebhookURL: "https://aws-napi.provider",
					APIToken:   "",
				}, nil)
				db.On("DeleteProvider", mock.Anything, "aws").Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Provider{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "aws", n.Name)
				assert.Equal(t, "https://aws-napi.provider", n.WebhookURL)
				assert.Equal(t, "", n.APIToken)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"name": tt.providerName})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.DeleteProvider(w, req)

			tt.assert(t, db, w)
		})
	}
}
