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
	KnowledgeBits    []int                  `json:"knowledgeBits"`
	ParentContainers []ContainerDescription `json:"parentContainers"`
	KnowledgeNodes   []int                  `json:"knowledgeNodes"`
	DoneDateTime     *time.Time             `json:"doneDateTime"`
}

// ProblemShort represents a new problem creation request
// @Description New problem creation request
type ProblemShort struct {
	Description string   `json:"description" example:"Fix login bug"`
	Tags        []string `json:"tags" example:"bug,urgent"`
	Notes       string   `json:"notes" example:"User cannot log in"`
}

type ProblemPartial struct {
	ID           int        `json:"id"`
	Description  *string    `json:"description"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	DoneDateTime *time.Time `json:"doneDateTime"`
	Solution     *string    `json:"solution"`
}
