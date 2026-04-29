package gamedataapi

import (
	"net/http"

	"github.com/Emyrk/chronicle/api/chronauth"
	"github.com/Emyrk/chronicle/api/httpmw"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/state/encounters/registry"
	"github.com/Emyrk/chronicle/database/authz"
	"github.com/Emyrk/chronicle/database/authz/policy"
	"github.com/Emyrk/chronicle/database/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	zed  *authz.Authz
	auth *chronauth.Service
	pool *pgxpool.Pool
	ps   pubsub.Pubsub
}

func New(zed *authz.Authz, auth *chronauth.Service, pool *pgxpool.Pool, ps pubsub.Pubsub) *Handler {
	return &Handler{zed: zed, auth: auth, pool: pool, ps: ps}
}

func (h *Handler) notifyInstanceDataChanged() {
	_ = h.ReloadWorldData()
}

func (h *Handler) ReloadWorldData() error {
	if err := h.ps.Publish(registry.InstanceRegistryChannel, []byte("reload")); err != nil {
		return err
	}
	return nil
}

// PublicRoutes returns routes that don't require authentication.
func (h *Handler) PublicRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/instances", h.PublicListInstances)
	return r
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(
		h.auth.Authenticated(false),
		httpmw.Can(h.zed, policy.New().GlobalChronicle().CanAdmin_world_data_User),
	)
	r.Post("/wdb/upload", h.UploadWDB)
	r.Post("/sql/import", h.ImportSQL)
	r.Post("/sql/import-url", h.ImportSQLFromURL)
	r.Post("/dbc/upload", h.UploadDBC)

	// Instance template CRUD
	r.Get("/instances", h.ListInstances)
	r.Post("/instances", h.UpsertInstance)
	r.Post("/instances/import", h.ImportInstanceDump)
	r.Delete("/instances/{instanceID}", h.DeleteInstance)
	r.Get("/instances/{instanceID}/units", h.ListInstanceUnits)
	r.Post("/instances/{instanceID}/units", h.BulkUpsertInstanceUnits)

	// World <-> Server assignment
	r.Post("/worlds/{worldID}/servers/{serverID}", h.AssignWorldToServer)
	r.Delete("/worlds/{worldID}/servers/{serverID}", h.UnassignWorldFromServer)

	return r
}
