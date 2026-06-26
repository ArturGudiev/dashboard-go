package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type State struct {
	ent.Schema
}

func (State) Fields() []ent.Field {
	return []ent.Field{
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
	}
}

func (State) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("requirements", StateRequirement.Type),
	}
}