package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"inet.af/netaddr"

	"github.com/olxbr/network-api/pkg/net"
	"github.com/olxbr/network-api/pkg/provider"
	"github.com/olxbr/network-api/pkg/types"
)

func (a *api) ListNetworks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nets, err := a.DB.ScanNetworks(ctx)

	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJson(w, types.NetworkListResponse{
		Items: nets,
	}, http.StatusOK)
}

func (a *api) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nr := &types.NetworkRequest{}
	err := json.NewDecoder(r.Body).Decode(nr)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	err = validate.Struct(nr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	pm := provider.New(a.DB, a.Secrets)
	nm := net.New(a.DB)
	pc, err := pm.GetClient(ctx, nr.Provider)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	p, err := a.DB.GetPool(ctx, nr.PoolID)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	n := &types.Network{
		ID:          types.NewUUID(),
		Account:     nr.Account,
		Region:      p.Region,
		Provider:    nr.Provider,
		Environment: nr.Environment,
		Info:        nr.Info,
	}

	if nr.AttachTGW != nil {
		n.AttachTGW = types.ToBool(nr.AttachTGW)
	}
	if nr.PrivateSubnet != nil {
		n.PrivateSubnet = types.ToBool(nr.PrivateSubnet)
	}
	if nr.PublicSubnet != nil {
		n.PublicSubnet = types.ToBool(nr.PublicSubnet)
	}

	if nr.Legacy != nil {
		n.Legacy = types.ToBool(nr.Legacy)
	}
	if nr.Reserved != nil {
		n.Reserved = types.ToBool(nr.Reserved)
	}

	if n.Reserved || n.Legacy {
		ipprefix, err := netaddr.ParseIPPrefix(nr.CIDR)
		if err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}

		err = nm.CheckNetwork(ctx, ipprefix)
		if err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		n.CIDR = ipprefix.String()
	} else {
		ipprefix, err := nm.AllocateNetwork(ctx, nr.PoolID, uint8(nr.SubnetSize))
		if err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		n.CIDR = ipprefix.String()
	}

	var wh *types.ProviderWebhookResponse
	if !n.Reserved || !n.Legacy {
		wh, err = pc.CreateNetwork(ctx, n)
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}
	}

	err = a.DB.PutNetwork(ctx, n)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	resp := &types.NetworkResponse{
		Network: n,
		Webhook: wh,
	}
	writeJson(w, resp, http.StatusCreated)
}

func (a *api) DetailNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	n, err := a.DB.GetNetwork(ctx, params["id"])

	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, n, http.StatusOK)
}

func (a *api) UpdateNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	n, err := a.DB.GetNetwork(ctx, params["id"])
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	nr := &types.NetworkUpdateRequest{}
	err = json.NewDecoder(r.Body).Decode(nr)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	if nr.VpcID != nil {
		n.VpcID = types.ToString(nr.VpcID)
	}

	if nr.Info != nil {
		n.Info = types.ToString(nr.Info)
	}

	err = a.DB.PutNetwork(ctx, n)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, n, http.StatusOK)
}

func (a *api) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	n, err := a.DB.GetNetwork(ctx, params["id"])
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	err = a.DB.DeleteNetwork(ctx, n.ID.String())
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, n, http.StatusOK)
}

func (a *api) GenerateSubnets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	n, err := a.DB.GetNetwork(ctx, params["id"])
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if n.Reserved || n.Legacy {
		writeError(w, errors.New("cannot generate subnets for reserved or legacy networks"), http.StatusBadRequest)
	}

	snets, err := net.GenerateSubnets(n)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, types.SubnetResponse{
		Subnets: snets,
	}, http.StatusOK)
}
