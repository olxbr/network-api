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
				PoolID:        "poolid",
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
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
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
		{
			name: "valid legacy network data",
			payload: types.NetworkRequest{
				Account:       "1234",
				PoolID:        "poolid",
				Provider:      "aws",
				Environment:   "prod",
				Info:          "First VPC",
				CIDR:          "10.10.0.0/16",
				AttachTGW:     types.Bool(true),
				PrivateSubnet: types.Bool(true),
				PublicSubnet:  types.Bool(true),
				Legacy:        types.Bool(true),
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {
				db.On("GetProvider", mock.Anything, "aws").Return(&types.Provider{
					WebhookURL: server.URL,
				}, nil)
				db.On("GetPool", mock.Anything, "poolid").Return(&types.Pool{
					Region:     "us-east-1",
					SubnetIP:   "10.0.0.0",
					SubnetMask: types.Int(8),
				}, nil)
				s.On("GetAPIToken", mock.Anything, "aws").Return("token", nil)
				db.On("ScanNetworks", mock.Anything).Return([]*types.Network{}, nil)
				db.On("PutNetwork", mock.Anything, mock.MatchedBy(func(n *types.Network) bool {
					return (n.Account == "1234" &&
						n.Region == "us-east-1" &&
						n.Provider == "aws" &&
						n.Environment == "prod" &&
						n.Info == "First VPC" &&
						n.CIDR == "10.10.0.0/16" &&
						n.AttachTGW == true &&
						n.PrivateSubnet == true &&
						n.PublicSubnet == true &&
						n.Legacy == true)
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
				assert.Equal(t, "10.10.0.0/16", n.Network.CIDR)
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

func TestCanDetailNetwork(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "valid network id",
			id:   "1234",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("GetNetwork", mock.Anything, "1234").Return(&types.Network{
					ID:          types.NewUUID(),
					Provider:    "aws",
					Region:      "us-east-1",
					Account:     "1234",
					Environment: "prod",
					CIDR:        "10.10.0.0/16",
					VpcID:       "vpc-123456789012",
					Info:        "Legacy VPC",
					Legacy:      true,
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Network{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "1234", n.Account)
				assert.Equal(t, "us-east-1", n.Region)
				assert.Equal(t, "aws", n.Provider)
				assert.Equal(t, "prod", n.Environment)
				assert.Equal(t, "Legacy VPC", n.Info)
				assert.Equal(t, "10.10.0.0/16", n.CIDR)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.DetailNetwork(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanUpdateNetwork(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		payload interface{}
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "valid update",
			id:   "1234",
			payload: &types.NetworkUpdateRequest{
				Info: types.String("Old Legacy VPC"),
			},
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("GetNetwork", mock.Anything, "1234").Return(&types.Network{
					ID:          types.NewUUID(),
					Provider:    "aws",
					Region:      "us-east-1",
					Account:     "1234",
					Environment: "prod",
					CIDR:        "10.10.0.0/16",
					VpcID:       "vpc-123456789012",
					Info:        "Legacy VPC",
					Legacy:      true,
				}, nil)
				db.On("PutNetwork", mock.Anything, mock.MatchedBy(func(n *types.Network) bool {
					return (n.Account == "1234" &&
						n.Region == "us-east-1" &&
						n.Provider == "aws" &&
						n.Environment == "prod" &&
						n.Info == "Old Legacy VPC" &&
						n.CIDR == "10.10.0.0/16" &&
						n.Legacy == true)
				})).Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Network{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "1234", n.Account)
				assert.Equal(t, "us-east-1", n.Region)
				assert.Equal(t, "aws", n.Provider)
				assert.Equal(t, "prod", n.Environment)
				assert.Equal(t, "Old Legacy VPC", n.Info)
				assert.Equal(t, "10.10.0.0/16", n.CIDR)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			payload := &bytes.Buffer{}
			err := json.NewEncoder(payload).Encode(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", payload)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.UpdateNetwork(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanDeleteNetwork(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "valid delete",
			id:   "1234",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				uuid := types.NewUUID()
				db.On("GetNetwork", mock.Anything, "1234").Return(&types.Network{
					ID:          uuid,
					Provider:    "aws",
					Region:      "us-east-1",
					Account:     "1234",
					Environment: "prod",
					CIDR:        "10.10.0.0/16",
					VpcID:       "vpc-123456789012",
					Info:        "Legacy VPC",
					Legacy:      true,
				}, nil)
				db.On("DeleteNetwork", mock.Anything, uuid.String()).Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Network{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "1234", n.Account)
				assert.Equal(t, "us-east-1", n.Region)
				assert.Equal(t, "aws", n.Provider)
				assert.Equal(t, "prod", n.Environment)
				assert.Equal(t, "Legacy VPC", n.Info)
				assert.Equal(t, "10.10.0.0/16", n.CIDR)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.DeleteNetwork(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanGenerateSubnets(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "legacy vpc",
			id:   "1234",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				uuid := types.NewUUID()
				db.On("GetNetwork", mock.Anything, "1234").Return(&types.Network{
					ID:          uuid,
					Provider:    "aws",
					Region:      "us-east-1",
					Account:     "1234",
					Environment: "prod",
					CIDR:        "10.10.0.0/16",
					VpcID:       "vpc-123456789012",
					Info:        "Legacy VPC",
					Legacy:      true,
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				e := &types.ErrorResponse{}
				err := json.NewDecoder(w.Body).Decode(e)
				require.NoError(t, err)
				assert.Len(t, e.Errors, 1)
				assert.Equal(t, "cannot generate subnets for reserved or legacy networks", e.Errors["_all"])
			},
		},
		{
			name: "valid network",
			id:   "1234",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				uuid := types.NewUUID()
				db.On("GetNetwork", mock.Anything, "1234").Return(&types.Network{
					ID:            uuid,
					Provider:      "aws",
					Region:        "us-east-1",
					Account:       "1234",
					Environment:   "prod",
					CIDR:          "10.1.0.0/16",
					VpcID:         "vpc-123456789012",
					Info:          "New VPC",
					PrivateSubnet: true,
					PublicSubnet:  true,
					AttachTGW:     true,
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				e := &types.SubnetResponse{}
				err := json.NewDecoder(w.Body).Decode(e)
				require.NoError(t, err)
				assert.Len(t, e.Subnets, 9)
				assert.Equal(t, "10.1.0.0/19", e.Subnets[0].CIDR)
				assert.Equal(t, "10.1.64.0/19", e.Subnets[1].CIDR)
				assert.Equal(t, "10.1.128.0/19", e.Subnets[2].CIDR)
				assert.Equal(t, types.Private, e.Subnets[2].Type)

				assert.Equal(t, "10.1.32.0/20", e.Subnets[3].CIDR)
				assert.Equal(t, "10.1.96.0/20", e.Subnets[4].CIDR)
				assert.Equal(t, "10.1.160.0/20", e.Subnets[5].CIDR)
				assert.Equal(t, types.Public, e.Subnets[5].Type)

				assert.Equal(t, "10.1.48.0/28", e.Subnets[6].CIDR)
				assert.Equal(t, "10.1.112.0/28", e.Subnets[7].CIDR)
				assert.Equal(t, "10.1.176.0/28", e.Subnets[8].CIDR)
				assert.Equal(t, types.TransitGateway, e.Subnets[8].Type)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.GenerateSubnets(w, req)

			tt.assert(t, db, w)
		})
	}
}
