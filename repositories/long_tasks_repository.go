package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/longtask"
	"arturgudiev/dashboard/models"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
)

type LongTasksRepository struct {
	client *ent.Client
}

func NewLongTasksRepository(client *ent.Client) *LongTasksRepository {
	return &LongTasksRepository{client: client}
}

func (r *LongTasksRepository) GetLongTask(ctx context.Context, id int) (*ent.LongTask, error) {
	return r.client.LongTask.Get(ctx, id)
}

func (r *LongTasksRepository) GetLongTaskById(ctx context.Context, id int) (*ent.LongTask, error) {
	return r.client.LongTask.Query().Where(longtask.IDEQ(id)).WithProgresses().First(ctx)
}

// GetLongTasks returns all long tasks, newest first (by id).
func (r *LongTasksRepository) GetLongTasks(ctx context.Context) ([]*ent.LongTask, error) {
	return r.client.LongTask.Query().
		Order(longtask.ByID(sql.OrderDesc())).
		All(ctx)
}

// GetLongTasks returns all long tasks, newest first (by id).
func (r *LongTasksRepository) GetLongTasksWithProgresses(ctx context.Context) ([]*ent.LongTask, error) {
	return r.client.LongTask.Query().
		WithProgresses().
		Order(longtask.ByID(sql.OrderDesc())).
		All(ctx)
}

func (r *LongTasksRepository) AddLongTask(
	ctx context.Context,
	description string,
	tags []string,
	notes string,
	progressTotal *float64,
	progressDone *float64,
	progressUnits *string,
) (*ent.LongTask, error) {
	done := false
	if progressTotal != nil && progressDone != nil && *progressTotal > 0 && *progressDone >= *progressTotal {
		done = true
	}
	newTask := r.client.LongTask.Create().
		SetDescription(description).
		SetTags(tags).
		SetNotes(notes).
		SetDone(done)

	if progressTotal != nil {
		newTask = newTask.SetProgressTotal(*progressTotal)
	}
	if progressDone != nil {
		newTask = newTask.SetProgressDone(*progressDone)
	}
	if progressUnits != nil {
		newTask = newTask.SetProgressUnits(*progressUnits)
	}
	if done {
		newTask = newTask.SetDoneDateTime(time.Now())
	}

	return newTask.Save(ctx)
}

// UpdateLongTaskProgressDoneAndDone updates progress_done and done (and done_date_time).
func (r *LongTasksRepository) UpdateLongTaskProgressDoneAndDone(
	ctx context.Context,
	id int,
	progressDone float64,
	done bool,
) (*ent.LongTask, error) {
	updateBuilder := r.client.LongTask.UpdateOneID(id).
		SetProgressDone(progressDone).
		SetDone(done)

	if done {
		updateBuilder = updateBuilder.SetDoneDateTime(time.Now())
	} else {
		updateBuilder = updateBuilder.ClearDoneDateTime()
	}

	return updateBuilder.Save(ctx)
}

func (r *LongTasksRepository) UpdateLongTask(ctx context.Context, partial models.LongTaskPartial) error {
	updateBuilder := r.client.LongTask.UpdateOneID(partial.ID)

	if partial.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*partial.Description)
	}
	if partial.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*partial.Notes)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}

