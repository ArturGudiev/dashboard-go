package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"fmt"
	"time"
)

// TaskService handles task-related business logic
type TaskService struct {
	client           *ent.Client
	containerService *ContainerService
	problemService   *ProblemService
}

// NewTaskService creates a new TaskService
func NewTaskService(client *ent.Client, containerService *ContainerService, problemService *ProblemService) *TaskService {
	return &TaskService{client: client, containerService: containerService, problemService: problemService}
}

// GetOpenDescendantTasks recursively gets all descendant tasks that are not done
func (s *TaskService) GetOpenDescendantTasks(ctx context.Context, parentTask *ent.Task) []*ent.Task {
	var result []*ent.Task

	// Get all child relationships where this task is the parent
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(schema.ContainerTypeTask),
			containerchild.ParentID(parentTask.ID),
			containerchild.ChildTypeEQ(schema.ContainerTypeTask),
		).
		All(ctx)

	if err != nil {
		return result
	}

	// Process each child - load tasks manually since edges don't exist
	for _, relation := range childRelations {
		childTask, err := s.client.Task.Get(ctx, relation.ChildID)
		if err != nil {
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

// FinishTaskRecursively finishes all open descendant tasks of the given task
func (s *TaskService) FinishTaskRecursively(ctx context.Context, task *ent.Task) error {
	// Recursively get all descendant tasks that are not done
	allTasksToFinish := s.GetOpenDescendantTasks(ctx, task)

	// Mark all tasks as done with current timestamp
	now := time.Now()
	for _, taskToFinish := range allTasksToFinish {
		_ = s.FinishTaskRecursively(ctx, taskToFinish)
	}

	s.client.Task.UpdateOneID(task.ID).
		SetDone(true).
		SetDoneDateTime(now).
		Save(ctx)

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
