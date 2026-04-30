package synthetic

import (
	"context"
	"log/slog"
	"time"

	"github.com/Emyrk/chronicle/combatlog/parser/guid"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/messages"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/state/encounters/registry"
	"github.com/Emyrk/chronicle/combatlog/parser/wotlk/synthetic/zonedetector"
	"github.com/Emyrk/chronicle/database/gamedb"
)

// NameResolver looks up a name for a GUID. Populated by the parser from
// combat log source/dest fields.
type NameResolver interface {
	Get(id guid.GUID) (string, bool)
}

// Synthetic processes the raw combat log events, and occasionally will insert
// or mutate synthetic events to help downstream consumers.
type Synthetic struct {
	logger *slog.Logger

	unitInfo     *unitInfo
	zoneDetector *zonedetector.ZoneDetector

	wowDB gamedb.GameDB

	unitInfoDur     time.Duration
	zoneDetectorDur time.Duration
}

func New(ctx context.Context, logger *slog.Logger, wowDB gamedb.GameDB, reg *registry.Registry, names NameResolver) *Synthetic {
	var zd *zonedetector.ZoneDetector
	if reg != nil {
		zonedetector.New(reg)
	}

	return &Synthetic{
		logger:       logger,
		wowDB:        wowDB,
		unitInfo:     newUnitInfo(ctx, logger, wowDB, names, wowDB),
		zoneDetector: zd,
	}
}

func (s *Synthetic) DetailedTimes() map[string]time.Duration {
	return map[string]time.Duration{
		"parser.synthetic.unit_info":     s.unitInfoDur,
		"parser.synthetic.zone_detector": s.zoneDetectorDur,
	}
}

func (s *Synthetic) ProcessMessages(msgs []messages.Message) ([]messages.Message, error) {
	now := time.Now()
	msgs = s.unitInfo.ProcessMessages(msgs)
	s.unitInfoDur += time.Since(now)

	if s.zoneDetector != nil {
		now = time.Now()
		msgs = s.zoneDetector.ProcessMessages(msgs)
		s.zoneDetectorDur += time.Since(now)
	}

	return msgs, nil
}
