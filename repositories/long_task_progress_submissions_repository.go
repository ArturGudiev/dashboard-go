package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/longtaskprogress"
	"arturgudiev/dashboard/ent/longtaskprogresssubmission"
	"arturgudiev/dashboard/models"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
)

type LongTasksProgressesSubmissionsRepository struct {
	client *ent.Client
}

func NewLongTasksProgressesSubmissionsRepository(client *ent.Client) *LongTasksProgressesSubmissionsRepository {
	return &LongTasksProgressesSubmissionsRepository{client: client}
}

func (r *LongTasksProgressesSubmissionsRepository) GetLongTaskProgressSubmissionById(ctx context.Context, id int) (*ent.LongTaskProgressSubmission, error) {
	return r.client.LongTaskProgressSubmission.Get(ctx, id)
}

// GetLongTaskProgressSubmissions returns all long task progress submissions, newest first (by id).
func (r *LongTasksProgressesSubmissionsRepository) GetLongTaskProgressSubmissionsByIDs(ctx context.Context, ids []int) ([]*ent.LongTaskProgressSubmission, error) {
	return r.client.LongTaskProgressSubmission.Query().
		Where(longtaskprogresssubmission.IDIn(ids...)).
		Order(longtaskprogresssubmission.ByExecutionDate(sql.OrderDesc())).
		All(ctx)
}

func (r *LongTasksProgressesSubmissionsRepository) AddLongTaskProgressSubmission(
	ctx context.Context,
	longTaskProgressID int,
	comments *string,
	progressToAdd *float64,
	progressToSet *float64,
	progressRaw *string,
	executionDate time.Time,
) (*ent.LongTaskProgressSubmission, error) {
	builder := r.client.LongTaskProgressSubmission.Create().
		SetLongTaskProgressID(longTaskProgressID).
		SetExecutionDate(executionDate)

	if comments != nil {
		builder = builder.SetComments(*comments)
	} else {
		builder = builder.SetComments("")
	}
	if progressToAdd != nil {
		builder = builder.SetProgressToAdd(*progressToAdd)
	}
	if progressToSet != nil {
		builder = builder.SetProgressToSet(*progressToSet)
	}
	if progressRaw != nil {
		builder = builder.SetProgressRaw(*progressRaw)
	}

	return builder.Save(ctx)
}

func (r *LongTasksProgressesSubmissionsRepository) GetLongTaskProgressSubmissionsByLongTaskProgressID(
	ctx context.Context,
	longTaskProgressID int,
) ([]*models.LongTaskProgressSubmission, error) {
	submissions, err := r.client.LongTaskProgressSubmission.Query().
		Where(longtaskprogresssubmission.LongTaskProgressIDEQ(longTaskProgressID)).
		Order(longtaskprogresssubmission.ByID(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.LongTaskProgressSubmission, len(submissions))
	for i, submission := range submissions {
		result[i] = &models.LongTaskProgressSubmission{
			ID:            submission.ID,
			Comments:      submission.Comments,
			ProgressToAdd: submission.ProgressToAdd,
			ProgressToSet: submission.ProgressToSet,
			ProgressRaw:   submission.ProgressRaw,
			ExecutionDate: submission.ExecutionDate,
			LongTaskProgressID: submission.LongTaskProgressID,
		}
	}
	return result, nil
}

func (r *LongTasksProgressesSubmissionsRepository) GetLongTaskProgressSubmissionsByLongTaskID(
	ctx context.Context,
	longTaskID int,
) ([]*models.LongTaskProgressSubmission, error) {
	submissions, err := r.client.LongTaskProgressSubmission.Query().
		Where(longtaskprogresssubmission.HasLongTaskProgressWith(
			longtaskprogress.LongTaskIDEQ(longTaskID),
		)).
		Order(longtaskprogresssubmission.ByID(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.LongTaskProgressSubmission, len(submissions))
	for i, submission := range submissions {
		result[i] = &models.LongTaskProgressSubmission{
			ID:            submission.ID,
			Comments:      submission.Comments,
			ProgressToAdd: submission.ProgressToAdd,
			ProgressToSet: submission.ProgressToSet,
			ProgressRaw:   submission.ProgressRaw,
			ExecutionDate: submission.ExecutionDate,
			LongTaskProgressID: submission.LongTaskProgressID,
		}
	}
	return result, nil
}

