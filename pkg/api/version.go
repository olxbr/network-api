package api

import (
	"net/http"

	"github.com/olxbr/network-api/cmd"
	"github.com/olxbr/network-api/pkg/types"
)

func (a *api) Version(w http.ResponseWriter, r *http.Request) {
	writeJson(w, types.Version{
		Name:      "Network API",
		Version:   cmd.Version,
		CommitID:  cmd.CommitID,
		BuildTime: cmd.BuildTime,
	}, http.StatusOK)
}
