package models

type DirectionFull struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	Tags             []string               `json:"tags"`
	Notes            string                 `json:"notes"`
	Closed           bool                   `json:"closed"`
	Tasks            []int                  `json:"tasks"`
	Questions        []int                  `json:"questions"`
	Problems         []int                  `json:"problems"`
	Stories          []int                  `json:"stories"`
	Directions       []int                  `json:"directions"`
	LongTasks        []int                  `json:"longTasks"`
	ParentContainers []ContainerDescription `json:"parentContainers"`
}

type DirectionShort struct {
	Description string   `json:"description" example:"Improve backend"`
	Tags        []string `json:"tags" example:"backend,infra"`
	Notes       string   `json:"notes" example:"Q2 goals"`
}

type DirectionPartial struct {
	ID          int     `json:"id"`
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
	Closed      *bool   `json:"closed"`
}

type DirectionStatsEntry struct {
	Date string `json:"date"`
	Text string `json:"text"`
}
