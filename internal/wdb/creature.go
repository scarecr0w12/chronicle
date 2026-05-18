package wdb

import "fmt"

// Creature represents a parsed creature from creaturecache.wdb.
// Field layout varies by client build; see ParseCreature for version-aware parsing.
type Creature struct {
	Entry        uint32
	Name         string
	Name2        string
	Name3        string
	Name4        string
	SubName      string
	IconName     string // cursor icon (WotLK+)
	TypeFlags    uint32
	Type         uint32 // CreatureType (beast, humanoid, etc.)
	Family       uint32
	Rank         uint32 // normal, elite, rare, etc.
	KillCredit   [2]uint32   // WotLK+
	DisplayID    [4]uint32   // vanilla: only [0]; WotLK: all 4
	HpMulti      float32     // WotLK+
	ManaMulti    float32     // WotLK+
	RacialLeader uint8
	QuestItems   [6]uint32   // WotLK+
	MovementID   uint32      // WotLK+
}

// ParseCreature parses a single creature record from creaturecache.wdb.
// build is the client build number from the WDB header (Header.Version).
func ParseCreature(rec Record, build uint32) (Creature, error) {
	c := Creature{Entry: rec.EntryID}
	r := newReader(rec.Data)
	var err error

	u := func() uint32 { var v uint32; if err == nil { v, err = r.Uint32() }; return v }
	f := func() float32 { var v float32; if err == nil { v, err = r.Float32() }; return v }
	s := func() string { var v string; if err == nil { v, err = r.String() }; return v }

	c.Name = s()
	c.Name2 = s()
	c.Name3 = s()
	c.Name4 = s()
	c.SubName = s()

	if build >= buildWotLK {
		c.IconName = s()
	}

	c.TypeFlags = u()
	c.Type = u()
	c.Family = u()
	c.Rank = u()

	if build >= buildWotLK {
		c.KillCredit[0] = u()
		c.KillCredit[1] = u()
		c.DisplayID[0] = u()
		c.DisplayID[1] = u()
		c.DisplayID[2] = u()
		c.DisplayID[3] = u()
		c.HpMulti = f()
		c.ManaMulti = f()
	} else {
		// Vanilla/TBC: two unknown u32 fields (unk padding + SpellDataID),
		// then a single display ID.
		_ = u() // unk
		_ = u() // SpellDataID (pet spell data, not stored)
		c.DisplayID[0] = u()
		// Civilian flag (u8) — vanilla only, not stored.
		if err == nil {
			_, err = r.Uint8()
		}
	}

	if err == nil {
		var b byte
		b, err = r.Uint8()
		c.RacialLeader = b
	}

	if build >= buildWotLK {
		for j := range 6 {
			c.QuestItems[j] = u()
		}
		c.MovementID = u()
	}

	if err != nil {
		return c, fmt.Errorf("parse creature %d: %w", rec.EntryID, err)
	}
	return c, nil
}
