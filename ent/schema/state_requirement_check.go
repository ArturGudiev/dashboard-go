package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type StateRequirementCheck struct {
	ent.Schema
}

func (StateRequirementCheck) Fields() []ent.Field {
	return []ent.Field {
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.Time("date_time").
			Default(time.Now().UTC()),
		field.Bool("is_fulfilled").
			Default(false),
		field.Int("state_requirement_id").
			Positive(),
	}
}

// Edges of the StateRequirement.
func (StateRequirementCheck) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("state_requirement", StateRequirement.Type).
			Ref("checks").
			Unique().
			Required().
			Field("state_requirement_id"),
	}
}