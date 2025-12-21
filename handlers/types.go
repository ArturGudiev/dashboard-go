package handlers

import "time"

// IDsRequest represents a request with a list of IDs
type IDsRequest struct {
	IDs []int `json:"ids" binding:"required"`
}

// TaskIDRequest represents a task request with just an ID (used in finish-tasks endpoint)
type TaskIDRequest struct {
	ID int `json:"_id"`
}

// ParentObj represents the parent object with an ID
type ParentObj struct {
	ID int `json:"_id"`
}

// ParentRequest represents a parent request with type and object
type ParentRequest struct {
	Type string    `json:"type"`
	Obj  ParentObj `json:"obj"`
}

// TaskRequest represents a full task request (used in new-task endpoint)
type TaskRequest struct {
	Description      string          `json:"description"`
	Tags             []string        `json:"tags,omitempty"`
	Done             bool            `json:"done,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	Problems         []int           `json:"problems,omitempty"`
	Questions        []int           `json:"questions,omitempty"`
	Actions          []int           `json:"actions,omitempty"`
	Definitions      []int           `json:"definitions,omitempty"`
	KnowledgeBits    []int           `json:"knowledgeBits,omitempty"`
	KnowledgeNodes   []int           `json:"knowledgeNodes,omitempty"`
	ParentContainers [][]interface{} `json:"parents,omitempty"`
	DoneDateTime     *time.Time      `json:"doneDateTime,omitempty"`
}

// NewTaskRequest represents a request to create a new task with optional parent
type NewTaskRequest struct {
	Task   TaskRequest    `json:"task"`
	Parent *ParentRequest `json:"parent,omitempty"`
}

// UpdateTaskRequest represents a request to update an existing task
type UpdateTaskRequest struct {
	ID               int             `json:"_id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags,omitempty"`
	Done             bool            `json:"done,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	Problems         []int           `json:"problems,omitempty"`
	Questions        []int           `json:"questions,omitempty"`
	Actions          []int           `json:"actions,omitempty"`
	Definitions      []int           `json:"definitions,omitempty"`
	KnowledgeBits    []int           `json:"knowledgeBits,omitempty"`
	KnowledgeNodes   []int           `json:"knowledgeNodes,omitempty"`
	ParentContainers [][]interface{} `json:"parents,omitempty"`
	DoneDateTime     *time.Time      `json:"doneDateTime,omitempty"`
}
