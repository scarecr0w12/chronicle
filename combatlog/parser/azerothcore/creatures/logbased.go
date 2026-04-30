package creatures

import (
	"time"

	"github.com/Emyrk/chronicle/combatlog/parser/common/characters"
	"github.com/Emyrk/chronicle/combatlog/parser/common/characters/period"
	"github.com/Emyrk/chronicle/combatlog/parser/guid"
	"github.com/Emyrk/chronicle/combatlog/parser/vanilla/messages"
)

type LogBased struct {
	*characters.Base[*period.InactivityPeriod]
}

func NewLogBasedCharacter(id guid.GUID, all *characters.Characters) *LogBased {
	return &LogBased{
		Base: characters.NewBaseCharacter[*period.InactivityPeriod](id, all),
	}
}

func (c *LogBased) Process(m messages.Message) error {
	// Timeouts should be checked on every timestamp
	cur, ok := c.Activity.Current()
	if ok {
		cur.HandleTimeout(m.Date())
	}

	switch msg := m.(type) {
	case *messages.UnitCombatEnter:
		if msg.VictimGUID == c.ID() {
			c.Start("combat enter", m)
		}
	case *messages.UnitEvade:
		if msg.UnitGUID == c.ID() {
			c.End("evade", m, period.EndStateReset)
		}
	case *messages.Slain:
		if c.ID() == msg.Victim {
			c.Died(characters.ReasonSlain, m)
			return nil
		}

		// Pets are tied to their owners.
		owner, ok := c.Owner()
		if ok && owner == msg.Victim {
			c.Died(characters.ReasonOwnerSlain, m)
			return nil
		}
	}

	return nil
}

func (c *LogBased) Start(reason string, m messages.Message) {
	c.Activity.Start(
		period.NewInactivityPeriod(c.ID(), time.Minute),
		reason, m,
	)
}
