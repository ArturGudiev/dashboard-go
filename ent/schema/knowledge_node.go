package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type KnowledgeNode struct {
	ent.Schema
}

func (KnowledgeNode) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.Strings("tags").
			Default([]string{}),
		field.String("notes").
			Default(""),
	}
}
