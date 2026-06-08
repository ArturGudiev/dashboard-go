package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// RepetitiveTask holds the schema definition for the RepetitiveTask entity.
type LongTask struct {
	ent.Schema
}

// Fields of the RepetitiveTask.
func (LongTask) Fields() []ent.Field {
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
			Default(false).
			StructTag(`json:"done"`),
		field.String("notes").
			Default(""),
		field.Time("done_date_time").
			Optional().
			Nillable(),
		field.Float("progress_total").
			Default(0).
			StructTag(`json:"progress_total"`),
		field.Float("progress_done").
			Default(0).
			StructTag(`json:"progress_done"`),
		field.String("progress_units").
			Default("percents"),
	}
}


// Edges of the LongTask.
func (LongTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("submissions", LongTaskSubmission.Type),
	}
}
