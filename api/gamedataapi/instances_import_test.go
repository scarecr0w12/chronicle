package gamedataapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Emyrk/chronicle/api/chroniclesdk"
	"github.com/Emyrk/chronicle/database/dbtestutil"
	"github.com/Emyrk/chronicle/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestImportWorldInstanceDump_ReplacesTemplateZonesAndUnits(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)

	store, _ := dbtestutil.NewDB(t)
	mapID := int32(48)
	bossCount := int32(7)

	first := chroniclesdk.ImportWorldInstanceDumpRequest{
		Instances: []chroniclesdk.WorldInstanceDumpEntry{{
			Name:      "Blackfathom Deeps",
			Category:  chroniclesdk.InstanceCategoryDungeon,
			MapID:     &mapID,
			BossCount: &bossCount,
			ZoneNames: []chroniclesdk.WorldInstanceZoneName{{ZoneName: "blackfathom deeps", DisplayName: "Blackfathom Deeps"}},
			Units: []chroniclesdk.WorldInstanceUnit{{
				EntryID:       4829,
				EncounterName: "Aku'mai",
				Boss:          true,
				Affiliation:   chroniclesdk.UnitAffiliationHostile,
			}},
		}},
	}

	result, err := importWorldInstanceDump(ctx, store, first)
	require.NoError(t, err)
	require.Equal(t, 1, result.InstancesImported)
	require.Equal(t, 1, result.ZoneNamesImported)
	require.Equal(t, 1, result.UnitsImported)

	templates, err := store.ListWorldInstanceTemplates(ctx)
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.Equal(t, int32(48), templates[0].MapID.Int32)

	zoneNames, err := store.GetWorldInstanceZoneNames(ctx, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, zoneNames, 1)
	require.Equal(t, "blackfathom deeps", zoneNames[0].ZoneName)

	units, err := store.GetWorldInstanceUnits(ctx, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, units, 1)
	require.Equal(t, int32(4829), units[0].EntryID)
	require.True(t, units[0].Boss)

	updatedMapID := int32(70)
	second := chroniclesdk.ImportWorldInstanceDumpRequest{
		Instances: []chroniclesdk.WorldInstanceDumpEntry{{
			Name:     "Blackfathom Deeps",
			Category: chroniclesdk.InstanceCategoryDungeon,
			MapID:    &updatedMapID,
			ZoneNames: []chroniclesdk.WorldInstanceZoneName{
				{ZoneName: "blackfathom deeps", DisplayName: "Blackfathom Deeps"},
				{ZoneName: "blackfathom depths", DisplayName: "Blackfathom Deeps"},
			},
			Units: []chroniclesdk.WorldInstanceUnit{{
				EntryID:     4832,
				Boss:        true,
				Affiliation: chroniclesdk.UnitAffiliationHostile,
			}},
		}},
	}

	_, err = importWorldInstanceDump(ctx, store, second)
	require.NoError(t, err)

	templates, err = store.ListWorldInstanceTemplates(ctx)
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.Equal(t, int32(70), templates[0].MapID.Int32)

	zoneNames, err = store.GetWorldInstanceZoneNames(ctx, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, zoneNames, 2)

	units, err = store.GetWorldInstanceUnits(ctx, templates[0].ID)
	require.NoError(t, err)
	require.Len(t, units, 1)
	require.Equal(t, int32(4832), units[0].EntryID)
	require.True(t, units[0].Boss)

	_, err = store.GetWorldInstanceTemplateByZoneName(ctx, "blackfathom depths")
	require.NoError(t, err)
	_, err = store.GetWorldInstanceTemplateByZoneName(ctx, "blackfathom deeps")
	require.NoError(t, err)
}

func TestImportWorldInstanceDump_RejectsEmptyPayload(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)

	store, _ := dbtestutil.NewDB(t)
	_, err := importWorldInstanceDump(ctx, store, chroniclesdk.ImportWorldInstanceDumpRequest{})
	require.Error(t, err)
	require.True(t, isInvalidInstanceDumpError(err))
	require.Contains(t, err.Error(), "instance dump is empty")
}

func TestImportWorldInstanceDump_RejectsDuplicateNames(t *testing.T) {
	t.Parallel()
	ctx := testutil.Context(t, testutil.WaitShort)

	store, _ := dbtestutil.NewDB(t)
	_, err := importWorldInstanceDump(ctx, store, chroniclesdk.ImportWorldInstanceDumpRequest{
		Instances: []chroniclesdk.WorldInstanceDumpEntry{
			{Name: "Blackfathom Deeps", Category: chroniclesdk.InstanceCategoryDungeon},
			{Name: " blackfathom deeps ", Category: chroniclesdk.InstanceCategoryDungeon},
		},
	})
	require.Error(t, err)
	require.True(t, isInvalidInstanceDumpError(err))
	require.Contains(t, err.Error(), "duplicate instance name")
}

func TestImportInstanceDump_ReturnsBadRequestForInvalidDump(t *testing.T) {
	t.Parallel()

	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/instances/import", strings.NewReader(`{
		"instances": [
			{"name":"Blackfathom Deeps","category":"dungeon"},
			{"name":"blackfathom deeps","category":"dungeon"}
		]
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ImportInstanceDump(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "Invalid instance dump.")
	require.Contains(t, rr.Body.String(), "duplicate instance name")
}