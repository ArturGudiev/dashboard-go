package models

type RepetitiveTaskShort struct {
	Description  string   `json:"description" example:"Fix login bug"`
	Tags         []string `json:"tags" example:"bug,urgent"`
	Notes        string   `json:"notes" example:"User cannot log in"`
	OnceInDays   *int     `json:"onceInDays" example:"7"`
	OnceInWeeks  *int     `json:"onceInWeeks" example:"2"`
	OnceInMonths *int     `json:"onceInMonths" example:"3"`
}

// RepetitiveTaskResponse documents repetitive task response payload in Swagger.
// ID intentionally does not use omitempty, so OpenAPI marks it as required.
type RepetitiveTaskPartial struct {
	ID          int     `json:"id"`
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
}

type RepetitiveTaskResponse struct {
	ID           int      `json:"id" binding:"required" example:"1"`
	Description  string   `json:"description,omitempty" example:"Fix login bug"`
	Tags         []string `json:"tags,omitempty" example:"bug,urgent"`
	Closed       bool     `json:"closed,omitempty"`
	Notes        string   `json:"notes,omitempty" example:"User cannot log in"`
	OnceInDays   *int     `json:"once_in_days,omitempty" example:"7"`
	OnceInWeeks  *int     `json:"once_in_weeks,omitempty" example:"2"`
	OnceInMonths *int     `json:"once_in_months,omitempty" example:"3"`
}

