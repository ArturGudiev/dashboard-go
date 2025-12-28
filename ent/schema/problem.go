package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Problem holds the schema definition for the Problem entity.
type Problem struct {
	ent.Schema
}

// Fields of the Problem.
func (Problem) Fields() []ent.Field {
	return []ent.Field{
		// id is automatically the primary key in Ent
		field.Int("id").
			Positive().
			Immutable(),
		field.String("description").
			NotEmpty(),
		field.Strings("tags").
			Optional().
			Default([]string{}),
		field.String("notes").
			Optional().
			Default(""),
		// Arrays of task IDs - stored as JSON (for other types like problems, questions, etc.)
		field.JSON("problems", []int{}).
			Optional().
			Default([]int{}),
		field.JSON("questions", []int{}).
			Optional().
			Default([]int{}),
		field.JSON("actions", []int{}).
			Optional().
			Default([]int{}),
		field.JSON("definitions", []int{}).
			Optional().
			Default([]int{}),
		field.JSON("knowledge_bits", []int{}).
			Optional().
			Default([]int{}),
		// ParentContainers is TaskContainerDescription[] which is [TaskContainerType, number][]
		// Stored as JSON array of arrays (different from the parent-child edge relationship)
		field.JSON("parent_containers", [][]interface{}{}).
			Optional().
			Default([][]interface{}{}),
		field.JSON("knowledge_nodes", []int{}).
			Optional().
			Default([]int{}),
		field.Time("done_date_time").
			Optional().
			Nillable(),
		// Solution field - can be null (string or null)
		field.String("solution").
			Optional().
			Nillable(),
	}
}

// Edges of the Problem.
func (Problem) Edges() []ent.Edge {
	// Note: Parent-child relationships are handled through ContainerChild join table
	// using parent_type/child_type enums, not through Ent edges
	return []ent.Edge{}
}
