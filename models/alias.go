package models

import (
	"arturgudiev/dashboard/ent/schema"
)

type AliasModel struct {
	ID     int                  `json:"id"`
	Type   schema.ContainerType `json:"type"`
	ItemID int                  `json:"itemId"`
	Alias  string               `json:"alias"`
}
