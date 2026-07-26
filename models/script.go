package models

import (
	"arturgudiev/dashboard/ent/schema"
	"time"
)

// ScriptParam describes an input parameter declared on a script.
type ScriptParam struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // string | boolean | number
	Default any    `json:"default,omitempty"`
}

// ScriptShort is used when creating a script.
// When Container is nil/omitted the script is global; otherwise it is local to that container.
type ScriptShort struct {
	Name        string                `json:"name"`
	Code        string                `json:"code"`
	Description string                `json:"description"`
	Params      []ScriptParam         `json:"params"`
	Container   *ContainerDescription `json:"container,omitempty"`
}

// ScriptPartial is used when updating a script.
type ScriptPartial struct {
	Name        *string        `json:"name"`
	Code        *string        `json:"code"`
	Description *string        `json:"description"`
	Params      *[]ScriptParam `json:"params"`
}

// ScriptListFilter controls which scripts are returned.
// Scope: "all" | "global" | "local" (default "all" when container is provided, else "global").
type ScriptListFilter struct {
	Query         string
	Scope         string
	ContainerType schema.ContainerType
	ContainerID   int
}

// ScriptListItem is a lightweight script for list/autocomplete.
type ScriptListItem struct {
	ID            int                   `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	IsGlobal      bool                  `json:"isGlobal"`
	ContainerType *schema.ContainerType `json:"containerType,omitempty"`
	ContainerID   *int                  `json:"containerID,omitempty"`
}

// ScriptFull is the full script representation.
type ScriptFull struct {
	ID            int                   `json:"id"`
	Name          string                `json:"name"`
	Code          string                `json:"code"`
	Description   string                `json:"description"`
	Params        []ScriptParam         `json:"params"`
	IsGlobal      bool                  `json:"isGlobal"`
	ContainerType *schema.ContainerType `json:"containerType,omitempty"`
	ContainerID   *int                  `json:"containerID,omitempty"`
	CreatedAt     time.Time             `json:"createdAt"`
	UpdatedAt     time.Time             `json:"updatedAt"`
}

// ScriptValidateRequest validates script source without executing host APIs.
type ScriptValidateRequest struct {
	Code string `json:"code"`
}

// ScriptValidateResponse is the result of syntax validation.
type ScriptValidateResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// ScriptRunRequest runs a script against a container.
type ScriptRunRequest struct {
	Container ContainerDescription `json:"container"`
	Params    map[string]any       `json:"params"`
}

// ScriptRunCreated lists entities created during a run.
type ScriptRunCreated struct {
	Tasks    []int `json:"tasks"`
	Problems []int `json:"problems"`
}

// ScriptRunResponse is the result of executing a script.
type ScriptRunResponse struct {
	OK      bool             `json:"ok"`
	Created ScriptRunCreated `json:"created"`
	Error   string           `json:"error,omitempty"`
}

// ToSchemaParams converts model params to schema params for Ent storage.
func ToSchemaParams(params []ScriptParam) []schema.ScriptParam {
	if params == nil {
		return []schema.ScriptParam{}
	}
	out := make([]schema.ScriptParam, len(params))
	for i, p := range params {
		out[i] = schema.ScriptParam{
			Name:    p.Name,
			Type:    p.Type,
			Default: p.Default,
		}
	}
	return out
}

// FromSchemaParams converts Ent schema params to model params.
func FromSchemaParams(params []schema.ScriptParam) []ScriptParam {
	if params == nil {
		return []ScriptParam{}
	}
	out := make([]ScriptParam, len(params))
	for i, p := range params {
		out[i] = ScriptParam{
			Name:    p.Name,
			Type:    p.Type,
			Default: p.Default,
		}
	}
	return out
}
