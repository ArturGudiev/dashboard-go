package models

import (
	"time"
)

// TaskFull represents a task with all fields plus children tasks at the top level
type TaskFull struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	Tags             []string               `json:"tags"`
	Done             bool                   `json:"done"`
	Notes            string                 `json:"notes"`
	Tasks            []int                  `json:"tasks"`
	Problems         []int                  `json:"problems"`
	Questions        []int                  `json:"questions"`
	Actions          []int                  `json:"actions"`
	Definitions      []int                  `json:"definitions"`
	KnowledgeBits    []int                  `json:"knowledgeBits"`
	ParentContainers []ContainerDescription `json:"parentContainers"`
	KnowledgeNodes   []int                  `json:"knowledgeNodes"`
	Variables        []ContainerVariable    `json:"variables"`
	DoneDateTime     *time.Time             `json:"doneDateTime"`
}

type TaskPartial struct {
	ID           int        `json:"id"`
	Description  *string    `json:"description"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	DoneDateTime *time.Time `json:"doneDateTime"`
	Done         *bool      `json:"done"`
}

type TaskFieldsPartial struct {
	Description  *string    `json:"description"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	DoneDateTime *time.Time `json:"doneDateTime"`
	Done         *bool      `json:"done"`
}

type TaskShort struct {
	Description string   `json:"description" example:"Fix login bug"`
	Tags        []string `json:"tags" example:"bug,urgent"`
	Notes       string   `json:"notes" example:"User cannot log in"`
}

