package models

import "time"

type LongTaskFull struct {
	ID           int                `json:"id" binding:"required"`
	Description  string             `json:"description" binding:"required"`
	Tags         []string           `json:"tags" binding:"required"`
	Done         bool               `json:"done" binding:"required"`
	DoneDateTime *time.Time         `json:"doneDateTime"`
	Progresses   []LongTaskProgress `json:"progresses" binding:"required"`
	Notes        string             `json:"notes" binding:"required"`
}
