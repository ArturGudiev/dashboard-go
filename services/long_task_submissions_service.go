package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
)

type LongTaskSubmissionsService struct {
	longTasksRepository          *repositories.LongTasksRepository
	longTaskSubmissionsRepository *repositories.LongTaskSubmissionsRepository
	longTasksProgressSubmissionsRepository *repositories.LongTasksProgressesSubmissionsRepository
}

func NewLongTaskSubmissionsService(
	longTasksRepository *repositories.LongTasksRepository,
	longTaskSubmissionsRepository *repositories.LongTaskSubmissionsRepository,
	longTasksProgressSubmissionsRepository *repositories.LongTasksProgressesSubmissionsRepository,
) *LongTaskSubmissionsService {
	return &LongTaskSubmissionsService{
		longTasksRepository: longTasksRepository,
		longTaskSubmissionsRepository: longTaskSubmissionsRepository,
		longTasksProgressSubmissionsRepository: longTasksProgressSubmissionsRepository,
	}
}

// AddLongTaskSubmission records a progress update submission and updates the parent LongTask progress/done.
func (s *LongTaskSubmissionsService) AddLongTaskSubmission(
	ctx context.Context,
	longTaskID int,
	comments *string,
	progressToAdd *float64,
	progressToSet *float64,
	progressRaw *string,
) (*ent.LongTaskSubmission, error) {
	submission, err := s.longTaskSubmissionsRepository.AddLongTaskSubmission(
		ctx,
		longTaskID,
		comments,
		progressToAdd,
		progressToSet,
		progressRaw,
	)
	if err != nil {
		return nil, err
	}

	if progressToAdd == nil && progressToSet == nil {
		return submission, nil
	}

	longTask, err := s.longTasksRepository.GetLongTask(ctx, longTaskID)
	if err != nil {
		return nil, err
	}

	if longTask.ProgressTotal == 0 || longTask.ProgressDone == nil {
		return submission, nil
	}

	newProgressDone := *longTask.ProgressDone
	if progressToSet != nil {
		newProgressDone = *progressToSet
	} else if progressToAdd != nil {
		newProgressDone += *progressToAdd
	}

	// done := *longTask.ProgressTotal > 0 && newProgressDone >= *longTask.ProgressTotal

	// We update the long task state after recording the submission.
	// _, err = s.longTasksRepository.UpdateLongTaskProgressDoneAndDone(ctx, longTaskID, newProgressDone, done)
	// if err != nil {
	// 	return nil, err
	// }

	return submission, nil
}

func (s *LongTaskSubmissionsService) GetLongTaskSubmissions(
	ctx context.Context,
	longTaskID int,
) ([]*models.LongTaskProgressSubmission, error) {
	submissions, err := s.longTasksProgressSubmissionsRepository.GetLongTaskProgressSubmissionsByLongTaskID(ctx, longTaskID)
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

