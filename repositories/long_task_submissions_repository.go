package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/longtasksubmission"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
)

type LongTaskSubmissionsRepository struct {
	client *ent.Client
}

func NewLongTaskSubmissionsRepository(client *ent.Client) *LongTaskSubmissionsRepository {
	return &LongTaskSubmissionsRepository{client: client}
}

func (r *LongTaskSubmissionsRepository) AddLongTaskSubmission(
	ctx context.Context,
	longTaskID int,
	comments *string,
	progressToAdd *float64,
	progressToSet *float64,
	progressRaw *string,
) (*ent.LongTaskSubmission, error) {
	createBuilder := r.client.LongTaskSubmission.Create().
		SetLongTaskID(longTaskID).
		SetExecutionDate(time.Now())

	if comments != nil {
		createBuilder = createBuilder.SetNillableComments(comments)
	}
	if progressToAdd != nil {
		createBuilder = createBuilder.SetNillableProgressToAdd(progressToAdd)
	}
	if progressToSet != nil {
		createBuilder = createBuilder.SetNillableProgressToSet(progressToSet)
	}
	if progressRaw != nil {
		createBuilder = createBuilder.SetNillableProgressRaw(progressRaw)
	}

	return createBuilder.Save(ctx)
}

// GetLongTaskSubmissions returns all submissions for a long task, newest first.
func (r *LongTaskSubmissionsRepository) GetLongTaskSubmissions(
	ctx context.Context,
	longTaskID int,
) ([]*ent.LongTaskSubmission, error) {
	return r.client.LongTaskSubmission.Query().
		Where(longtasksubmission.LongTaskIDEQ(longTaskID)).
		Order(longtasksubmission.ByExecutionDate(sql.OrderDesc())).
		All(ctx)
}

