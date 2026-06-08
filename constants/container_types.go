package constants

import (
	"arturgudiev/dashboard/ent/schema"
)

var CapitalisedContainerTypes = map[schema.ContainerType]string{
	schema.ContainerTypeEpic:          "Epic",
	schema.ContainerTypeStory:         "Story",
	schema.ContainerTypeTask:          "Task",
	schema.ContainerTypeProblem:       "Problem",
	schema.ContainerTypeQuestion:      "Question",
	schema.ContainerTypeKnowledgeBit:  "KnowledgeBit",
	schema.ContainerTypeKnowledgeNode: "KnowledgeNode",
	schema.ContainerTypeLongTask:      "LongTask",
}
