package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Direction holds the schema definition for the Direction entity.
type Direction struct {
	ent.Schema
}

// Fields of the Direction.
func (Direction) Fields() []ent.Field {
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
		field.Bool("closed").
			Default(false),
	}
}

// Edges of the Direction.
func (Direction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("submissions", DirectionSubmission.Type),
	}
}
