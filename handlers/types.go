package handlers

import (
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"time"
)

// IDsRequest represents a request with a list of IDs
type IDsRequest struct {
	IDs []int `json:"ids" binding:"required"`
}

// TaskIDRequest represents a task request with just an ID (used in finish-tasks endpoint)
type TaskIDRequest struct {
	ID int `json:"_id"`
}

// ParentObj represents the parent object with an ID
type ParentObj struct {
	ID int `json:"_id"`
}

// ParentRequest represents a parent request with type and object
type ParentRequest struct {
	Type string    `json:"type"`
	Obj  ParentObj `json:"obj"`
}

// TaskRequest represents a full task request (used in new-task endpoint)
type TaskRequest struct {
	Description      string          `json:"description"`
	Tags             []string        `json:"tags,omitempty"`
	Done             bool            `json:"done,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	Problems         []int           `json:"problems,omitempty"`
	Questions        []int           `json:"questions,omitempty"`
	Actions          []int           `json:"actions,omitempty"`
	Definitions      []int           `json:"definitions,omitempty"`
	KnowledgeBits    []int           `json:"knowledgeBits,omitempty"`
	KnowledgeNodes   []int           `json:"knowledgeNodes,omitempty"`
	ParentContainers [][]interface{} `json:"parents,omitempty"`
	DoneDateTime     *time.Time      `json:"doneDateTime,omitempty"`
}

// NewTaskRequest represents a request to create a new task with optional parent
type NewTaskRequest struct {
	Task   models.TaskShort             `json:"task"`
	Parent *models.ContainerDescription `json:"parent,omitempty"`
}

// NewTaskRequest represents a request to create a new task with optional parent
type ChangeTasksOrderRequest struct {
	ContainerType   schema.ContainerType `json:"containerType" binding:"required"`
	ContainerID     int                  `json:"containerID" binding:"required"`
	TasksInNewOrder []int                `json:"tasksInNewOrder" binding:"required"`
}

// PatchTaskByIDRequest represents optional fields for PATCH /task/:id
type PatchTaskByIDRequest struct {
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
}

// PatchContainerByIDRequest is used by PATCH endpoints that partially update a container name/notes.
type PatchContainerByIDRequest = PatchTaskByIDRequest

// UpdateTaskRequest represents a request to update an existing task
type UpdateTaskRequest struct {
	ID               int             `json:"id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags,omitempty"`
	Done             bool            `json:"done,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	Problems         []int           `json:"problems,omitempty"`
	Questions        []int           `json:"questions,omitempty"`
	Actions          []int           `json:"actions,omitempty"`
	Definitions      []int           `json:"definitions,omitempty"`
	KnowledgeBits    []int           `json:"knowledgeBits,omitempty"`
	KnowledgeNodes   []int           `json:"knowledgeNodes,omitempty"`
	ParentContainers [][]interface{} `json:"parents,omitempty"`
	DoneDateTime     *time.Time      `json:"doneDateTime,omitempty"`
}

// TaskResponse represents a task response with all fields always included
type TaskResponse struct {
	ID               int             `json:"id"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags"`
	Done             bool            `json:"done"`
	Notes            string          `json:"notes"`
	Problems         []int           `json:"problems"`
	Questions        []int           `json:"questions"`
	Actions          []int           `json:"actions"`
	Definitions      []int           `json:"definitions"`
	KnowledgeBits    []int           `json:"knowledge_bits"`
	ParentContainers [][]interface{} `json:"parent_containers"`
	KnowledgeNodes   []int           `json:"knowledge_nodes"`
	DoneDateTime     *time.Time      `json:"done_date_time"`
}

type DoneTasksResponse struct {
	DoneTasks int `json:"doneTasks"`
}

// NewProblemRequest represents a request to create a new problem with optional parent
// @Description Request to create a new problem with optional parent container
type NewProblemRequest struct {
	Problem models.ProblemShort          `json:"problem"`
	Parent  *models.ContainerDescription `json:"parent,omitempty"`
}

// SolveProblemRequest represents a request to solve a problem
type SolveProblemRequest struct {
	Solution string `json:"solution" binding:"required"`
}

// NewQuestionRequest represents a request to create a new question with optional parent
// @Description Request to create a new question with optional parent container
type NewQuestionRequest struct {
	Question models.QuestionShort         `json:"question"`
	Parent   *models.ContainerDescription `json:"parent,omitempty"`
}

// AnswerQuestionRequest represents a request to answer a question
type AnswerQuestionRequest struct {
	Answer string `json:"answer" binding:"required"`
}

// ParentsPathRequest represents a request to get parents path
type ParentsPathRequest struct {
	ID   int    `json:"id" binding:"required"`
	Type string `json:"type" binding:"required"`
}

// NewTaskRequest represents a request to create a new task with optional parent
type NewLogMessageRequest struct {
	Description   string                `json:"description"`
	ContainerType *schema.ContainerType `json:"containerType"`
	ContainerID   *int                  `json:"containerID"`
	LogType       *string               `json:"logType"`
}

type logMessagesQuery struct {
	ContainerType *schema.ContainerType `form:"containerType"`
	ContainerID   *int                  `form:"containerID"`
	PerPage       *int                  `form:"perPage"`
	Page          *int                  `form:"page"`
	Global        *bool                 `form:"global"`
}

// type logMessagesResponse struct {
// 	LogMessages []*ent.LogMessage `json:"logMessages"`
// 	Total       int                  `json:"total"`
// 	Page        int                  `json:"page"`
// 	PerPage     int                  `json:"perPage"`
// }

type NewRepetitiveTaskRequest struct {
	RepetitiveTask models.RepetitiveTaskShort   `json:"repetitiveTask"`
	Parent         *models.ContainerDescription `json:"parent,omitempty"`
}

// RepetitiveTasksQuery represents a request to get repetitive tasks
type RepetitiveTasksQuery struct {
	Actual *bool `form:"actual"`
}

type NewLongTaskRequest struct {
	LongTask models.LongTaskShort         `json:"longTask"`
	Parent   *models.ContainerDescription `json:"parent,omitempty"`
}

// LongTasksQuery represents query params for GET /long-tasks
type LongTasksQuery struct {
	Open *bool `form:"open"`
}

type AddLongTaskSubmissionRequest struct {
	Comments      *string  `json:"comments"`
	ProgressToAdd *float64 `json:"progressToAdd"`
	ProgressToSet *float64 `json:"progressToSet"`
}

// AddContainerVariableRequest adds a variable to a container's variables stack.
type AddContainerVariableRequest struct {
	ContainerType schema.ContainerType `json:"containerType" binding:"required"`
	ContainerID   int                  `json:"containerID" binding:"required,gt=0"`
	VariableName  string               `json:"variableName" binding:"required"`
	VariableValue string               `json:"variableValue"`
}

// PatchContainerVariableRequest represents optional fields for PATCH /container-variables/:id
type PatchContainerVariableRequest struct {
	VariableName  *string `json:"variableName"`
	VariableValue *string `json:"variableValue"`
}
