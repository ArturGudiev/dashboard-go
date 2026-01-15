package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Epic struct {
	ent.Schema
}

func (Epic) Fields() []ent.Field {
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
