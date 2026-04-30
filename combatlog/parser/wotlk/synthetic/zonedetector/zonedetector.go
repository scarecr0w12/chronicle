package zonedetector

import (
	"github.com/Emyrk/chronicle/combatlog/parser/types/zone"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/messages"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/state/encounters/registry"
)

// ZoneDetector infers the current zone from creature entry IDs seen in combat.
// WotLK logs lack ZONE_CHANGED events, so we match creature entries against
// known instance hostile lists and emit synthetic Zone messages.
type ZoneDetector struct {
	// entryToZone maps creature entry ID → zone name (lowercase).
	entryToZone map[uint32]string
	currentZone string
}

// New builds a ZoneDetector from a registry, indexing all hostile entries
// across all registered instances.
func New(reg *registry.Registry) *ZoneDetector {
	lookup := make(map[uint32]string)
	for _, entry := range reg.Entries() {
		for entryID := range entry.HostileEntries {
			for _, zn := range entry.ZoneNames {
				lookup[entryID] = zn
				break // one zone name per entry is sufficient
			}
		}
	}
	return &ZoneDetector{
		entryToZone: lookup,
	}
}

// ProcessMessages scans messages for creature GUIDs that belong to a known
// instance. When a new zone is detected, a synthetic Zone message is prepended.
func (zd *ZoneDetector) ProcessMessages(msgs []messages.Message) []messages.Message {
	if zd == nil {
    return msgs
  }
  for _, msg := range msgs {
		for _, g := range msg.Affects() {
			entry, ok := g.GetEntry()
			if !ok {
				continue
			}
			zoneName, ok := zd.entryToZone[entry]
			if !ok {
				continue
			}
			if zoneName == zd.currentZone {
				continue
			}
			zd.currentZone = zoneName
			synthetic := &messages.Zone{
				MessageBase: messages.Base(msg.Date()),
				Zone: zone.Zone{
					Seen:       msg.Date(),
					Name:       zoneName,
					IsInstance: true,
				},
			}
			return append([]messages.Message{synthetic}, msgs...)
		}
	}
	return msgs
}

// LastZone returns the most recently detected zone name, or "" if none.
func (zd *ZoneDetector) LastZone() string {
	return zd.currentZone
}
