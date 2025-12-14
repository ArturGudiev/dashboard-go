-- Create container_children table with composite primary key on 4 columns
-- This replaces the Ent-generated table structure

-- Drop the table if it exists (be careful in production!)
DROP TABLE IF EXISTS container_children CASCADE;

-- Create the table with the exact structure you want
CREATE TABLE container_children (
    parent_type VARCHAR(50) NOT NULL,
    parent_id INTEGER NOT NULL,
    child_type VARCHAR(50) NOT NULL,
    child_id INTEGER NOT NULL,
    PRIMARY KEY (parent_type, parent_id, child_type, child_id),
    CONSTRAINT container_children_parent_fk 
        FOREIGN KEY (parent_id) 
        REFERENCES tasks(id) 
        ON DELETE CASCADE,
    CONSTRAINT container_children_child_fk 
        FOREIGN KEY (child_id) 
        REFERENCES tasks(id) 
        ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_container_children_parent_id ON container_children(parent_id);
CREATE INDEX idx_container_children_child_id ON container_children(child_id);
CREATE INDEX idx_container_children_parent_type ON container_children(parent_type);
CREATE INDEX idx_container_children_child_type ON container_children(child_type);

-- Add check constraints to ensure enum values
ALTER TABLE container_children 
ADD CONSTRAINT check_parent_type 
CHECK (parent_type IN ('epic', 'story', 'task', 'question', 'problem', 'knowledge-node', 'knowledge-bit', 'definition', 'action', 'scheduled-task', 'state'));

ALTER TABLE container_children 
ADD CONSTRAINT check_child_type 
CHECK (child_type IN ('epic', 'story', 'task', 'question', 'problem', 'knowledge-node', 'knowledge-bit', 'definition', 'action', 'scheduled-task', 'state'));


