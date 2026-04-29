package registry

import (
	"testing"

	"github.com/Emyrk/chronicle/combatlog/parser/types/zone"
	"github.com/Emyrk/chronicle/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestCommonFactoryFromTemplate_PreservesMapIDMatching(t *testing.T) {
	t.Parallel()

	factory := commonFactoryFromTemplate(
		database.WorldInstanceTemplate{
			ID:    uuid.New(),
			Name:  "Ulduar",
			MapID: pgtype.Int4{Int32: 603, Valid: true},
		},
		[]database.WorldInstanceZoneName{{ZoneName: "ulduar", DisplayName: "Ulduar"}},
		nil,
	)

	require.True(t, factory.MatchZone(zone.Zone{Name: "unknown", MapID: 603}))
	require.False(t, factory.MatchZone(zone.Zone{Name: "unknown", MapID: 604}))
	require.True(t, factory.MatchZone(zone.Zone{Name: "ulduar"}))
	require.Equal(t, []uint32{603}, factory.MapIDs)
	require.Equal(t, "Ulduar", factory.Name)
	require.Equal(t, []string{"ulduar"}, factory.ZoneNames)
}
