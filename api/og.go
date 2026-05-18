package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Emyrk/chronicle/api/db2sdk"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Emyrk/chronicle/combatlog/parser/guid"
	"github.com/Emyrk/chronicle/database"
	"github.com/Emyrk/chronicle/database/authz"
	"github.com/Emyrk/chronicle/frontend"
)

// OGRoutes returns the Open Graph route definitions for the frontend handler.
func (api *API) OGRoutes() []frontend.OGRoute {
	return []frontend.OGRoute{
		{
			Pattern: "/instances/{idOrSlug}",
			Resolve: func(r *http.Request) *frontend.OGData {
				return api.instanceOG(chi.URLParam(r, "idOrSlug"))
			},
		},
		{
			Pattern: "/s/{code}",
			Resolve: func(r *http.Request) *frontend.OGData {
				return api.shareOG(chi.URLParam(r, "code"))
			},
		},
		{
			Pattern: "/armory/{realm}/{player}",
			Resolve: func(r *http.Request) *frontend.OGData {
				return api.armoryOG(chi.URLParam(r, "realm"), chi.URLParam(r, "player"))
			},
		},
	}
}

func (api *API) instanceOG(idOrSlug string) *frontend.OGData {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inst, err := resolveInstance(ctx, api.Opts.Zed, idOrSlug)
	if err != nil {
		return nil
	}

	return api.buildInstanceOG(ctx, inst)
}

func (api *API) shareOG(code string) *frontend.OGData {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := api.Opts.Zed
	shared, err := db.GetSharedViewByCode(ctx, code)
	if err != nil {
		return nil
	}

	// Resolve instance: try by ID first, fall back to slug if ID was nulled (reparse).
	var inst database.LogInstancesGuild
	if shared.InstanceID.Valid {
		inst, err = db.Instance(ctx, shared.InstanceID.UUID)
	} else if shared.InstanceSlug != "" {
		inst, err = db.InstanceBySlug(ctx, pgtype.Text{String: shared.InstanceSlug, Valid: true})
	} else {
		return nil
	}
	if err != nil {
		return nil
	}

	return api.buildInstanceOG(ctx, inst)
}

func (api *API) buildInstanceOG(ctx context.Context, inst database.LogInstancesGuild) *frontend.OGData {
	db := api.Opts.Zed

	encounters, _ := db.EncountersByInstanceID(ctx, inst.ID)
	players, _ := db.InstancePlayersByInstanceID(ctx, inst.ID)

	bossKills := 0
	var startDate time.Time
	var endDate time.Time
	var combatDuration time.Duration
	for _, e := range encounters {
		if e.Boss && (e.KillType == database.KillTypeClean || e.KillType == database.KillTypePartial) {
			bossKills++
		}
		if e.StartTime.Valid && e.EndTime.Valid {
			combatDuration += e.EndTime.Time.Sub(e.StartTime.Time)
		}
		if e.StartTime.Valid && (startDate.IsZero() || e.StartTime.Time.Before(startDate)) {
			startDate = e.StartTime.Time
		}
		if e.EndTime.Valid && (endDate.IsZero() || e.EndTime.Time.After(endDate)) {
			endDate = e.EndTime.Time
		}
	}

	dur := endDate.Sub(startDate)
	hours := int(dur.Hours())
	minutes := int(dur.Minutes()) % 60

	var title strings.Builder
	if inst.GuildName.String != "" {
		title.WriteString(fmt.Sprintf("%s — ", inst.GuildName.String))
	}
	title.WriteString(inst.Name)
	title.WriteString(" on [" + inst.RealmName + "]")

	var desc strings.Builder
	sep := " · "
	desc.WriteString(startDate.Format("Jan 2, 2006"))
	desc.WriteString(sep)

	desc.WriteString(fmt.Sprintf("%dh %dm", hours, minutes))
	desc.WriteString(sep)

	desc.WriteString(fmt.Sprintf("%d players", len(players)))
	desc.WriteString("\n")
	desc.WriteString("Raid performance and contribution analysis tool by Chronicle.")

	identifier := inst.ID.String()
	if inst.HashedSlug.Valid && inst.HashedSlug.String != "" {
		identifier = inst.HashedSlug.String
	}

	return &frontend.OGData{
		Title:       title.String(),
		Description: desc.String(),
		URL:         fmt.Sprintf("https://chronicleclassic.com/instances/%s", identifier),
	}
}

func (api *API) armoryOG(realm, player string) *frontend.OGData {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := api.Opts.Zed

	realmID, err := uuid.Parse(realm)
	if err != nil {
		r, dbErr := api.Opts.Zed.GetWoWServerRealmByName(ctx, realm)
		if dbErr != nil {
			return nil
		}
		realmID = r.ID
	}

	var identifier guid.GUID
	if g, parseErr := guid.FromString(player); parseErr == nil {
		identifier = g
	}

	p, err := db.GetGamePlayerByGUID(ctx, database.GetGamePlayerByGUIDParams{
		RealmID:    realmID,
		Identifier: identifier,
		Name:       player,
	})
	if err != nil {
		return nil
	}

	var title strings.Builder
	title.WriteString(fmt.Sprintf("%s — Character", p.Name))

	var desc strings.Builder
	guild := ""
	if p.GuildName.String != "" {
		guild = fmt.Sprintf(" <%s>", p.GuildName.String)
	}

	race := string(db2sdk.HeroRace(p.Race))
	switch p.Race {
	case database.WowPlayableRaceScourge:
		race = "Undead"
	case database.WowPlayableRaceBloodElf:
		race = "Blood Elf"
	case database.WowPlayableRaceNightElf:
		race = "Night Elf"
	}

	class := string(db2sdk.HeroClass(p.Class))
	class = strings.ToLower(class)
	class = strings.ToUpper(string(class[0])) + class[1:]

	desc.WriteString(fmt.Sprintf("%s%s (%s) — %d %s %s",
		p.Name, guild, p.RealmName,
		60, race, class,
	))

	return &frontend.OGData{
		Title:       title.String(),
		Description: desc.String(),
		URL:         fmt.Sprintf("https://chronicleclassic.com/armory/%s/%s", realm, p.ID),
	}
}

func resolveInstance(ctx context.Context, db *authz.Authz, idOrSlug string) (database.LogInstancesGuild, error) {
	id, err := uuid.Parse(idOrSlug)
	if err == nil {
		return db.Instance(ctx, id)
	}
	return db.InstanceBySlug(ctx, pgtype.Text{String: idOrSlug, Valid: true})
}
