package models

import (
	"arturgudiev/dashboard/ent/schema"
)

type AliasModel struct {
	ID       int                   `json:"id" binding:"required"`
	Type     schema.AliasType      `json:"type" binding:"required"`
	ItemID   *int                  `json:"itemId,omitempty"`
	FilePath *string               `json:"filePath,omitempty"`
	Alias    string                `json:"alias" binding:"required"`
}
