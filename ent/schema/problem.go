package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Problem holds the schema definition for the Problem entity.
type Problem struct {
	ent.Schema
}

// Fields of the Problem.
func (Problem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.Strings("tags").
			Default([]string{}),
		field.String("notes").
			Default(""),
		field.Time("done_date_time").
			Optional().
			Nillable(),
		field.String("solution").
			Optional().
			Nillable(),
	}
}

// Edges of the Problem.
func (Problem) Edges() []ent.Edge {
	// Note: Parent-child relationships are handled through ContainerChild join table
	// using parent_type/child_type enums, not through Ent edges
	return []ent.Edge{}
}
