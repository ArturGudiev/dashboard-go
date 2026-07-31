package models

import "time"

type EpicFull struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	Tags             []string               `json:"tags"`
	Notes            string                 `json:"notes"`
	Closed           bool                   `json:"closed"`
	Epics            []int                  `json:"epics"`
	Stories          []int                  `json:"stories"`
	Tasks            []int                  `json:"tasks"`
	Questions        []int                  `json:"questions"`
	Problems         []int                  `json:"problems"`
	LongTasks        []int                  `json:"longTasks"`
	Actions          []int                  `json:"actions"`
	Definitions      []int                  `json:"definitions"`
	KnowledgeBits    []int                  `json:"knowledgeBits"`
	ParentContainers []ContainerDescription `json:"parentContainers"`
	KnowledgeNodes   []int                  `json:"knowledgeNodes"`
	DoneDateTime     *time.Time             `json:"doneDateTime"`
}

type EpicShort struct {
	Description string   `json:"description" example:"Fix login bug"`
	Tags        []string `json:"tags" example:"bug,urgent"`
	Notes       string   `json:"notes" example:"User cannot log in"`
}

type EpicPartial struct {
	ID           int        `json:"id"`
	Description  *string    `json:"description"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	Closed       *bool      `json:"closed"`
	DoneDateTime *time.Time `json:"doneDateTime"`
}

type NewEpicRequest struct {
	Epic   EpicShort             `json:"epic"`
	Parent *ContainerDescription `json:"parent,omitempty"`
}
