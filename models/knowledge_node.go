package models

import "time"

type KnowledgeNodeFull struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	Tags             []string               `json:"tags"`
	Notes            string                 `json:"notes"`
	Closed           bool                   `json:"closed"`
	Epics            []int                  `json:"epics"`
	Stories          []int                  `json:"stories"`
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

type KnowledgeNodeShort struct {
	Name  string   `json:"name" example:"Fix login bug"`
	Tags  []string `json:"tags" example:"bug,urgent"`
	Notes string   `json:"notes" example:"User cannot log in"`
}

type KnowledgeNodePartial struct {
	ID           int        `json:"id"`
	Name         *string    `json:"name"`
	Tags         *[]string  `json:"tags"`
	Notes        *string    `json:"notes"`
	DoneDateTime *time.Time `json:"doneDateTime"`
}

type NewKnowledgeNodeRequest struct {
	Epic   EpicShort             `json:"epic"`
	Parent *ContainerDescription `json:"parent,omitempty"`
}
