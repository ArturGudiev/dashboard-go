package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VariablesStack holds variable stacks scoped to a container.
type VariablesStack struct {
	ent.Schema
}

// Fields of the VariablesStack.
func (VariablesStack) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Enum("container_type").
			GoType(ContainerType("")),
		field.Int("container_id").
			Positive(),
	}
}

// Edges of the VariablesStack.
func (VariablesStack) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("container_variables", ContainerVariables.Type),
	}
}

// Indexes of the VariablesStack.
func (VariablesStack) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("container_type", "container_id").
			Unique(),
	}
}

// Annotations of the VariablesStack.
func (VariablesStack) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "variables_stacks"},
	}
}
