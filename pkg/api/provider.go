package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olxbr/network-api/pkg/types"
)

func (a *api) ListProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providers, err := a.DB.ScanProviders(ctx)

	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJson(w, types.ProviderResponse{
		Items: providers,
	}, http.StatusOK)
}

func (a *api) CreateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pr := &types.ProviderRequest{}
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

	p := &types.Provider{
		ID:         types.NewUUID(),
		Name:       pr.Name,
		WebhookURL: pr.WebhookURL,
	}

	err = a.Secrets.PutAPIToken(ctx, p.Name, pr.APIToken)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	err = a.DB.PutProvider(ctx, p)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusCreated)
}

func (a *api) DetailProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	p, err := a.DB.GetProvider(ctx, params["region"])

	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusOK)
}

func (a *api) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	p, err := a.DB.GetProvider(ctx, params["name"])
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	pr := &types.ProviderUpdateRequest{}
	err = json.NewDecoder(r.Body).Decode(pr)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	if pr.WebhookURL != nil {
		p.WebhookURL = types.ToString(pr.WebhookURL)
	}

	if pr.APIToken != nil {
		err = a.Secrets.PutAPIToken(ctx, p.Name, types.ToString(pr.APIToken))
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}
	}

	err = a.DB.PutProvider(ctx, p)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusOK)
}

func (a *api) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)

	p, err := a.DB.GetProvider(ctx, params["name"])
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}

	err = a.DB.DeleteProvider(ctx, p.Name)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, p, http.StatusOK)
}
