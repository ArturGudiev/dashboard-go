package models

import "time"

// QuestionFull represents a complete question with all related data
// @Description Complete question information with all related entities
type QuestionFull struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	Tags             []string               `json:"tags"`
	Notes            string                 `json:"notes"`
	Answer           *string                `json:"answer"`
	Tasks            []int                  `json:"tasks"`
	Questions        []int                  `json:"questions"`
	Problems         []int                  `json:"problems"`
	Actions          []int                  `json:"actions"`
	Definitions      []int                  `json:"definitions"`
	KnowledgeBits    []int                  `json:"knowledgeBits"`
	ParentContainers []ContainerDescription `json:"parentContainers"`
	KnowledgeNodes   []int                  `json:"knowledgeNodes"`
	DoneDateTime     *time.Time             `json:"doneDateTime"`
}

// QuestionShort represents a new question creation request
// @Description New question creation request
type QuestionShort struct {
	Description string   `json:"description" example:"Fix login bug"`
	Tags        []string `json:"tags" example:"bug,urgent"`
	Notes       string   `json:"notes" example:"User cannot log in"`
}

type QuestionPartial struct {
	ID           int        `json:"id"`
	Description  *string    `json:"description"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	DoneDateTime *time.Time `json:"doneDateTime"`
	Answer       *string    `json:"answer"`
}
