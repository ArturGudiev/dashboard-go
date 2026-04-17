package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// RepetitiveTask holds the schema definition for the RepetitiveTask entity.
type RepetitiveTask struct {
	ent.Schema
}

// Fields of the RepetitiveTask.
func (RepetitiveTask) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.Strings("tags").
			Default([]string{}),
		field.Bool("closed").
			Default(false),
		field.String("notes").
			Default(""),
		field.Int("once_in_days").
			Optional().
			Nillable(),
		field.Int("once_in_weeks").
			Optional().
			Nillable(),
		field.Int("once_in_months").
			Optional().
			Nillable(),
	}
}

// Edges of the RepetitiveTask.
func (RepetitiveTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("executions", RepetitiveTaskExecution.Type),
	}
}
