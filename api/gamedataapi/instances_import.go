package gamedataapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Emyrk/chronicle/api/chroniclesdk"
	"github.com/Emyrk/chronicle/api/httpapi"
	"github.com/Emyrk/chronicle/database"
	"github.com/jackc/pgx/v5/pgtype"
)

const maxInstanceDumpFileSize = 20 * 1024 * 1024

type invalidInstanceDumpError struct {
	message string
}

func (e invalidInstanceDumpError) Error() string {
	return e.message
}

func newInvalidInstanceDumpError(format string, args ...interface{}) error {
	return invalidInstanceDumpError{message: fmt.Sprintf(format, args...)}
}

func isInvalidInstanceDumpError(err error) bool {
	var invalidErr invalidInstanceDumpError
	return errors.As(err, &invalidErr)
}

func (h *Handler) ImportInstanceDump(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, ok := readInstanceDumpRequest(ctx, w, r)
	if !ok {
		return
	}

	result, err := importWorldInstanceDump(ctx, database.New(h.pool), req)
	if err != nil {
		if isInvalidInstanceDumpError(err) {
			httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
				Message: "Invalid instance dump.",
				Detail:  err.Error(),
			})
			return
		}
		httpapi.InternalServerError(w, err)
		return
	}

	h.notifyInstanceDataChanged()
	httpapi.Write(ctx, w, http.StatusOK, result)
}

func readInstanceDumpRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (chroniclesdk.ImportWorldInstanceDumpRequest, bool) {
	var req chroniclesdk.ImportWorldInstanceDumpRequest

	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		file, header, err := r.FormFile("dump_file")
		if err != nil {
			httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
				Message: "Failed to get dump_file from form",
				Detail:  err.Error(),
			})
			return req, false
		}
		defer func() { _ = file.Close() }()

		if header.Size > maxInstanceDumpFileSize {
			httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
				Message: fmt.Sprintf("File too large (%d bytes), maximum is %d bytes", header.Size, maxInstanceDumpFileSize),
			})
			return req, false
		}

		if err := json.NewDecoder(file).Decode(&req); err != nil {
			httpapi.Write(ctx, w, http.StatusBadRequest, chroniclesdk.Response{
				Message: "Failed to parse instance dump JSON",
				Detail:  err.Error(),
			})
			return req, false
		}
		return req, true
	}

	if !httpapi.Read(ctx, w, r, &req) {
		return req, false
	}

	return req, true
}

func importWorldInstanceDump(
	ctx context.Context,
	store database.Store,
	req chroniclesdk.ImportWorldInstanceDumpRequest,
) (chroniclesdk.ImportWorldInstanceDumpResponse, error) {
	if len(req.Instances) == 0 {
		return chroniclesdk.ImportWorldInstanceDumpResponse{}, newInvalidInstanceDumpError("instance dump is empty")
	}

	seenNames := make(map[string]int, len(req.Instances))
	for index, inst := range req.Instances {
		normalizedName := strings.ToLower(strings.TrimSpace(inst.Name))
		if normalizedName == "" {
			return chroniclesdk.ImportWorldInstanceDumpResponse{}, newInvalidInstanceDumpError("instance %d: name is required", index)
		}
		if previousIndex, ok := seenNames[normalizedName]; ok {
			return chroniclesdk.ImportWorldInstanceDumpResponse{}, newInvalidInstanceDumpError(
				"instance %d (%s): duplicate instance name also provided at index %d",
				index,
				inst.Name,
				previousIndex,
			)
		}
		seenNames[normalizedName] = index
	}

	result := chroniclesdk.ImportWorldInstanceDumpResponse{}
	err := store.InTx(func(tx database.Store) error {
		for index, inst := range req.Instances {
			if inst.Category == "" {
				return newInvalidInstanceDumpError("instance %d (%s): category is required", index, inst.Name)
			}

			abbrev := pgtype.Text{}
			if inst.Abbreviation != "" {
				abbrev = pgtype.Text{String: inst.Abbreviation, Valid: true}
			}
			bossCount := pgtype.Int4{}
			if inst.BossCount != nil {
				bossCount = pgtype.Int4{Int32: *inst.BossCount, Valid: true}
			}
			mapID := pgtype.Int4{}
			if inst.MapID != nil {
				mapID = pgtype.Int4{Int32: *inst.MapID, Valid: true}
			}
			background := pgtype.Text{}
			if inst.Background != "" {
				background = pgtype.Text{String: inst.Background, Valid: true}
			}

			tmpl, err := tx.UpsertWorldInstanceTemplate(ctx, database.UpsertWorldInstanceTemplateParams{
				Name:         inst.Name,
				Abbreviation: abbrev,
				Category:     database.InstanceCategory(inst.Category),
				MapID:        mapID,
				BossCount:    bossCount,
				Background:   background,
			})
			if err != nil {
				return fmt.Errorf("instance %d (%s): upsert template: %w", index, inst.Name, err)
			}

			if err := tx.DeleteWorldInstanceZoneNames(ctx, tmpl.ID); err != nil {
				return fmt.Errorf("instance %d (%s): delete zone names: %w", index, inst.Name, err)
			}
			if err := tx.DeleteWorldInstanceUnits(ctx, tmpl.ID); err != nil {
				return fmt.Errorf("instance %d (%s): delete units: %w", index, inst.Name, err)
			}

			for _, zoneName := range inst.ZoneNames {
				if zoneName.ZoneName == "" {
					return newInvalidInstanceDumpError("instance %d (%s): zone_name is required", index, inst.Name)
				}
				if zoneName.DisplayName == "" {
					return newInvalidInstanceDumpError("instance %d (%s): display_name is required", index, inst.Name)
				}
				if err := tx.InsertWorldInstanceZoneName(ctx, database.InsertWorldInstanceZoneNameParams{
					InstanceID:  tmpl.ID,
					ZoneName:    zoneName.ZoneName,
					DisplayName: zoneName.DisplayName,
				}); err != nil {
					return fmt.Errorf("instance %d (%s): insert zone name: %w", index, inst.Name, err)
				}
				result.ZoneNamesImported++
			}

			for _, unit := range inst.Units {
				if unit.EntryID == 0 {
					return newInvalidInstanceDumpError("instance %d (%s): unit entry_id is required", index, inst.Name)
				}
				encounterName := pgtype.Text{}
				if unit.EncounterName != "" {
					encounterName = pgtype.Text{String: unit.EncounterName, Valid: true}
				}
				overrideName := pgtype.Text{}
				if unit.OverrideName != "" {
					overrideName = pgtype.Text{String: unit.OverrideName, Valid: true}
				}
				if err := tx.UpsertWorldInstanceUnit(ctx, database.UpsertWorldInstanceUnitParams{
					InstanceID:    tmpl.ID,
					EntryID:       unit.EntryID,
					OverrideName:  overrideName,
					EncounterName: encounterName,
					Boss:          unit.Boss,
					Affiliation:   database.UnitAffiliation(unit.Affiliation),
				}); err != nil {
					return fmt.Errorf("instance %d (%s): upsert unit %d: %w", index, inst.Name, unit.EntryID, err)
				}
				result.UnitsImported++
			}

			result.InstancesImported++
		}
		return nil
	}, nil)
	if err != nil {
		return chroniclesdk.ImportWorldInstanceDumpResponse{}, err
	}

	return result, nil
}