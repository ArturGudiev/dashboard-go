-- Drop the old task_children table since we're now using container_children
-- This table is no longer needed as all relationships are handled in container_children

DROP TABLE IF EXISTS task_children CASCADE;


