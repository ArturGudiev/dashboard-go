package models

import "arturgudiev/dashboard/ent/schema"

// ContainerDescription represents a container with ID and type
// @Description Container description with ID and type
type ContainerDescription struct {
	ID   int                  `json:"id" example:"1"`
	Type schema.ContainerType `json:"type" example:"task"`
}
