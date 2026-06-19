package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// LongTaskSubmission holds the schema definition for the LongTaskSubmission entity.
type LongTaskProgressSubmission struct {
	ent.Schema
}

// Fields of the LongTaskProgress.
func (LongTaskProgressSubmission) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("comments").
			NotEmpty(),
		field.Float("progress_to_add").
			Optional().
			Nillable(),
		field.Float("progress_to_set").
			Optional().
			Nillable(),
		field.String("progress_raw").
			Optional().
			Nillable(),
		field.Time("execution_date").
			Default(time.Now().UTC()),
		field.Int("long_task_progress_id").
			Positive(),
	}
}

// Edges of the LongTaskProgress.
func (LongTaskProgressSubmission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("long_task_progress", LongTaskProgress.Type).
			Ref("progress_submissions").
			Unique().
			Required().
			Field("long_task_progress_id"),
	}
}
