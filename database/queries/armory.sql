-- name: UpsertGuild :one
INSERT INTO
  guilds (realm_id, name, created_at)
VALUES
  ($1, $2, $3)
ON CONFLICT (realm_id, name) DO UPDATE
  SET realm_id = EXCLUDED.realm_id  -- no-op, just to return the row
RETURNING *
;


-- name: UpsertPlayers :batchexec
INSERT INTO
  game_players (
    id, realm_id, name, guild_id,
    class, gender, race,
    gear, level, talents,
    updated_from_instance,
    updated_at
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id, realm_id) DO UPDATE
  SET name = EXCLUDED.name,
      guild_id = EXCLUDED.guild_id,
      class = EXCLUDED.class,
      race = EXCLUDED.race,
      gender = EXCLUDED.gender,
      gear = EXCLUDED.gear,
      level = EXCLUDED.level,
      talents = CASE
        WHEN EXCLUDED.talents IS NOT NULL AND EXCLUDED.talents != 'null'::jsonb
        THEN EXCLUDED.talents
        ELSE game_players.talents
      END,
      updated_from_instance = EXCLUDED.updated_from_instance,

      updated_at = EXCLUDED.updated_at
WHERE
  EXCLUDED.updated_at > game_players.updated_at;
;


-- name: GetGamePlayerByGUID :one
SELECT
  gp.*,
  COALESCE(wow_server_realms.name, 'Unknown') as realm_name,
  g.name as guild_name
FROM
  game_players gp
LEFT JOIN guilds g ON g.id = gp.guild_id
LEFT JOIN wow_server_realms ON gp.realm_id = wow_server_realms.id
WHERE
  gp.realm_id = @realm_id
  AND (gp.id = @identifier::wow_guid OR lower(gp.name) = lower(@name))
;


-- name: SearchGamePlayers :many
SELECT
  gp.id,
  gp.realm_id,
  gp.name,
  gp.class,
  gp.race,
  gp.gender,
  gp.level,
  gp.guild_id,
  gp.updated_at,
  COALESCE(wow_server_realms.name, 'Unknown') as realm_name,
  COALESCE(g.name, '') as guild_name
FROM
  game_players gp
LEFT JOIN guilds g ON g.id = gp.guild_id
LEFT JOIN wow_server_realms ON gp.realm_id = wow_server_realms.id
WHERE
  gp.name ILIKE @search_term || '%'
  AND CASE
    WHEN @filter_class::text != '' THEN gp.class::text = @filter_class
    ELSE true
  END
  AND CASE
    WHEN @filter_realm::uuid != '00000000-0000-0000-0000-000000000000' THEN gp.realm_id = @filter_realm
    ELSE true
  END
  AND CASE
    WHEN @filter_guild::text != '' THEN g.name ILIKE '%' || @filter_guild || '%'
    ELSE true
  END
ORDER BY
  gp.level DESC, gp.updated_at DESC
LIMIT @result_limit
OFFSET @result_offset
;

-- name: CensusPlayerCounts :many
SELECT
  class,
  race,
  COUNT(*) AS count
FROM game_players
WHERE updated_at >= @updated_after::timestamptz
  AND (cardinality(@realm_ids::uuid[]) = 0 OR realm_id = ANY(@realm_ids::uuid[]))
GROUP BY class, race
ORDER BY class, race;
