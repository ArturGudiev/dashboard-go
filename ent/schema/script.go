package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Script holds a global or container-scoped executable script.
type Script struct {
	ent.Schema
}

// Fields of the Script.
func (Script) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.Text("code").
			NotEmpty(),
		field.String("description").
			Default(""),
		field.JSON("params", []ScriptParam{}).
			Optional(),
		field.Enum("container_type").
			GoType(ContainerType("")).
			Optional().
			Nillable(),
		field.Int("container_id").
			Positive().
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now().UTC).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now().UTC).
			UpdateDefault(time.Now().UTC),
	}
}

// Indexes of the Script.
func (Script) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("container_type", "container_id"),
	}
}

// Annotations of the Script.
func (Script) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "scripts"},
	}
}

// ScriptParam describes an input parameter declared on a script.
type ScriptParam struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // string | boolean | number
	Default any    `json:"default,omitempty"`
}
