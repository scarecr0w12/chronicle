package gamedataapi

import (
	"net/http"

	"github.com/Emyrk/chronicle/api/chroniclesdk"
	"github.com/Emyrk/chronicle/api/httpapi"
	"github.com/Emyrk/chronicle/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AssignWorldToServer assigns a world to a server.
func (h *Handler) AssignWorldToServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	worldID, err := uuid.Parse(chi.URLParam(r, "worldID"))
	if err != nil {
		httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
			Message: "Invalid world ID",
		})
		return
	}
	serverID, err := uuid.Parse(chi.URLParam(r, "serverID"))
	if err != nil {
		httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
			Message: "Invalid server ID",
		})
		return
	}

	store := database.New(h.pool)
	err = store.AssignWorldToServer(ctx, database.AssignWorldToServerParams{
		ServerID: serverID,
		WorldID:  worldID,
	})
	if err != nil {
		httpapi.InternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnassignWorldFromServer removes a world from a server.
func (h *Handler) UnassignWorldFromServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	worldID, err := uuid.Parse(chi.URLParam(r, "worldID"))
	if err != nil {
		httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
			Message: "Invalid world ID",
		})
		return
	}
	serverID, err := uuid.Parse(chi.URLParam(r, "serverID"))
	if err != nil {
		httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
			Message: "Invalid server ID",
		})
		return
	}

	store := database.New(h.pool)
	err = store.UnassignWorldFromServer(ctx, database.UnassignWorldFromServerParams{
		ServerID: serverID,
		WorldID:  worldID,
	})
	if err != nil {
		httpapi.InternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
