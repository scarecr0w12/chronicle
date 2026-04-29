package instances

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlackfathomDeepsHostiles_BossCoverage(t *testing.T) {
	t.Parallel()

	hostiles := BlackfathomDeepsHostiles()
	require.True(t, hostiles[4829].Boss)
	require.Equal(t, "Aku'mai", hostiles[4829].EncounterName)
	require.True(t, hostiles[4832].Boss)
	require.Equal(t, "Twilight Lord Kelris", hostiles[4832].EncounterName)
	require.True(t, hostiles[6243].Boss)
	require.Equal(t, "Gelihast", hostiles[6243].EncounterName)
	require.True(t, hostiles[4810].Hostile)
	require.False(t, hostiles[4810].Boss)
	require.False(t, hostiles[4787].Hostile)
}
