package models



type StateFull struct {
	ID                int     `json:"id"`
	Description       string  `json:"description"`
	Tags              []string `json:"tags"`
	Notes             string  `json:"notes"`
	Closed            bool    `json:"closed"`
	States            []int   `json:"states"`
	StateRequirements []int   `json:"stateRequirements"`
	IsFulfilled       *bool   `json:"isFulfilled"`
}

type StateRequirementFull struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	StateID     int    `json:"state_id"`
	OnceInDays  *int   `json:"once_in_days,omitempty"`
	IsFulfilled *bool  `json:"isFulfilled"`
}

type StateRequirementShort struct {
	Description string `json:"description" example:"Fix login bug"`
	OnceInDays  *int   `json:"onceInDays" example:"7"`
}

type StateRequirementPartial struct {
	ID          int     `json:"id"`
	Description *string `json:"description"`
	StateID     *int    `json:"state_id"`
	OnceInDays  *int    `json:"once_in_days"`
}

// type NewStateRequest struct {
// 	State  StateShort            `json:"state"`
// 	Parent *ContainerDescription `json:"parent,omitempty"`
// }
