CREATE TABLE world_creature_spawn (
    guid integer NOT NULL,
    id integer DEFAULT 0 NOT NULL,
    id2 integer DEFAULT 0 NOT NULL,
    id3 integer DEFAULT 0 NOT NULL,
    id4 integer DEFAULT 0 NOT NULL,
    map integer DEFAULT 0 NOT NULL
);

CREATE INDEX idx_world_creature_spawn_id ON world_creature_spawn (id);

ALTER TABLE ONLY world_creature_spawn
    ADD CONSTRAINT world_creature_spawn_pkey PRIMARY KEY (guid);
