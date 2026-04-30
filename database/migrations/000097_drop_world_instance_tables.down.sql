CREATE TYPE instance_category AS ENUM ('raid', 'dungeon', 'pvp');

CREATE TYPE unit_affiliation AS ENUM ('unknown', 'friendly', 'neutral', 'hostile', 'vary');

CREATE TABLE world_instance_template (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT NOT NULL,
  abbreviation  TEXT,
  category      instance_category NOT NULL,
  boss_count    INT,
  background    TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(name)
);

CREATE TABLE world_instance_zone_names (
  instance_id   UUID NOT NULL REFERENCES world_instance_template(id) ON DELETE CASCADE,
  zone_name     TEXT NOT NULL,
  display_name  TEXT NOT NULL,

  PRIMARY KEY (instance_id, zone_name)
);

CREATE INDEX idx_world_instance_zone_name
  ON world_instance_zone_names(zone_name) INCLUDE (instance_id);

CREATE TABLE world_instance_units (
  instance_id    UUID NOT NULL REFERENCES world_instance_template(id) ON DELETE CASCADE,
  entry_id       INT NOT NULL,
  override_name  TEXT,
  encounter_name TEXT,
  boss           BOOLEAN NOT NULL DEFAULT FALSE,
  affiliation    unit_affiliation NOT NULL DEFAULT 'hostile',

  PRIMARY KEY (instance_id, entry_id)
);
