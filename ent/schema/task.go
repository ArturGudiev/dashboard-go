package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
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
		field.Bool("done").
			Default(false),
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
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		// Parent-child relationships through ContainerChild join table
		// A task can be a parent to multiple children
		edge.From("children_relations", ContainerChild.Type).
			Ref("parent"),
		// A task can be a child to multiple parents
		edge.From("parents_relations", ContainerChild.Type).
			Ref("child"),
	}
}

