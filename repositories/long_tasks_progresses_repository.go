package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/longtaskprogress"
	"context"

	"entgo.io/ent/dialect/sql"
)

type LongTasksProgressesRepository struct {
	client *ent.Client
}

func NewLongTasksProgressesRepository(client *ent.Client) *LongTasksProgressesRepository {
	return &LongTasksProgressesRepository{client: client}
}

func (r *LongTasksProgressesRepository) GetLongTaskProgressById(ctx context.Context, id int) (*ent.LongTaskProgress, error) {
	return r.client.LongTaskProgress.Get(ctx, id)
}


// GetLongTasks returns all long tasks, newest first (by id).
func (r *LongTasksProgressesRepository) GetLongTaskProgressesByIDs(ctx context.Context, ids []int) ([]*ent.LongTaskProgress, error) {
	return r.client.LongTaskProgress.Query().
		Where(longtaskprogress.IDIn(ids...)).
		Order(longtaskprogress.ByLongTaskID(sql.OrderDesc())).
		All(ctx)	
}

func (r *LongTasksProgressesRepository) AddLongTaskProgress(
	ctx context.Context,
	name string,
	longTaskID int,
	value *float64,
	total *float64,
	units *string,
) (*ent.LongTaskProgress, error) {
	newTask := r.client.LongTaskProgress.Create().
		SetName(name).
		SetLongTaskID(longTaskID)

	if value != nil {
		newTask = newTask.SetValue(*value)
	}
	if total != nil {
		newTask = newTask.SetTotal(*total)
	}
	if units != nil {
		newTask = newTask.SetUnits(*units)
	}

	return newTask.Save(ctx)
}

// UpdateLongTaskProgressDoneAndDone updates progress_done and done (and done_date_time).
func (r *LongTasksRepository) UpdateLongTaskProgressValue(
	ctx context.Context,
	id int,
	value float64,
) (*ent.LongTaskProgress, error) {
	updateBuilder := r.client.LongTaskProgress.UpdateOneID(id).
		SetValue(value)

	return updateBuilder.Save(ctx)
}

