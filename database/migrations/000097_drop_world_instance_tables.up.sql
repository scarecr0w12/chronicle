BEGIN;

DROP TABLE IF EXISTS world_instance_units;
DROP TABLE IF EXISTS world_instance_zone_names;
DROP TABLE IF EXISTS world_instance_template;
DROP TYPE IF EXISTS unit_affiliation;
DROP TYPE IF EXISTS instance_category;

COMMIT;