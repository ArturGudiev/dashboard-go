package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/dialect/entsql"
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
			Values(
				"epic",
				"story",
				"task",
				"question",
				"problem",
				"knowledge-node",
				"knowledge-bit",
				"definition",
				"action",
				"scheduled-task",
				"state",
			),
		field.Int("parent_id").
			Positive(),
		field.Enum("child_type").
			Values(
				"epic",
				"story",
				"task",
				"question",
				"problem",
				"knowledge-node",
				"knowledge-bit",
				"definition",
				"action",
				"scheduled-task",
				"state",
			),
		field.Int("child_id").
			Positive(),
	}
}

// Edges of the ContainerChild.
func (ContainerChild) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("parent", Task.Type).
			Unique().
			Required().
			Field("parent_id"),
		edge.To("child", Task.Type).
			Unique().
			Required().
			Field("child_id"),
	}
}

// Indexes of the ContainerChild.
func (ContainerChild) Indexes() []ent.Index {
	return []ent.Index{
		// Composite unique index on all 4 columns (acts as composite key for business logic)
		// This ensures no duplicate relationships with the same types
		index.Fields("parent_type", "parent_id", "child_type", "child_id").
			Unique(),
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


