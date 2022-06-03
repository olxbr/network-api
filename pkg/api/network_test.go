package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fakeDb "github.com/olxbr/network-api/pkg/db/fake"
	fakeSecrets "github.com/olxbr/network-api/pkg/secret/fake"
	"github.com/olxbr/network-api/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCanListNetworks(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "empty response",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{}, nil)
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
				db.On("ScanNetworks", mock.Anything).Return(nil, fmt.Errorf("error"))
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, "{\"errors\":{\"_all\":\"error\"}}\n", w.Body.String())
			},
		},
		{
			name: "networks",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{
					{
						ID:          types.NewUUID(),
						CIDR:        "10.0.0.0/16",
						Region:      "us-east-1",
						Provider:    "aws",
						Account:     "123456789012",
						Environment: "prod",
						Info:        "First VPC",
					},
					{
						ID:          types.NewUUID(),
						CIDR:        "10.1.0.0/16",
						Region:      "us-east-1",
						Provider:    "aws",
						Account:     "123456789012",
						Environment: "qa",
						Info:        "Second VPC",
					},
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.NetworkListResponse{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.Equal(t, 2, len(n.Items))
				assert.Equal(t, "10.0.0.0/16", n.Items[0].CIDR)
				assert.Equal(t, "10.1.0.0/16", n.Items[1].CIDR)
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

			api.ListNetworks(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanCreateNetwork(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"id\":\"123456789012\",\"statusCode\":201}"))
	}))

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
				assert.Equal(t, "{\"errors\":{\"_all\":\"json: cannot unmarshal string into Go value of type types.NetworkRequest\"}}\n", w.Body.String())
			},
		},
		{
			name: "missing payload data",
			payload: types.NetworkRequest{
				Account: "",
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'NetworkRequest.Account' Error:Field validation for 'Account' failed on the 'required' tag"))
			},
		},
		{
			name: "valid payload data",
			payload: types.NetworkRequest{
				Account:       "1234",
				Region:        "us-east-1",
				Provider:      "aws",
				Environment:   "prod",
				Info:          "First VPC",
				SubnetSize:    20,
				AttachTGW:     types.Bool(true),
				PrivateSubnet: types.Bool(true),
				PublicSubnet:  types.Bool(true),
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {
				db.On("GetProvider", mock.Anything, "aws").Return(&types.Provider{
					WebhookURL: server.URL,
				}, nil)
				s.On("GetAPIToken", mock.Anything, "aws").Return("token", nil)
				db.On("GetPool", mock.Anything, "us-east-1").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{}, nil)
				db.On("PutNetwork", mock.Anything, mock.MatchedBy(func(n *types.Network) bool {
					return (n.Account == "1234" &&
						n.Region == "us-east-1" &&
						n.Provider == "aws" &&
						n.Environment == "prod" &&
						n.Info == "First VPC" &&
						n.CIDR == "10.0.0.0/20" &&
						n.AttachTGW == true &&
						n.PrivateSubnet == true &&
						n.PublicSubnet == true)
				})).Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusCreated, w.Code)
				n := &types.NetworkResponse{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.Equal(t, "1234", n.Network.Account)
				assert.Equal(t, "us-east-1", n.Network.Region)
				assert.Equal(t, "aws", n.Network.Provider)
				assert.Equal(t, "prod", n.Network.Environment)
				assert.Equal(t, "First VPC", n.Network.Info)
				assert.Equal(t, "10.0.0.0/20", n.Network.CIDR)
				assert.Equal(t, "123456789012", n.Webhook.ID)
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

			api.CreateNetwork(w, req)

			tt.assert(t, db, w)
		})
	}
}
