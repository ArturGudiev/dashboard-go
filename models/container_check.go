package models

import "arturgudiev/dashboard/ent/schema"

// ContainerCheck is a check item scoped to a container.
type ContainerCheck struct {
	ID            int                  `json:"id"`
	Description   string               `json:"description"`
	ContainerType schema.ContainerType `json:"containerType"`
	ContainerID   int                  `json:"containerID"`
}
