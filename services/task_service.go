package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// TaskService handles task-related business logic
type TaskService struct {
	client                   *ent.Client
	containerService         *ContainerService
	problemService           *ProblemService
	tasksRepository          *TasksRepository
	childContainerRepository *ChildContainerRepository
}

// NewTaskService creates a new TaskService
func NewTaskService(client *ent.Client, containerService *ContainerService, problemService *ProblemService,
	tasksRepository *TasksRepository, childContainerRepository *ChildContainerRepository) *TaskService {
	return &TaskService{client: client, containerService: containerService, problemService: problemService,
		tasksRepository: tasksRepository, childContainerRepository: childContainerRepository}
}

// GetOpenDescendantTasks recursively gets all descendant tasks that are not done
func (s *TaskService) GetOpenDescendantTasks(ctx context.Context, parentType schema.ContainerType, parentID int) []*ent.Task {
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, parentType, parentID, schema.ContainerTypeTask)
	if err != nil {
		return []*ent.Task{}
	}

	var result []*ent.Task
	for _, relation := range childRelations {
		childTask, err := s.tasksRepository.GetTask(ctx, relation.ChildID)
		if err != nil {
			continue
		}

		// Only include tasks that are not done
		if !childTask.Done {
			result = append(result, childTask)
		}

		// Recursively get descendants of this child
		descendants := s.GetOpenDescendantTasks(ctx, schema.ContainerTypeTask, childTask.ID)
		result = append(result, descendants...)
	}

	return result
}

func (s *TaskService) FinishTaskRecursively(ctx context.Context, task *ent.Task) error {
	allTasksToFinish := s.GetOpenDescendantTasks(ctx, schema.ContainerTypeTask, task.ID)

	now := time.Now()
	for _, taskToFinish := range allTasksToFinish {
		_ = s.FinishTaskRecursively(ctx, taskToFinish)
	}

	_, err := s.client.Task.UpdateOneID(task.ID).
		SetDone(true).
		SetDoneDateTime(now).
		Save(ctx)

	return err
}

func (s *TaskService) FinishTaskById(ctx context.Context, taskID int) (*ent.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.FinishTaskRecursively(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to finish task recursively: %v", err)
	}

	updatedTask, err := s.client.Task.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated task: %v", err)
	}

	return updatedTask, nil
}

func (s *TaskService) GetTaskFull(ctx context.Context, ID int) (*models.TaskFull, error) {
	task, errTask := s.tasksRepository.GetTask(ctx, ID)
	if errTask != nil {
		return nil, errTask
	}
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeProblem, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeProblem, ID)
	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeProblem, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil {
		return nil, errors.New("problem not found")
	}
	TaskFull := &models.TaskFull{
		ID:               ID,
		Description:      task.Description,
		Tags:             task.Tags,
		Notes:            task.Notes,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        []int{},
		Actions:          []int{},
		Definitions:      []int{},
		KnowledgeBits:    []int{},
		ParentContainers: parentContainers,
		KnowledgeNodes:   []int{},
		DoneDateTime:     task.DoneDateTime,
	}

	return TaskFull, nil
}

func (s *TaskService) GetTasksFull(ctx context.Context, IDs []int) ([]*models.TaskFull, error) {
	if len(IDs) == 0 {
		return []*models.TaskFull{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]*models.TaskFull, 0, len(IDs))
	var firstErr error

	for _, id := range IDs {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			taskFull, err := s.GetTaskFull(ctx, taskID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			results = append(results, taskFull)
		}(id)
	}

	wg.Wait()

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *TaskService) AddAnonymousTask(ctx context.Context) (*ent.Task, error) {
	description := "Anonymous task"
	tags := []string{}
	notes := ""
	doneDateTime := time.Now()
	done := true

	fields := models.TaskFieldsPartial{
		Description:  &description,
		Tags:         &tags,
		Notes:        &notes,
		DoneDateTime: &doneDateTime,
		Done:         &done,
	}

	newTask, err := s.tasksRepository.AddTaskByFields(ctx, fields)
	return newTask, err
}

func (s *TaskService) AddTask(ctx context.Context, task models.TaskShort, parent *models.ContainerDescription) (*models.TaskFull, error) {
	newTask, err := s.tasksRepository.AddTask(ctx, task.Description, task.Tags, task.Notes, false)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeTask, newTask.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetTaskFull(ctx, newTask.ID)
}

func (s *TaskService) GetDoneTasksCount(ctx context.Context) (int, error) {
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfTomorrow := startOfToday.AddDate(0, 0, 1)
	return s.tasksRepository.getDoneTasksCountInRange(ctx, startOfToday, startOfTomorrow)
}

func (s *TaskService) UpdateTask(ctx context.Context, taskPartial models.TaskPartial) (*models.TaskFull, error) {
	err := s.tasksRepository.UpdateTask(ctx, taskPartial)
	if err != nil {
		return nil, err
	}
	return s.GetTaskFull(ctx, taskPartial.ID)
}
