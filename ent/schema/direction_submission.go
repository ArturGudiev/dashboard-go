package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DirectionSubmission holds the schema definition for the DirectionSubmission entity.
type DirectionSubmission struct {
	ent.Schema
}

// Fields of the DirectionSubmission.
func (DirectionSubmission) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("direction_id").
			Positive(),
		field.Time("execution_date"),
		field.String("text").
			Optional().
			Nillable().
			StructTag(`json:"text"`),
	}
}

// Edges of the DirectionSubmission.
func (DirectionSubmission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("direction", Direction.Type).
			Ref("submissions").
			Unique().
			Required().
			Field("direction_id"),
	}
}
