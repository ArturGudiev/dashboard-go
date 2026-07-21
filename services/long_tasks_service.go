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
	longTasksRepository           *repositories.LongTasksRepository
	childContainerRepository      *ChildContainerRepository
	longTasksProgressesRepository *repositories.LongTasksProgressesRepository
}

func NewLongTasksService(
	longTasksRepository *repositories.LongTasksRepository,
	childContainerRepository *ChildContainerRepository,
	longTasksProgressesRepository *repositories.LongTasksProgressesRepository,
) *LongTasksService {
	return &LongTasksService{
		longTasksRepository:           longTasksRepository,
		childContainerRepository:      childContainerRepository,
		longTasksProgressesRepository: longTasksProgressesRepository,
	}
}

func filterOpenLongTasks(longTasks []*ent.LongTask) []*ent.LongTask {
	return slices.DeleteFunc(longTasks, func(task *ent.LongTask) bool {
		return task.Done
	})
}

func shouldReturnOpenLongTasksOnly(open *bool) bool {
	return open == nil || *open
}

func (s *LongTasksService) GetLongTasks(ctx context.Context, open *bool) ([]*ent.LongTask, error) {
	longTasks, err := s.longTasksRepository.GetLongTasks(ctx)
	if err != nil {
		return nil, err
	}

	if shouldReturnOpenLongTasksOnly(open) {
		longTasks = filterOpenLongTasks(longTasks)
	}

	return longTasks, nil
}

func (s *LongTasksService) GetLongTasksFull(ctx context.Context, open *bool) ([]*models.LongTaskFull, error) {
	longTasks, err := s.longTasksRepository.GetLongTasksWithProgresses(ctx)
	if err != nil {
		return nil, err
	}

	if shouldReturnOpenLongTasksOnly(open) {
		longTasks = filterOpenLongTasks(longTasks)
	}

	longTasksFull := make([]*models.LongTaskFull, len(longTasks))
	for i, longTask := range longTasks {
		progresses := make([]models.LongTaskProgress, len(longTask.Edges.Progresses))
		for j, progress := range longTask.Edges.Progresses {
			progresses[j] = models.LongTaskProgress{
				ID:    progress.ID,
				Name:  progress.Name,
				Value: progress.Value,
				Total: progress.Total,
				Units: progress.Units,
			}
		}
		longTasksFull[i] = &models.LongTaskFull{
			ID:           longTask.ID,
			Description:  longTask.Description,
			Tags:         longTask.Tags,
			Done:         longTask.Done,
			DoneDateTime: longTask.DoneDateTime,
			Progresses:   progresses,
			Notes:        longTask.Notes,
		}
	}
	return longTasksFull, nil

}

func (s *LongTasksService) UpdateLongTask(ctx context.Context, partial models.LongTaskPartial) (*ent.LongTask, error) {
	if err := s.longTasksRepository.UpdateLongTask(ctx, partial); err != nil {
		return nil, err
	}
	return s.longTasksRepository.GetLongTaskById(ctx, partial.ID)
}

func (s *LongTasksService) CloseLongTask(ctx context.Context, id int) (*ent.LongTask, error) {
	done := true
	return s.UpdateLongTask(ctx, models.LongTaskPartial{
		ID:   id,
		Done: &done,
	})
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
