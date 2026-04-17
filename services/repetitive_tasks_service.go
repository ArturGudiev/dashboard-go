package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/repositories"
	"context"
)

// TaskService handles task-related business logic
type RepetitiveTaskService struct {
	client                    *ent.Client
	containerService          *ContainerService
	problemService            *ProblemService
	repetitiveTasksRepository *repositories.RepetitiveTasksRepository
	repetitiveTaskExecutionsRepository *repositories.RepetitiveTaskExecutionsRepository
	childContainerRepository  *ChildContainerRepository
}

// NewRepetitiveTaskService creates a new RepetitiveTaskService
func NewRepetitiveTaskService(client *ent.Client, containerService *ContainerService, problemService *ProblemService,
	repetitiveTasksRepository *repositories.RepetitiveTasksRepository, childContainerRepository *ChildContainerRepository) *RepetitiveTaskService {
	return &RepetitiveTaskService{
		client:                    client,
		containerService:          containerService,
		problemService:            problemService,
		repetitiveTasksRepository: repetitiveTasksRepository,
		childContainerRepository:  childContainerRepository,
	}
}

func (s *RepetitiveTaskService) GetRepetitiveTasks(ctx context.Context) ([]*ent.RepetitiveTask, error) {
	repetitiveTasks, err := s.repetitiveTasksRepository.GetRepetitiveTasks(ctx)
	if err != nil {
		return nil, err
	}

	return repetitiveTasks, nil
}


func (s *RepetitiveTaskService) GetRepetitiveTaskById(ctx context.Context, ID int) (*ent.RepetitiveTask, error) {
	repetitiveTasks, err := s.repetitiveTasksRepository.GetRepetitiveTaskById(ctx, ID)
	if err != nil {
		return nil, err
	}

	return repetitiveTasks, nil
}

func (s *RepetitiveTaskService) IsTaskActual(ctx context.Context, task ent.RepetitiveTask) (bool, error) {
	executions, err := s.repetitiveTaskExecutionsRepository.GetRepetitiveTaskExecutions(ctx, task.ID)
	if (err != nil) {
		return nil, err
	}
	if (len(executions) == 0) {
		return true, nil
	}
	latest := executions[0]
	if (task.OnceInDays != nil) {
		latest.ExecutionDate, 
		// now := time.Now();
	}

	return repetitiveTasks, nil
}