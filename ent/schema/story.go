package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Story struct {
	ent.Schema
}

func (Story) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
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
		field.Time("done_date_time").
			Optional().
			Nillable(),
	}
}
