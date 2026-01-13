package models

import "time"

type StoryFull struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	Tags             []string               `json:"tags"`
	Notes            string                 `json:"notes"`
	Closed           bool                   `json:"closed"`
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

type StoryShort struct {
	Description string   `json:"description" example:"Fix login bug"`
	Tags        []string `json:"tags" example:"bug,urgent"`
	Notes       string   `json:"notes" example:"User cannot log in"`
}

type StoryPartial struct {
	ID           int        `json:"id"`
	Description  *string    `json:"description"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	DoneDateTime *time.Time `json:"doneDateTime"`
	Answer       *string    `json:"answer"`
}

type NewStoryRequest struct {
	Story  StoryShort            `json:"story"`
	Parent *ContainerDescription `json:"parent,omitempty"`
}
