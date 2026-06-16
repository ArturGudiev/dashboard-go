package models

import "time"

type LongTaskFull struct {
	ID          int      `json:"id"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Done        bool     `json:"done"`
	DoneDateTime *time.Time `json:"doneDateTime"`
	Progresses  []LongTaskProgress `json:"progresses"`
	Notes       string   `json:"notes"`
}
