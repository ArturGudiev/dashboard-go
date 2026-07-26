package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ContainerCheck holds a check item scoped to a container.
type ContainerCheck struct {
	ent.Schema
}

// Fields of the ContainerCheck.
func (ContainerCheck) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.Enum("container_type").
			GoType(ContainerType("")),
		field.Int("container_id").
			Positive(),
	}
}

// Indexes of the ContainerCheck.
func (ContainerCheck) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("container_type", "container_id"),
	}
}

// Annotations of the ContainerCheck.
func (ContainerCheck) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "container_checks"},
	}
}
