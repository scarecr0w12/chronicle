package instances

import (
	"github.com/Emyrk/chronicle/combatlog/parser/guid"
)

type EncounterFuncResult struct {
	// EncounterName, if set, will be used to identify a named encounter.
	EncounterName string
	// Boss indicate if this fight has some bosses that need to be manually included.
	// This is important for fights where the boss comes into the fight after some
	// period of time.
	Bosses []uint32
}

type Identity struct {
	Hostile bool
	// Name is the display name for this unit (e.g. "Lava Surger", "Ragnaros").
	Name string
	// EncounterName, if set, will be used to identify a named encounter.
	EncounterName string
	// Boss indicates if the unit is considered a boss for encounter purposes.
	Boss bool

	EncounterNameFn func(f Fight) *EncounterFuncResult
}

type Identifier struct {
	byEntryId    map[uint32]Identity
	unknownUnits map[uint32]int // creature entry IDs not in hostiles map, with hit count
}

func NewIdentifier(byEntryId map[uint32]Identity) *Identifier {
	return &Identifier{
		byEntryId:    byEntryId,
		unknownUnits: make(map[uint32]int),
	}
}

func (i *Identifier) AddEntryId(entryId uint32, identity Identity) {
	i.byEntryId[entryId] = identity
}

func (i *Identifier) IdentifyUnit(id guid.GUID) Identity {
	if id.IsPlayer() {
		return Identity{Hostile: false}
	}

	entryID, ok := id.GetEntry()
	if !ok {
		return Identity{Hostile: false}
	}

	identity, exists := i.byEntryId[entryID]
	if !exists {
		i.unknownUnits[entryID]++
		return Identity{Hostile: false}
	}
	return identity
}

// UnknownUnits returns creature entry IDs that were looked up but not found in the
// hostiles map, with the number of times each was seen.
func (i *Identifier) UnknownUnits() map[uint32]int {
	return i.unknownUnits
}

// HostileEntries returns the raw creature entry → Identity map.
func (i *Identifier) HostileEntries() map[uint32]Identity {
	return i.byEntryId
}
