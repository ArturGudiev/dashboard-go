-- Drop foreign key constraints from container_children table
-- These constraints prevent polymorphic relationships (tasks, problems, etc.)
-- Since ContainerChild is polymorphic, we rely on application-level validation
-- instead of database foreign keys

-- Drop the foreign key constraints
ALTER TABLE container_children 
DROP CONSTRAINT IF EXISTS container_children_tasks_parent;

ALTER TABLE container_children 
DROP CONSTRAINT IF EXISTS container_children_tasks_child;

-- Also drop any constraints that might have been created by the SQL file
ALTER TABLE container_children 
DROP CONSTRAINT IF EXISTS container_children_parent_fk;

ALTER TABLE container_children 
DROP CONSTRAINT IF EXISTS container_children_child_fk;

