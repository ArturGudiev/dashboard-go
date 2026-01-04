package models

import "arturgudiev/dashboard/ent/schema"

type ContainerDescription struct {
	ID            int                  `json:"id"`
	ContainerType schema.ContainerType `json:"containerType"`
}
