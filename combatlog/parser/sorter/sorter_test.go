package sorter_test

import (
	"bytes"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Emyrk/chronicle/combatlog/parser/sorter"
	"github.com/Emyrk/chronicle/internal/testutil"
	"github.com/stretchr/testify/require"
)

// TestSortByTimestamp verifies that lines are sorted primarily by timestamp.
func TestSortByTimestamp(t *testing.T) {
	t.Parallel()

	logs := []string{
		"12/11 12:40:06.392  First event.",
		"12/11 12:40:06.549  Second event.",
		"12/11 12:40:06.710  Third event.",
		"12/11 12:40:07.015  Fourth event.",
	}

	// Shuffle and verify timestamp ordering is preserved
	for i := 0; i < 10; i++ {
		cpy := slices.Clone(logs)
		rand.Shuffle(len(cpy), func(i, j int) { cpy[i], cpy[j] = cpy[j], cpy[i] })

		logger := testutil.Logger(t)
		var out bytes.Buffer
		_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(cpy, "\n")), &out, false)
		require.NoError(t, err)

		got := removeEmpty(strings.Split(out.String(), "\n"))
		require.Equal(t, logs, got)
	}
}

// TestPriorityOrder verifies that at the same timestamp:
// 1. ZONE_INFO comes first
// 2. UNIT_INFO comes second
// 3. COMBATANT_INFO comes third
// 4. Other lines maintain their original input order
func TestPriorityOrder(t *testing.T) {
	t.Parallel()

	// All lines have the same timestamp - order should be deterministic based on type priority
	logs := []string{
		"12/11 12:40:06.593  ZONE_INFO: 11.12.25 12:52:59&blackrock spire&0",
		"12/11 12:40:06.593  UNIT_INFO: 11.12.25 12:40:06&0xF130001CF827939E&0&Mana Spring Totem IV&1&0x00000000000C270C&,10494=1`",
		"12/11 12:40:06.593  COMBATANT_INFO: 11.12.25 12:54:23&Maldrissa&WARLOCK&Orc&3&Chotuk&Exalted&5&nil",
		"12/11 12:40:06.593  Some regular event.",
	}

	// Shuffle and verify priority ordering
	for i := 0; i < 10; i++ {
		cpy := slices.Clone(logs)
		rand.Shuffle(len(cpy), func(i, j int) { cpy[i], cpy[j] = cpy[j], cpy[i] })

		logger := testutil.Logger(t)
		var out bytes.Buffer
		_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(cpy, "\n")), &out, false)
		require.NoError(t, err)

		got := removeEmpty(strings.Split(out.String(), "\n"))
		require.Equal(t, logs, got)
	}
}

// TestOriginalOrderPreserved verifies that lines with the same timestamp and no
// special priority maintain their original input order.
func TestOriginalOrderPreserved(t *testing.T) {
	t.Parallel()

	// All same timestamp, no priority types - should maintain input order
	logs := []string{
		"12/11 12:40:06.593  First regular event.",
		"12/11 12:40:06.593  Second regular event.",
		"12/11 12:40:06.593  Third regular event.",
		"12/11 12:40:06.593  Fourth regular event.",
	}

	// Without shuffling, order should be preserved
	logger := testutil.Logger(t)
	var out bytes.Buffer
	_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(logs, "\n")), &out, false)
	require.NoError(t, err)

	got := removeEmpty(strings.Split(out.String(), "\n"))
	require.Equal(t, logs, got)
}

// TestMixedTimestampsAndPriorities tests a realistic scenario with multiple
// timestamps and priority types.
func TestMixedTimestampsAndPriorities(t *testing.T) {
	t.Parallel()

	// Input in the order we'll provide it (not shuffled for this test)
	input := []string{
		"12/11 12:40:06.593  Regular event A.",
		"12/11 12:40:06.593  ZONE_INFO: 11.12.25 12:52:59&blackrock spire&0",
		"12/11 12:40:06.593  Regular event B.",
		"12/11 12:40:06.593  UNIT_INFO: 11.12.25 12:40:06&0xF130001CF827939E&0&Mana Spring&1&0x00000000000C270C&,10494=1`",
	}

	// Expected: ZONE_INFO first, then UNIT_INFO, then regular events in original order
	expected := []string{
		"12/11 12:40:06.593  ZONE_INFO: 11.12.25 12:52:59&blackrock spire&0",
		"12/11 12:40:06.593  UNIT_INFO: 11.12.25 12:40:06&0xF130001CF827939E&0&Mana Spring&1&0x00000000000C270C&,10494=1`",
		"12/11 12:40:06.593  Regular event A.",
		"12/11 12:40:06.593  Regular event B.",
	}

	logger := testutil.Logger(t)
	var out bytes.Buffer
	_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(input, "\n")), &out, false)
	require.NoError(t, err)

	got := removeEmpty(strings.Split(out.String(), "\n"))
	require.Equal(t, expected, got)
}

// TestSortByTimestamp_EpochMillis verifies that epoch-millis lines sort by timestamp.
func TestSortByTimestamp_EpochMillis(t *testing.T) {
	t.Parallel()

	logs := []string{
		"1777515101755  SPELL_DAMAGE,0x000000000001D2D8,\"Skulkemage\",0x548,0xF130002964000014,\"Mother Smolderweb\",0xa48,10181,\"Frostbolt\",0x10,978,943,16,0,0,0,nil,nil,nil",
		"1777515101757  SPELL_DAMAGE,0x000000000001D2D8,\"Skulkemage\",0x548,0xF130002964000014,\"Mother Smolderweb\",0xa48,10181,\"Frostbolt\",0x10,500,480,8,0,0,0,nil,nil,nil",
		"1777515101764  CHRONICLE_ENCOUNTER_CREDIT,229,45,0,10596,0,270,\"Mother Smolderweb\",0,0xF130002964000014,\"Mother Smolderweb\"",
		"1777515101800  SPELL_CAST_SUCCESS,0xF130002388000170,\"Rage Talon Dragonspawn\",0xa18,0xF130002388000170,\"Rage Talon Dragonspawn\",0xa18,8876,\"Thrash\",0x1",
	}

	for i := 0; i < 10; i++ {
		cpy := slices.Clone(logs)
		rand.Shuffle(len(cpy), func(i, j int) { cpy[i], cpy[j] = cpy[j], cpy[i] })

		logger := testutil.Logger(t)
		var out bytes.Buffer
		_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(cpy, "\n")), &out, true)
		require.NoError(t, err)

		got := removeEmpty(strings.Split(out.String(), "\n"))
		require.Equal(t, logs, got)
	}
}

// TestPriorityOrder_EpochMillis verifies AzerothCore meta-event priority ordering.
func TestPriorityOrder_EpochMillis(t *testing.T) {
	t.Parallel()

	// All lines share the same timestamp — order determined by priority.
	logs := []string{
		`1777515068242  CHRONICLE_HEADER,"","3.3.5a",12340`,
		`1777515068242  CHRONICLE_ZONE_INFO,"Blackrock Spire",229,45,"party"`,
		`1777515068242  CHRONICLE_UNIT_INFO,0xF130002388000170,"Rage Talon Dragonspawn",59,0xa28,0x0000000000000000,5707,"NEUTRAL",false`,
		`1777515068242  SPELL_CAST_SUCCESS,0xF130002388000170,"Rage Talon Dragonspawn",0xa18,0xF130002388000170,"Rage Talon Dragonspawn",0xa18,8876,"Thrash",0x1`,
	}

	for i := 0; i < 10; i++ {
		cpy := slices.Clone(logs)
		rand.Shuffle(len(cpy), func(i, j int) { cpy[i], cpy[j] = cpy[j], cpy[i] })

		logger := testutil.Logger(t)
		var out bytes.Buffer
		_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(cpy, "\n")), &out, true)
		require.NoError(t, err)

		got := removeEmpty(strings.Split(out.String(), "\n"))
		require.Equal(t, logs, got)
	}
}

// TestEpochMillisRoundTrip verifies output preserves epoch-millis format.
func TestEpochMillisRoundTrip(t *testing.T) {
	t.Parallel()

	input := []string{
		`1777515068242  CHRONICLE_HEADER,"","3.3.5a",12340`,
		`1777515101755  SPELL_DAMAGE,0x000000000001D2D8,"Skulkemage",0x548`,
	}

	logger := testutil.Logger(t)
	var out bytes.Buffer
	_, _, err := sorter.SortLogs(t.Context(), logger, strings.NewReader(strings.Join(input, "\n")), &out, true)
	require.NoError(t, err)

	got := removeEmpty(strings.Split(out.String(), "\n"))
	require.Equal(t, input, got)

	// Verify lines start with numeric epoch timestamps, not date format.
	for _, line := range got {
		parts := strings.SplitN(line, "  ", 2)
		require.Len(t, parts, 2)
		_, err := strconv.ParseInt(parts[0], 10, 64)
		require.NoError(t, err, "expected epoch-millis timestamp in output")
	}
}

func removeEmpty(lines []string) []string {
	cpy := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			cpy = append(cpy, line)
		}
	}
	return cpy
}
