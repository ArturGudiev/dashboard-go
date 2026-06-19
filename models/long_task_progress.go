package models

type LongTaskProgress struct {
	ID    int      `json:"id" binding:"required"`
	Name  string   `json:"name" binding:"required"`
	Value *float64 `json:"value"`
	Total *float64 `json:"total"`
	Units *string  `json:"units"`
}

type LongTaskProgressFull struct {
	ID          int                           `json:"id" binding:"required"`
	Name        string                        `json:"name" binding:"required"`
	Value       *float64                      `json:"value"`
	Total       *float64                      `json:"total"`
	Units       *string                       `json:"units"`
	Submissions []*LongTaskProgressSubmission `json:"submissions" binding:"required"`
}
