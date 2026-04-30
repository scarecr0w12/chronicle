package registry

import (
	"context"
	"log/slog"
	"sync"

	"github.com/Emyrk/chronicle/api/db2sdk"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/state/encounters/instances"
	"github.com/Emyrk/chronicle/database"
	"github.com/Emyrk/chronicle/database/authz"
	"github.com/Emyrk/chronicle/database/pubsub"
	"github.com/google/uuid"
)

const InstanceRegistryChannel = "instance_registry_changed"

// DBRegistry is a Registry that can reload its entries from the database.
// It wraps a *Registry with the ability to atomically swap it on reload.
// Parse jobs call Registry() to get an immutable snapshot.
//
// When a pubsub.Pubsub is provided, Reload broadcasts a notification so
// other server replicas also reload. Each DBRegistry subscribes to the
// channel on creation and reloads automatically when notified.
type DBRegistry struct {
	mu       sync.RWMutex
	registry *Registry
	logger   *slog.Logger
	store    *authz.Authz
	ps       pubsub.Pubsub
	cancel   func() // unsubscribe from pubsub
	fallback *Registry
}

// NewDBRegistry creates a new DBRegistry and performs an initial load from the
// database. If ps is non-nil, subscribes to reload notifications from other
// server replicas.
func NewDBRegistry(
	ctx context.Context,
	logger *slog.Logger,
	store *authz.Authz,
	ps pubsub.Pubsub,
	fallback *Registry,
) (*DBRegistry, error) {
	dr := &DBRegistry{
		logger:   logger,
		store:    store,
		ps:       ps,
		fallback: fallback,
	}
	if err := dr.reload(ctx); err != nil {
		return nil, err
	}

	if ps != nil {
		cancel, err := ps.Subscribe(InstanceRegistryChannel, func(ctx context.Context, _ []byte) {
			if reloadErr := dr.reload(ctx); reloadErr != nil {
				logger.Error("failed to reload instance registry via pubsub", "error", reloadErr)
			}
		})
		if err != nil {
			return nil, err
		}
		dr.cancel = cancel
	}

	return dr, nil
}

// Close unsubscribes from pubsub notifications.
func (dr *DBRegistry) Close() {
	if dr.cancel != nil {
		dr.cancel()
	}
}

// Reload fetches all instance data from the DB, rebuilds the registry, and
// notifies other server replicas via pubsub (if configured).
func (dr *DBRegistry) Reload(ctx context.Context) error {
	if err := dr.reload(ctx); err != nil {
		return err
	}
	if dr.ps != nil {
		if err := dr.ps.Publish(InstanceRegistryChannel, []byte("reload")); err != nil {
			dr.logger.Error("failed to publish instance registry reload", "error", err)
		}
	}
	return nil
}

// reload fetches all instance data from the DB and rebuilds the registry.
// Does NOT publish — used internally and by the pubsub listener to avoid loops.
func (dr *DBRegistry) reload(ctx context.Context) error {
	r := NewRegistry(dr.logger)

	templates, err := dr.store.ListWorldInstanceTemplates(ctx)
	if err != nil {
		return err
	}

	allZoneNames, err := dr.store.ListWorldInstanceZoneNames(ctx)
	if err != nil {
		return err
	}

	allUnits, err := dr.store.ListWorldInstanceUnits(ctx)
	if err != nil {
		return err
	}

	// Group zone names and units by instance ID.
	zoneNamesByInstance := make(map[uuid.UUID][]database.WorldInstanceZoneName)
	for _, zn := range allZoneNames {
		zoneNamesByInstance[zn.InstanceID] = append(zoneNamesByInstance[zn.InstanceID], zn)
	}

	unitsByInstance := make(map[uuid.UUID][]database.ListWorldInstanceUnitsRow)
	for _, u := range allUnits {
		unitsByInstance[u.InstanceID] = append(unitsByInstance[u.InstanceID], u)
	}

	for _, tmpl := range templates {
		zoneNames := zoneNamesByInstance[tmpl.ID]
		units := unitsByInstance[tmpl.ID]

		// Build Identity map from DB units.
		hostiles := make(map[uint32]instances.Identity)
		for _, u := range units {
			id := instances.Identity{
				Affiliation: db2sdk.Affiliation(u.Affiliation),
			}
			if u.Boss {
				id.Boss = true
				if u.EncounterName.Valid {
					id.EncounterName = u.EncounterName.String
				}
			}
			hostiles[uint32(u.EntryID)] = id
		}

		// Build zone name list.
		names := make([]string, 0, len(zoneNames))
		for _, zn := range zoneNames {
			names = append(names, zn.ZoneName)
		}

		factory := &instances.CommonFactory{
			Name:      tmpl.Name,
			ZoneNames: names,
			Hostiles:  instances.FromMap(hostiles),
		}

		r.RegisterEntry(FromCommonFactory(factory))
	}

	if dr.fallback != nil {
		r.SetFallback(dr.fallback)
	}

	dr.logger.Info("reloaded instance registry from database",
		slog.Int("instances", len(templates)),
	)

	dr.mu.Lock()
	dr.registry = r
	dr.mu.Unlock()
	return nil
}

// Registry returns an immutable snapshot of the current registry.
// Safe for concurrent use by parse jobs.
func (dr *DBRegistry) Registry() *Registry {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	return dr.registry
}
