package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.Strings("tags").
			Default([]string{}),
		field.Bool("done").
			Default(false),
		field.String("notes").
			Default(""),
		field.Time("done_date_time").
			Optional().
			Nillable(),
	}
}

// Edges of the Task.
//func (Task) Edges() []ent.Edge {
//	// Note: Parent-child relationships are handled through ContainerChild join table
//	// using parent_type/child_type enums, not through Ent edges
//	// This allows polymorphic relationships (tasks, problems, etc.)
//	return []ent.Edge{}
//}
