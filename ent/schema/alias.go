package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Alias struct {
	ent.Schema
}

func (Alias) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.Enum("type").
			GoType(AliasType("")),
		field.Int("item_id").
			Positive().
			Optional().
			Nillable(),
		field.String("file_path").
			Optional().
			Nillable(),
		field.String("alias").
			NotEmpty().
			Unique(),
	}
}

// Annotations of the Alias.
func (Alias) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Explicitly set table name to aliases
		entsql.Annotation{
			Table: "aliases",
		},
	}
}
