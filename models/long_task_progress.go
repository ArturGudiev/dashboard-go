package models

import "arturgudiev/dashboard/ent"

type LongTaskProgress struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Value       *float64 `json:"value"`
	Total       *float64 `json:"total"`
	Units       *string  `json:"units"`
}

type LongTaskProgressFull struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Value       *float64 `json:"value"`
	Total       *float64 `json:"total"`
	Units       *string  `json:"units"`
	Submissions []*ent.LongTaskProgressSubmission `json:"submissions"`
}

