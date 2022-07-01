package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olxbr/network-api/pkg/types"
)

func (a *api) ListPools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pools, err := a.DB.ScanPools(ctx)

	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJson(w, types.PoolListResponse{
		Items: pools,
	}, http.StatusOK)
}

func (a *api) CreatePool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pr := &types.PoolRequest{}
	err := json.NewDecoder(r.Body).Decode(pr)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	err = validate.Struct(pr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	p := &types.Pool{
		ID:       types.NewUUID(),
		Name:     pr.Name,
		Region:   pr.Region,
		SubnetIP: pr.SubnetIP,
	}

	if pr.SubnetMask != nil {
		p.SubnetMask = pr.SubnetMask
	} else {
		p.SubnetMaxIP = pr.SubnetMaxIP
	}

	err = a.DB.PutPool(ctx, p)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusCreated)
}

func (a *api) DetailPool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	p, err := a.DB.GetPool(ctx, params["region"])

	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusOK)
}

func (a *api) DeletePool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	p, err := a.DB.GetPool(ctx, params["region"])
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	err = a.DB.DeletePool(ctx, p.Region)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusOK)
}
