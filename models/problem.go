package models

import (
	"time"
)

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
