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
	return s.longTasksProgressesRepository.GetLongTaskProgressesByIDs(ctx, []int{longTaskID})
}

func (s *LongTaskProgressesService) AddLongTaskProgressSubmission(
	ctx context.Context,
	longTaskProgressID int,
	comments *string,
	progressToAdd *float64,
	progressToSet *float64,
	progressRaw *float64,
	executionDate *time.Time,
) (*ent.LongTaskProgressSubmission, error) {
	execDate := time.Now()
	if executionDate != nil {
		execDate = *executionDate
	}

	var progressToAddInt *int
	if progressToAdd != nil {
		v := int(*progressToAdd)
		progressToAddInt = &v
	}

	return s.longTasksProgressSubmissionsRepository.AddLongTaskProgressSubmission(
		ctx,
		longTaskProgressID,
		comments,
		progressToAddInt,
		progressToSet,
		progressRaw,
		execDate,
	)
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
