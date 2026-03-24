package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type LogMessage struct {
	ent.Schema
}

// Fields of the Problem.
func (LogMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.String("notes").
			Default(""),
		field.Time("created"),
		field.Enum("container_type").
			GoType(ContainerType("")).
			Optional().
			Nillable(),
		field.Int("container_id").
			Optional().
			Nillable().
			Positive(),
	}
}
