package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/longtaskprogresssubmission"
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
	progressToAdd *int,
	progressToSet *float64,
	progressRaw *float64,
	executionDate time.Time,
) (*ent.LongTaskProgressSubmission, error) {
	return r.client.LongTaskProgressSubmission.Create().Save(ctx)
}


func (r *LongTasksProgressesSubmissionsRepository) GetLongTaskProgressSubmissionsByLongTaskProgressID(
	ctx context.Context,
	longTaskProgressID int,
) ([]*ent.LongTaskProgressSubmission, error) {
	return r.client.LongTaskProgressSubmission.Query().
		Where(longtaskprogresssubmission.LongTaskProgressIDEQ(longTaskProgressID)).
		Order(longtaskprogresssubmission.ByID(sql.OrderDesc())).
		All(ctx);	
}
