package api

import (
	"net/http"

	"github.com/Emyrk/chronicle/api/chroniclesdk"
	"github.com/Emyrk/chronicle/api/db2sdk"
	"github.com/Emyrk/chronicle/api/httpapi"
)

// ListPublicRealms returns all WoW server realms.
//
//	@Summary	List all realms
//	@Tags		Realms
//	@Produce	json
//	@Success	200	{array}	chroniclesdk.WoWServerRealm
//	@Router		/realms [get]
func (api *API) ListPublicRealms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	realms, err := api.Opts.Zed.ListAllWoWServerRealms(ctx)
	if err != nil {
		httpapi.HandleResponseError(ctx, w, err, httpapi.APIError{
			Response: chroniclesdk.Response{
				Message: "Failed to fetch realms",
				Detail:  err.Error(),
			},
		})
		return
	}

	resp := make([]chroniclesdk.WoWServerRealm, 0, len(realms))
	for _, r := range realms {
		resp = append(resp, db2sdk.WoWServerRealm(r))
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	httpapi.Write(ctx, w, http.StatusOK, resp)
}
