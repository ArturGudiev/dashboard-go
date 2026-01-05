package models

import (
	"time"
)

// ProblemFull represents a complete problem with all related data
// @Description Complete problem information with all related entities
type ProblemFull struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	Tags             []string               `json:"tags"`
	Notes            string                 `json:"notes"`
	Solution         *string                `json:"solution"`
	Tasks            []int                  `json:"tasks"`
	Problems         []int                  `json:"problems"`
	Questions        []int                  `json:"questions"`
	Actions          []int                  `json:"actions"`
	Definitions      []int                  `json:"definitions"`
	KnowledgeBits    []int                  `json:"knowledge_bits"`
	ParentContainers []ContainerDescription `json:"parent_containers"`
	KnowledgeNodes   []int                  `json:"knowledge_nodes"`
	DoneDateTime     *time.Time             `json:"done_date_time"`
}

// NewProblem represents a new problem creation request
// @Description New problem creation request
type NewProblem struct {
	Description string   `json:"description" example:"Fix login bug"`
	Tags        []string `json:"tags,omitempty" example:"bug,urgent"`
	Notes       string   `json:"notes,omitempty" example:"User cannot log in"`
}
