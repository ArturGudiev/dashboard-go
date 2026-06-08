package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"slices"
)

type LongTasksService struct {
	longTasksRepository      *repositories.LongTasksRepository
	childContainerRepository *ChildContainerRepository
}

func NewLongTasksService(
	longTasksRepository *repositories.LongTasksRepository,
	childContainerRepository *ChildContainerRepository,
) *LongTasksService {
	return &LongTasksService{
		longTasksRepository:      longTasksRepository,
		childContainerRepository: childContainerRepository,
	}
}

func (s *LongTasksService) GetLongTasks(ctx context.Context, open *bool) ([]*ent.LongTask, error) {
	longTasks, err := s.longTasksRepository.GetLongTasks(ctx)
	if err != nil {
		return nil, err
	}

	if open != nil && *open {
		longTasks = slices.DeleteFunc(longTasks, func(task *ent.LongTask) bool {
			return task.Done
		})
	}

	return longTasks, nil
}

func (s *LongTasksService) GetLongTaskById(ctx context.Context, id int) (*ent.LongTask, error) {
	return s.longTasksRepository.GetLongTaskById(ctx, id)
}

func (s *LongTasksService) UpdateLongTask(ctx context.Context, partial models.LongTaskPartial) (*ent.LongTask, error) {
	if err := s.longTasksRepository.UpdateLongTask(ctx, partial); err != nil {
		return nil, err
	}
	return s.longTasksRepository.GetLongTaskById(ctx, partial.ID)
}

func (s *LongTasksService) AddLongTask(
	ctx context.Context,
	task models.LongTaskShort,
	parent *models.ContainerDescription,
) (*ent.LongTask, error) {
	newLongTask, err := s.longTasksRepository.AddLongTask(
		ctx,
		task.Description,
		task.Tags,
		task.Notes,
		task.ProgressTotal,
		task.ProgressDone,
		task.ProgressUnits,
	)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeLongTask, newLongTask.ID)
		if err != nil {
			return nil, err
		}
	}
	return newLongTask, nil
}

func (s *LongTasksService) UpdateLongTaskProgressDoneAndDone(
	ctx context.Context,
	id int,
	progressDone float64,
	done bool,
) (*ent.LongTask, error) {
	return s.longTasksRepository.UpdateLongTaskProgressDoneAndDone(ctx, id, progressDone, done)
}
