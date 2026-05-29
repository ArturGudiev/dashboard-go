package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ContainerVariables holds a single variable entry within a VariablesStack.
type ContainerVariables struct {
	ent.Schema
}

// Fields of the ContainerVariables.
func (ContainerVariables) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Int("variables_stack_id").
			Positive(),
		field.String("variable_name").
			NotEmpty(),
		field.String("variable_value").
			Default(""),
	}
}

// Edges of the ContainerVariables.
func (ContainerVariables) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("variables_stack", VariablesStack.Type).
			Ref("container_variables").
			Unique().
			Required().
			Field("variables_stack_id"),
	}
}

// Indexes of the ContainerVariables.
func (ContainerVariables) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("variables_stack_id", "variable_name").
			Unique(),
		index.Fields("variables_stack_id"),
	}
}

// Annotations of the ContainerVariables.
func (ContainerVariables) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "container_variables"},
	}
}
