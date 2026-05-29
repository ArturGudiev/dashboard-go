package models

// ContainerVariable is a name/value pair stored on a container's variables stack.
type ContainerVariable struct {
	ID            int    `json:"id"`
	VariableName  string `json:"variableName"`
	VariableValue string `json:"variableValue"`
}
