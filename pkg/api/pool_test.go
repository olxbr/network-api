package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	fakeDb "github.com/olxbr/network-api/pkg/db/fake"
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
