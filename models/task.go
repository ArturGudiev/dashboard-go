package models

import (
	"arturgudiev/dashboard/ent"
	"time"
)

// TaskFull represents a task with all fields plus children tasks at the top level
type TaskFull struct {
	ID               int             `json:"id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags,omitempty"`
	Done             bool            `json:"done"`
	Notes            string          `json:"notes,omitempty"`
	Problems         []int           `json:"problems,omitempty"`
	Questions        []int           `json:"questions,omitempty"`
	Actions          []int           `json:"actions,omitempty"`
	Definitions      []int           `json:"definitions,omitempty"`
	KnowledgeBits    []int           `json:"knowledge_bits,omitempty"`
	ParentContainers [][]interface{} `json:"parent_containers,omitempty"`
	KnowledgeNodes   []int           `json:"knowledge_nodes,omitempty"`
	DoneDateTime     *time.Time      `json:"done_date_time,omitempty"`
	Tasks            []*ent.Task     `json:"tasks,omitempty"`
}
