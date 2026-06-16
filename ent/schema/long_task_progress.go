package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// LongTaskSubmission holds the schema definition for the LongTaskSubmission entity.
type LongTaskProgress struct {
	ent.Schema
}

// Fields of the LongTaskProgress.
func (LongTaskProgress) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.Int("long_task_id").
			Positive(),
		field.Float("value").
			Optional().
			Nillable(),
		field.Float("total").
			Optional().
			Nillable(),
		field.String("units").
			Optional().
			Nillable(),
	}
}

// Edges of the LongTaskProgress.
func (LongTaskProgress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("long_task", LongTask.Type).
			Ref("progresses").
			Unique().
			Required().
			Field("long_task_id"),
		edge.To("progress_submissions", LongTaskProgressSubmission.Type),
	}
}

