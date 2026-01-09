package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ContainerChild holds the schema definition for the ContainerChild entity (join table).
// This table represents parent-child relationships between containers of various types.
type ContainerChild struct {
	ent.Schema
}

// Fields of the ContainerChild.
func (ContainerChild) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.Enum("parent_type").
			GoType(ContainerType("")),
		field.Int("parent_id").
			Positive(),
		field.Enum("child_type").
			GoType(ContainerType("")),
		field.Int("child_id").
			Positive(),
		// child_order: position of this child within the parent's children list
		// When querying children of a parent, order by this field
		field.Int("child_order").
			Default(0).
			NonNegative(),
		// parent_order: position of this parent within the child's parents list
		// When querying parents of a child, order by this field
		field.Int("parent_order").
			Default(0).
			NonNegative(),
	}
}

// Indexes of the ContainerChild.
func (ContainerChild) Indexes() []ent.Index {
	return []ent.Index{
		// Composite unique index on all 4 columns (acts as composite key for business logic)
		// This ensures no duplicate relationships with the same types
		index.Fields("parent_type", "parent_id", "child_type", "child_id").
			Unique(),
		// Index for efficient ordered queries: get children of a parent ordered by child_order
		index.Fields("parent_id", "child_order"),
		// Index for efficient ordered queries: get parents of a child ordered by parent_order
		index.Fields("child_id", "parent_order"),
		// Indexes for querying by parent or child
		index.Fields("parent_id"),
		index.Fields("child_id"),
	}
}

// Annotations of the ContainerChild.
func (ContainerChild) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Explicitly set table name to container_children
		entsql.Annotation{
			Table: "container_children",
		},
	}
}
