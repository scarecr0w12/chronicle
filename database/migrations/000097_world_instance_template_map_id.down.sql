DROP INDEX IF EXISTS idx_world_instance_template_map_id;

ALTER TABLE world_instance_template
  DROP COLUMN map_id;