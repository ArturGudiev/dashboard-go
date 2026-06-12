package models

type LongTaskShort struct {
	Description   string   `json:"description" example:"Read Go book"`
	Tags          []string `json:"tags" example:"learning,go"`
	Notes         string   `json:"notes" example:"Chapter 1-5"`
	ProgressTotal *float64 `json:"progressTotal" example:"100"`
	ProgressDone  *float64 `json:"progressDone" example:"0"`
	ProgressUnits *string  `json:"progressUnits" example:"percents"`
}

type LongTaskPartial struct {
	ID          int     `json:"id"`
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
}
