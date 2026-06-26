package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type StateRequirement struct {
	ent.Schema
}

func (StateRequirement) Fields() []ent.Field {
	return []ent.Field {
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.Int("state_id").
			Positive(),
		field.Int("once_in_days").
			Optional().
			Nillable(),
	}
}

// Edges of the StateRequirement.
func (StateRequirement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("state", State.Type).
			Ref("requirements").
			Unique().
			Required().
			Field("state_id"),
		edge.To("checks", StateRequirementCheck.Type),
	}
}