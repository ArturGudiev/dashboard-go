package models

import "time"

type LongTaskProgressSubmission struct {
	ID            int       `json:"id"`
	Comments      string    `json:"comments"`
	ProgressToAdd *float64  `json:"progressToAdd"`
	ProgressToSet *float64  `json:"progressToSet"`
	ProgressRaw   *string   `json:"progressRaw"`
	LongTaskProgressID int `json:"longTaskProgressID"`
	ExecutionDate time.Time `json:"executionDate"`
}
