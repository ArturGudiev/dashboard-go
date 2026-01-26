package schema

// AliasType defines the type for alias entities (includes ContainerType values + 'file')
type AliasType string

// AliasType values.
const (
	AliasTypeEpic          AliasType = "epic"
	AliasTypeStory         AliasType = "story"
	AliasTypeTask          AliasType = "task"
	AliasTypeQuestion      AliasType = "question"
	AliasTypeProblem       AliasType = "problem"
	AliasTypeKnowledgeNode AliasType = "knowledge-node"
	AliasTypeKnowledgeBit  AliasType = "knowledge-bit"
	AliasTypeDefinition    AliasType = "definition"
	AliasTypeAction        AliasType = "action"
	AliasTypeScheduledTask AliasType = "scheduled-task"
	AliasTypeState         AliasType = "state"
	AliasTypeFile          AliasType = "file"
)

// Values returns all valid AliasType values.
func (AliasType) Values() []string {
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
		"file",
	}
}

// ToContainerType converts AliasType to ContainerType if it's a valid container type.
// Returns false if the type is "file" or invalid.
func (at AliasType) ToContainerType() (ContainerType, bool) {
	if at == AliasTypeFile {
		return "", false
	}
	return ContainerType(at), true
}
