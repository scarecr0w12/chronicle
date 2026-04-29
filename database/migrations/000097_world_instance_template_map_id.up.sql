ALTER TABLE world_instance_template
  ADD COLUMN map_id INT;

CREATE INDEX idx_world_instance_template_map_id
  ON world_instance_template(map_id)
  WHERE map_id IS NOT NULL;