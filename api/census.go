package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Emyrk/chronicle/api/chroniclesdk"
	"github.com/Emyrk/chronicle/api/db2sdk"
	"github.com/Emyrk/chronicle/api/httpapi"
	"github.com/Emyrk/chronicle/database"
)

// Census returns player counts grouped by class and race.
//
//	@Summary	Census player counts
//	@Tags		Census
//	@Produce	json
//	@Param		days		query	int		false	"Number of days to look back (default 90, max 365)"
//	@Param		realm_id	query	[]string	false	"Realm IDs to filter by (repeatable)"
//	@Success	200	{array}	chroniclesdk.CensusEntry
//	@Router		/census [get]
func (api *API) Census(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse days parameter (default 90, clamp 1-365)
	days := 90
	if d := r.URL.Query().Get("days"); d != "" {
		parsed, err := strconv.Atoi(d)
		if err != nil || parsed < 1 {
			parsed = 1
		} else if parsed > 365 {
			parsed = 365
		}
		days = parsed
	}

	// Parse realm_id parameters
	realmStrs := r.URL.Query()["realm_id"]
	realmIDs := make([]uuid.UUID, 0, len(realmStrs))
	for _, s := range realmStrs {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		realmIDs = append(realmIDs, id)
	}

	updatedAfter := time.Now().AddDate(0, 0, -days)

	rows, err := api.Opts.Zed.CensusPlayerCounts(ctx, database.CensusPlayerCountsParams{
		UpdatedAfter: pgtype.Timestamptz{
			Time:  updatedAfter,
			Valid: true,
		},
		RealmIds: realmIDs,
	})
	if err != nil {
		httpapi.HandleResponseError(ctx, w, err, httpapi.APIError{
			Response: chroniclesdk.Response{
				Message: "Failed to fetch census data",
				Detail:  err.Error(),
			},
		})
		return
	}

	entries := make([]chroniclesdk.CensusEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, chroniclesdk.CensusEntry{
			Class: db2sdk.HeroClass(row.Class).String(),
			Race:  db2sdk.HeroRace(row.Race).String(),
			Count: row.Count,
		})
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	httpapi.Write(ctx, w, http.StatusOK, entries)
}
