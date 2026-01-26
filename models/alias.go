package models

import (
	"arturgudiev/dashboard/ent/schema"
)

type AliasModel struct {
	ID       int                   `json:"id"`
	Type     schema.AliasType      `json:"type"`
	ItemID   *int                  `json:"itemId,omitempty"`
	FilePath *string               `json:"filePath,omitempty"`
	Alias    string                `json:"alias"`
}
