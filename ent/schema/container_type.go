package schema

// ContainerType defines the type for container entities (task, problem, etc.)
type ContainerType string

// ContainerType values.
const (
	ContainerTypeEpic          ContainerType = "epic"
	ContainerTypeStory         ContainerType = "story"
	ContainerTypeTask          ContainerType = "task"
	ContainerTypeQuestion      ContainerType = "question"
	ContainerTypeProblem       ContainerType = "problem"
	ContainerTypeKnowledgeNode ContainerType = "knowledge-node"
	ContainerTypeKnowledgeBit  ContainerType = "knowledge-bit"
	ContainerTypeDefinition    ContainerType = "definition"
	ContainerTypeAction        ContainerType = "action"
	ContainerTypeScheduledTask ContainerType = "scheduled-task"
	ContainerTypeState         ContainerType = "state"
)

// Values returns all valid ContainerType values.
func (ContainerType) Values() []string {
	return []string{
		"epic",
		"story",
		"task",
		"question",
		"problem",
		"knowledge-node",
		"knowledge-bit",
		"definition",
		"action",
		"scheduled-task",
		"state",
	}
}


