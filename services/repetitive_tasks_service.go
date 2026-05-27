package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"fmt"
	"slices"
	"time"
)

// TaskService handles task-related business logic
type RepetitiveTaskService struct {
	client                             *ent.Client
	containerService                   *ContainerService
	problemService                     *ProblemService
	repetitiveTasksRepository          *repositories.RepetitiveTasksRepository
	repetitiveTaskExecutionsRepository *repositories.RepetitiveTaskExecutionsRepository
	childContainerRepository           *ChildContainerRepository
}

// NewRepetitiveTaskService creates a new RepetitiveTaskService
func NewRepetitiveTaskService(client *ent.Client, containerService *ContainerService, problemService *ProblemService,
	repetitiveTasksRepository *repositories.RepetitiveTasksRepository, repetitiveTaskExecutionsRepository *repositories.RepetitiveTaskExecutionsRepository, childContainerRepository *ChildContainerRepository) *RepetitiveTaskService {
	return &RepetitiveTaskService{
		client:                             client,
		containerService:                   containerService,
		problemService:                     problemService,
		repetitiveTasksRepository:          repetitiveTasksRepository,
		repetitiveTaskExecutionsRepository: repetitiveTaskExecutionsRepository,
		childContainerRepository:           childContainerRepository,
	}
}

func (s *RepetitiveTaskService) GetRepetitiveTasks(ctx context.Context, actual *bool) ([]*ent.RepetitiveTask, error) {

	repetitiveTasks, err := s.repetitiveTasksRepository.GetRepetitiveTasks(ctx)
	if err != nil {
		return nil, err
	}

	if actual == nil {
		fmt.Println(actual)
		return repetitiveTasks, nil
	}
	fmt.Println(*actual)

	repetitiveTasks = slices.DeleteFunc(repetitiveTasks, func(rTask *ent.RepetitiveTask) bool {
		isActual, err := s.IsRepetitiveTaskActual(ctx, *rTask)
		if err != nil {
			return false
		}
		if *actual == true {
			return !isActual
		}
		return isActual
	})

	repetitiveTasks = s.filterActualRepetitiveTasks(ctx, repetitiveTasks)
	return repetitiveTasks, nil
}

func (s *RepetitiveTaskService) filterActualRepetitiveTasks(ctx context.Context, repetitiveTasks []*ent.RepetitiveTask) []*ent.RepetitiveTask {
	for i, task := range repetitiveTasks {
		actual, err := s.IsRepetitiveTaskActual(ctx, *task)
		if err != nil {
			continue
		}
		if !actual {
			repetitiveTasks = repetitiveTasks[:i]
			break
		}
	}
	return repetitiveTasks
}

func (s *RepetitiveTaskService) GetRepetitiveTaskById(ctx context.Context, ID int) (*ent.RepetitiveTask, error) {
	repetitiveTasks, err := s.repetitiveTasksRepository.GetRepetitiveTaskById(ctx, ID)
	if err != nil {
		return nil, err
	}

	return repetitiveTasks, nil
}

func (s *RepetitiveTaskService) IsRepetitiveTaskActual(ctx context.Context, task ent.RepetitiveTask) (bool, error) {
	executions, err := s.repetitiveTaskExecutionsRepository.GetRepetitiveTaskExecutions(ctx, task.ID)
	if err != nil {
		return false, err
	}
	if len(executions) == 0 {
		return true, nil
	}

	latestExecutionDate := executions[0].ExecutionDate
	now := time.Now()

	if task.OnceInDays != nil {
		nextExecutionDate := latestExecutionDate.AddDate(0, 0, *task.OnceInDays)
		return !nextExecutionDate.After(now), nil
	}

	if task.OnceInWeeks != nil {
		nextExecutionDate := latestExecutionDate.AddDate(0, 0, *task.OnceInWeeks*7)
		return !nextExecutionDate.After(now), nil
	}

	if task.OnceInMonths != nil {
		nextExecutionDate := latestExecutionDate.AddDate(0, *task.OnceInMonths, 0)
		return !nextExecutionDate.After(now), nil
	}

	return false, nil
}

func (s *RepetitiveTaskService) UpdateRepetitiveTask(ctx context.Context, partial models.RepetitiveTaskPartial) (*ent.RepetitiveTask, error) {
	if err := s.repetitiveTasksRepository.UpdateRepetitiveTask(ctx, partial); err != nil {
		return nil, err
	}
	return s.repetitiveTasksRepository.GetRepetitiveTaskById(ctx, partial.ID)
}

func (s *RepetitiveTaskService) AddRepetitiveTask(ctx context.Context, task models.RepetitiveTaskShort, parent *models.ContainerDescription) (*ent.RepetitiveTask, error) {
	newRepetitiveTask, err := s.repetitiveTasksRepository.AddRepetitiveTask(ctx, task.Description, task.Tags, task.Notes, task.OnceInDays, task.OnceInWeeks, task.OnceInMonths)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeRepetitiveTask, newRepetitiveTask.ID)
		if err != nil {
			return nil, err
		}
	}
	return newRepetitiveTask, nil
}
