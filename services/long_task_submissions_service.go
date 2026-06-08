package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/repositories"
	"context"
)

type LongTaskSubmissionsService struct {
	longTasksRepository          *repositories.LongTasksRepository
	longTaskSubmissionsRepository *repositories.LongTaskSubmissionsRepository
}

func NewLongTaskSubmissionsService(
	longTasksRepository *repositories.LongTasksRepository,
	longTaskSubmissionsRepository *repositories.LongTaskSubmissionsRepository,
) *LongTaskSubmissionsService {
	return &LongTaskSubmissionsService{
		longTasksRepository: longTasksRepository,
		longTaskSubmissionsRepository: longTaskSubmissionsRepository,
	}
}

// AddLongTaskSubmission records a progress update submission and updates the parent LongTask progress/done.
func (s *LongTaskSubmissionsService) AddLongTaskSubmission(
	ctx context.Context,
	longTaskID int,
	comments *string,
	progressToAdd *float64,
	progressToSet *float64,
) (*ent.LongTaskSubmission, error) {
	submission, err := s.longTaskSubmissionsRepository.AddLongTaskSubmission(
		ctx,
		longTaskID,
		comments,
		progressToAdd,
		progressToSet,
	)
	if err != nil {
		return nil, err
	}

	longTask, err := s.longTasksRepository.GetLongTask(ctx, longTaskID)
	if err != nil {
		return nil, err
	}

	newProgressDone := longTask.ProgressDone
	if progressToSet != nil {
		newProgressDone = *progressToSet
	} else if progressToAdd != nil {
		newProgressDone += *progressToAdd
	}

	done := longTask.ProgressTotal > 0 && newProgressDone >= longTask.ProgressTotal

	// We update the long task state after recording the submission.
	_, err = s.longTasksRepository.UpdateLongTaskProgressDoneAndDone(ctx, longTaskID, newProgressDone, done)
	if err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *LongTaskSubmissionsService) GetLongTaskSubmissions(
	ctx context.Context,
	longTaskID int,
) ([]*ent.LongTaskSubmission, error) {
	return s.longTaskSubmissionsRepository.GetLongTaskSubmissions(ctx, longTaskID)
}

