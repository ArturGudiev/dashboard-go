package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/models"
	"context"
	"fmt"
	"time"
)

// TaskService handles task-related business logic
type TaskService struct {
	client *ent.Client
}

// NewTaskService creates a new TaskService
func NewTaskService(client *ent.Client) *TaskService {
	return &TaskService{client: client}
}

// GetOpenDescendantTasks recursively gets all descendant tasks that are not done
func (s *TaskService) GetOpenDescendantTasks(ctx context.Context, parentTask *ent.Task) []*ent.Task {
	var result []*ent.Task

	// Get all child relationships where this task is the parent
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ParentID(parentTask.ID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		WithChild().
		All(ctx)

	if err != nil {
		return result
	}

	// Process each child
	for _, relation := range childRelations {
		childTask := relation.Edges.Child
		if childTask == nil {
			continue
		}

		// Only include tasks that are not done
		if !childTask.Done {
			result = append(result, childTask)
		}

		// Recursively get descendants of this child
		descendants := s.GetOpenDescendantTasks(ctx, childTask)
		result = append(result, descendants...)
	}

	return result
}

func (s *TaskService) GetParent(ctx context.Context, task *ent.Task) *ent.Task {
	// Get all parent relationships where this task is the child
	parentRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ChildID(task.ID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		WithParent().
		All(ctx)

	if err != nil || len(parentRelations) == 0 {
		return nil
	}

	// Get the parent task from the first relation
	parentTask := parentRelations[0].Edges.Parent
	if parentTask != nil {
		return parentTask
	}

	// If parent wasn't loaded via edge, get it by ID
	parentID := parentRelations[0].ParentID
	parentTask, err = s.client.Task.Get(ctx, parentID)
	if err != nil {
		return nil
	}

	return parentTask
}

func (s *TaskService) GetParentsPathDescriptions(ctx context.Context, task *ent.Task) []string {
	parentsPath := s.GetParentsPath(ctx, task)
	var descriptions []string
	if parentsPath == nil {
		return descriptions
	}
	mapper := func(el *ent.Task) string { return el.Description }
	for _, el := range parentsPath {
		descriptions = append(descriptions, mapper(el))
	}
	return descriptions
}

func (s *TaskService) GetParentsPath(ctx context.Context, task *ent.Task) []*ent.Task {
	var items []*ent.Task
	parent := s.GetParent(ctx, task)

	for parent != nil {
		items = append(items, parent)
		parent = s.GetParent(ctx, parent)
	}

	return items
}

// FinishTaskRecursively finishes all open descendant tasks of the given task
func (s *TaskService) FinishTaskRecursively(ctx context.Context, task *ent.Task) error {
	// Recursively get all descendant tasks that are not done
	allTasksToFinish := s.GetOpenDescendantTasks(ctx, task)

	// Mark all tasks as done with current timestamp
	now := time.Now()
	for _, taskToFinish := range allTasksToFinish {
		_, err := s.client.Task.UpdateOneID(taskToFinish.ID).
			SetDone(true).
			SetDoneDateTime(now).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// FinishTaskById finishes a task and all its descendants by task ID
func (s *TaskService) FinishTaskById(ctx context.Context, taskID int) (*ent.Task, error) {
	// Get the task by ID
	task, err := s.client.Task.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Finish task recursively
	if err := s.FinishTaskRecursively(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to finish task recursively: %v", err)
	}

	// Return the updated task
	updatedTask, err := s.client.Task.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated task: %v", err)
	}

	return updatedTask, nil
}

// GetChildSubtasks returns all child tasks for a given parent task ID
func (s *TaskService) GetChildSubtasks(ctx context.Context, parentID int) ([]*ent.Task, error) {
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		Order(containerchild.ByChildOrder()).
		WithChild().
		All(ctx)

	if err != nil {
		return nil, err
	}

	var childTasks []*ent.Task
	for _, relation := range childRelations {
		if relation.Edges.Child != nil {
			childTasks = append(childTasks, relation.Edges.Child)
		}
	}

	return childTasks, nil
}

// AddSubtask creates a new subtask for the given parent task
func (s *TaskService) AddSubtask(ctx context.Context, parentID int, description string) (*ent.Task, error) {
	// Create the new task
	newTask, err := s.client.Task.Create().
		SetDescription(description).
		SetDone(false).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count children: %v", err)
	}

	// Get the count of existing parents to set parent_order
	parentCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
			containerchild.ChildID(newTask.ID),
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count parents: %v", err)
	}

	// Create the parent-child relationship
	_, err = s.client.ContainerChild.Create().
		SetParentType(containerchild.ParentTypeTask).
		SetParentID(parentID).
		SetChildType(containerchild.ChildTypeTask).
		SetChildID(newTask.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %v", err)
	}

	return newTask, nil
}

// GetTaskFull returns a task with all fields plus children tasks at the top level
func (s *TaskService) GetTaskFull(ctx context.Context, taskID int) (*models.TaskFull, error) {
	// Get the task
	task, err := s.client.Task.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get children tasks (tasks where this task is the parent)
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ParentID(taskID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		Order(containerchild.ByChildOrder()).
		WithChild().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %v", err)
	}

	tasks := make([]*ent.Task, 0, len(childRelations))
	for _, relation := range childRelations {
		if relation.Edges.Child != nil {
			tasks = append(tasks, relation.Edges.Child)
		}
	}

	// Build TaskFull from Task
	taskFull := &models.TaskFull{
		ID:               task.ID,
		Description:      task.Description,
		Tags:             task.Tags,
		Done:             task.Done,
		Notes:            task.Notes,
		Problems:         task.Problems,
		Questions:        task.Questions,
		Actions:          task.Actions,
		Definitions:      task.Definitions,
		KnowledgeBits:    task.KnowledgeBits,
		ParentContainers: task.ParentContainers,
		KnowledgeNodes:   task.KnowledgeNodes,
		DoneDateTime:     task.DoneDateTime,
		Tasks:            tasks,
	}

	return taskFull, nil
}
