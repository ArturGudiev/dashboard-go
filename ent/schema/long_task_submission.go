package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// LongTaskSubmission holds the schema definition for the LongTaskSubmission entity.
type LongTaskSubmission struct {
	ent.Schema
}

// Fields of the LongTaskSubmission.
func (LongTaskSubmission) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("long_task_id").
			Positive(),
		field.Time("execution_date"),
		field.String("comments").
			Optional().
			Nillable(),
		field.Float("progress_to_add").
			Optional().
			Nillable(),
		field.Float("progress_to_set").
			Optional().
			Nillable(),
	}
}

// Edges of the LongTaskSubmission.
func (LongTaskSubmission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("long_task", LongTask.Type).
			Ref("submissions").
			Unique().
			Required().
			Field("long_task_id"),
	}
}
