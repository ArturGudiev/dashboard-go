package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/repositories"
	"context"
)

type RepetitiveTaskExecutionService struct {
	repetitiveTaskExecutionsRepository *repositories.RepetitiveTaskExecutionsRepository
}

func NewRepetitiveTaskExecutionService(
	repetitiveTaskExecutionsRepository *repositories.RepetitiveTaskExecutionsRepository,
) *RepetitiveTaskExecutionService {
	return &RepetitiveTaskExecutionService{
		repetitiveTaskExecutionsRepository: repetitiveTaskExecutionsRepository,
	}
}

func (s *RepetitiveTaskExecutionService) AddRepetitiveTaskExecution(ctx context.Context, repetitiveTaskID int) (*ent.RepetitiveTaskExecution, error) {
	return s.repetitiveTaskExecutionsRepository.AddRepetitiveTaskExecution(ctx, repetitiveTaskID)
}

func (s *RepetitiveTaskExecutionService) GetRepetitiveTaskExecutions(ctx context.Context, repetitiveTaskID int) ([]*ent.RepetitiveTaskExecution, error) {
	return s.repetitiveTaskExecutionsRepository.GetRepetitiveTaskExecutions(ctx, repetitiveTaskID)
}
