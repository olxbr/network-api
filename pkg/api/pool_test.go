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

func TestCanListPools(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name: "empty response",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanPools", mock.Anything).Return([]*types.Pool{}, nil)
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
				db.On("ScanPools", mock.Anything).Return(nil, fmt.Errorf("error"))
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, "{\"errors\":{\"_all\":\"error\"}}\n", w.Body.String())
			},
		},
		{
			name: "pools",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("ScanPools", mock.Anything).Return([]*types.Pool{
					{
						ID:         types.NewUUID(),
						Name:       "pool-us",
						Region:     "us-east-1",
						SubnetIP:   "10.2.0.0",
						SubnetMask: types.Int(16),
					},
					{
						ID:          types.NewUUID(),
						Name:        "pool-sa",
						Region:      "sa-east-1",
						SubnetIP:    "10.0.0.0",
						SubnetMaxIP: types.String("10.2.255.255"),
					},
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.PoolListResponse{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.Equal(t, 2, len(n.Items))
				assert.Equal(t, "10.2.0.0", n.Items[0].SubnetIP)
				assert.Equal(t, "10.0.0.0", n.Items[1].SubnetIP)
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

			api.ListPools(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanCreatePool(t *testing.T) {
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
				assert.Equal(t, "{\"errors\":{\"_all\":\"json: cannot unmarshal string into Go value of type types.PoolRequest\"}}\n", w.Body.String())
			},
		},
		{
			name: "missing payload data",
			payload: types.PoolRequest{
				Name:     "",
				Region:   "",
				SubnetIP: "",
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'PoolRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"))
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'PoolRequest.Region' Error:Field validation for 'Region' failed on the 'required' tag"))
				assert.True(t, strings.Contains(w.Body.String(), "Key: 'PoolRequest.SubnetIP' Error:Field validation for 'SubnetIP' failed on the 'required' tag"))
			},
		},
		{
			name: "valid payload data",
			payload: types.PoolRequest{
				Name:       "pool-us",
				Region:     "us-east-1",
				SubnetIP:   "10.2.0.0",
				SubnetMask: types.Int(16),
			},
			prepare: func(t *testing.T, db *fakeDb.Database, s *fakeSecrets.Secrets) {
				db.On("PutPool", mock.Anything, mock.MatchedBy(func(n *types.Pool) bool {
					return (n.Name == "pool-us" &&
						n.Region == "us-east-1" &&
						n.SubnetIP == "10.2.0.0" &&
						types.ToInt(n.SubnetMask) == 16)
				})).Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusCreated, w.Code)
				n := &types.Pool{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.Equal(t, "pool-us", n.Name)
				assert.Equal(t, "us-east-1", n.Region)
				assert.Equal(t, "10.2.0.0", n.SubnetIP)
				assert.Equal(t, 16, types.ToInt(n.SubnetMask))
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

			api.CreatePool(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanDetailPool(t *testing.T) {
	tests := []struct {
		name    string
		region  string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name:   "valid pool region",
			region: "us-east-1",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("GetPool", mock.Anything, "us-east-1").Return(&types.Pool{
					ID:         types.NewUUID(),
					Name:       "pool-us",
					Region:     "us-east-1",
					SubnetIP:   "10.2.0.0",
					SubnetMask: types.Int(16),
				}, nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Pool{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "pool-us", n.Name)
				assert.Equal(t, "us-east-1", n.Region)
				assert.Equal(t, "10.2.0.0", n.SubnetIP)
				assert.Equal(t, 16, types.ToInt(n.SubnetMask))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"region": tt.region})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.DetailPool(w, req)

			tt.assert(t, db, w)
		})
	}
}

func TestCanDeletePool(t *testing.T) {
	tests := []struct {
		name    string
		region  string
		prepare func(t *testing.T, db *fakeDb.Database)
		assert  func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder)
	}{
		{
			name:   "valid delete",
			region: "us-east-1",
			prepare: func(t *testing.T, db *fakeDb.Database) {
				db.On("GetPool", mock.Anything, "us-east-1").Return(&types.Pool{
					ID:         types.NewUUID(),
					Name:       "pool-us",
					Region:     "us-east-1",
					SubnetIP:   "10.2.0.0",
					SubnetMask: types.Int(16),
				}, nil)
				db.On("DeletePool", mock.Anything, "us-east-1").Return(nil)
			},
			assert: func(t *testing.T, db *fakeDb.Database, w *httptest.ResponseRecorder) {
				db.AssertExpectations(t)
				assert.Equal(t, http.StatusOK, w.Code)
				n := &types.Pool{}
				err := json.NewDecoder(w.Body).Decode(n)
				require.NoError(t, err)
				assert.NotNil(t, n.ID)
				assert.Equal(t, "pool-us", n.Name)
				assert.Equal(t, "us-east-1", n.Region)
				assert.Equal(t, "10.2.0.0", n.SubnetIP)
				assert.Equal(t, 16, types.ToInt(n.SubnetMask))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDb.Database{}
			tt.prepare(t, db)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, map[string]string{"region": tt.region})
			w := httptest.NewRecorder()
			api := New(db, nil)

			api.DeletePool(w, req)

			tt.assert(t, db, w)
		})
	}
}
