package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/olxbr/network-api/pkg/db"
	"github.com/olxbr/network-api/pkg/secret"
	"github.com/olxbr/network-api/pkg/types"
)

type api struct {
	Router  *mux.Router
	Secrets secret.Secrets
	DB      db.Database
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func New(database db.Database, s secret.Secrets) *api {
	r := mux.NewRouter()
	a := &api{
		Router:  r,
		Secrets: s,
		DB:      database,
	}
	a.RegisterRoutes()
	return a
}

func (a *api) RegisterRoutes() {
	a.Router.HandleFunc("/", a.Version)
	api := a.Router.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/networks", a.ListNetworks).Methods(http.MethodGet)
	v1.HandleFunc("/networks", a.CreateNetwork).Methods(http.MethodPost)
	v1.HandleFunc("/networks/{id}", a.DetailNetwork).Methods(http.MethodGet)
	v1.HandleFunc("/networks/{id}", a.UpdateNetwork).Methods(http.MethodPut)
	v1.HandleFunc("/networks/{id}", a.DeleteNetwork).Methods(http.MethodDelete)
	v1.HandleFunc("/networks/{id}/subnets", a.GenerateSubnets).Methods(http.MethodGet)

	v1.HandleFunc("/pools", a.ListPools).Methods(http.MethodGet)
	v1.HandleFunc("/pools", a.CreatePool).Methods(http.MethodPost)
	v1.HandleFunc("/pools/{id}", a.DetailPool).Methods(http.MethodGet)
	v1.HandleFunc("/pools/{id}", a.DeletePool).Methods(http.MethodDelete)

	v1.HandleFunc("/providers", a.ListProviders).Methods(http.MethodGet)
	v1.HandleFunc("/providers", a.CreateProvider).Methods(http.MethodPost)
	v1.HandleFunc("/providers/{name}", a.DetailProvider).Methods(http.MethodGet)
	v1.HandleFunc("/providers/{name}", a.UpdateProvider).Methods(http.MethodPut)
	v1.HandleFunc("/providers/{name}", a.DeleteProvider).Methods(http.MethodDelete)
}

func (a *api) GetHandler() http.Handler {
	return a.Router
}

func writeJson(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("failed to write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	e := json.NewEncoder(w).Encode(types.NewSingleErrorResponse(err.Error()))
	if e != nil {
		log.Printf("failed to write json for error: %v", err)
	}
}
