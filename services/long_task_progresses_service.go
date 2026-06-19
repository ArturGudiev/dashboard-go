package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"time"
)

type LongTaskProgressesService struct {
	longTasksProgressesRepository          *repositories.LongTasksProgressesRepository
	longTasksProgressSubmissionsRepository *repositories.LongTasksProgressesSubmissionsRepository
}

func NewLongTaskProgressesService(
	longTasksProgressesRepository *repositories.LongTasksProgressesRepository,
	longTasksProgressSubmissionsRepository *repositories.LongTasksProgressesSubmissionsRepository,
) *LongTaskProgressesService {
	return &LongTaskProgressesService{
		longTasksProgressesRepository:          longTasksProgressesRepository,
		longTasksProgressSubmissionsRepository: longTasksProgressSubmissionsRepository,
	}
}

func (s *LongTaskProgressesService) GetLongTaskProgresses(
	ctx context.Context,
	longTaskID int,
) ([]*ent.LongTaskProgress, error) {
	return s.longTasksProgressesRepository.GetLongTaskProgressesByLongTaskID(ctx, longTaskID)
}

func (s *LongTaskProgressesService) AddLongTaskProgressSubmission(
	ctx context.Context,
	longTaskProgressID int,
	comments *string,
	progressToAdd *float64,
	progressToSet *float64,
	progressRaw *string,
	executionDate *time.Time,
) (*ent.LongTaskProgressSubmission, error) {
	execDate := time.Now()
	if executionDate != nil {
		execDate = *executionDate
	}

	submission, err := s.longTasksProgressSubmissionsRepository.AddLongTaskProgressSubmission(
		ctx,
		longTaskProgressID,
		comments,
		progressToAdd,
		progressToSet,
		progressRaw,
		execDate,
	)
	if err != nil {
		return nil, err
	}

	if progressToAdd == nil && progressToSet == nil {
		return submission, nil
	}

	progress, err := s.longTasksProgressesRepository.GetLongTaskProgressById(ctx, longTaskProgressID)
	if err != nil {
		return nil, err
	}

	if progress.Total == nil || progress.Value == nil {
		return submission, nil
	}

	newValue := *progress.Value
	if progressToSet != nil {
		newValue = *progressToSet
	} else if progressToAdd != nil {
		newValue += *progressToAdd
	}

	if _, err := s.longTasksProgressesRepository.UpdateLongTaskProgressValue(ctx, longTaskProgressID, newValue); err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *LongTaskProgressesService) AddLongTaskProgress(
	ctx context.Context,
	name string,
	longTaskID int,
	value *float64,
	total *float64,
	units *string,
) (*ent.LongTaskProgress, error) {
	return s.longTasksProgressesRepository.AddLongTaskProgress(ctx, name, longTaskID, value, total, units)
}

func (s *LongTaskProgressesService) GetLongTaskProgressSubmissions(
	ctx context.Context,
	longTaskProgressID int,
) ([]*models.LongTaskProgressSubmission, error) {
	return s.longTasksProgressSubmissionsRepository.GetLongTaskProgressSubmissionsByLongTaskProgressID(ctx, longTaskProgressID)
}

func (s *LongTaskProgressesService) GetLongTaskProgressByID(ctx context.Context, id int) (*models.LongTaskProgressFull, error) {
	progress, err := s.longTasksProgressesRepository.GetLongTaskProgressById(ctx, id)
	if err != nil {
		return nil, err
	}

	submissions, err := s.longTasksProgressSubmissionsRepository.GetLongTaskProgressSubmissionsByLongTaskProgressID(ctx, id)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	return &models.LongTaskProgressFull{
		ID:          progress.ID,
		Name:        progress.Name,
		Value:       progress.Value,
		Total:       progress.Total,
		Units:       progress.Units,
		Submissions: submissions,
	}, nil
}
