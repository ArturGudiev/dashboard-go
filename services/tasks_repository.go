package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/task"
	"arturgudiev/dashboard/models"
	"context"
	"time"
)

type TasksRepository struct {
	client *ent.Client
}

func NewTasksRepository(client *ent.Client) *TasksRepository {
	return &TasksRepository{client: client}
}

func (r *TasksRepository) GetTask(ctx context.Context, ID int) (*ent.Task, error) {
	task, err := r.client.Task.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TasksRepository) SetTaskDone(ctx context.Context, ID int, solution string) error {
	updateBuilder := r.client.Task.UpdateOneID(ID).
		SetDone(true)

	_, err := updateBuilder.Save(ctx)
	return err
}

func (r *TasksRepository) AddTask(ctx context.Context, description string, tags []string, notes string, done bool, dueDateTime *time.Time) (*ent.Task, error) {
	builder := r.client.Task.Create().
		SetDescription(description).
		SetTags(tags).
		SetNotes(notes).
		SetDone(done)
	if dueDateTime != nil {
		builder = builder.SetDueDateTime(*dueDateTime)
	}
	task, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TasksRepository) AddTaskByFields(ctx context.Context, fields models.TaskFieldsPartial) (*ent.Task, error) {
	task := r.client.Task.Create()

	if fields.Description != nil {
		task.SetDescription(*fields.Description)
	}

	if fields.Tags != nil {
		task.SetTags(*fields.Tags)
	}

	if fields.Notes != nil {
		task.SetNotes(*fields.Notes)
	}

	if fields.Done != nil {
		task.SetDone(*fields.Done)
	}

	if fields.DoneDateTime != nil {
		task.SetDoneDateTime(*fields.DoneDateTime)
	}

	if fields.DueDateTime != nil {
		task.SetDueDateTime(*fields.DueDateTime)
	}

	newTask, err := task.Save(ctx)

	if err != nil {
		return nil, err
	}
	return newTask, nil
}

func (r *TasksRepository) UpdateTask(ctx context.Context, task models.TaskPartial) error {
	updateBuilder := r.client.Task.UpdateOneID(task.ID)

	if task.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*task.Description)
	}

	if task.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*task.Notes)
	}

	if task.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*task.Tags)
	}

	if task.Done != nil {
		updateBuilder = updateBuilder.SetDone(*task.Done)
	}

	if task.DoneDateTime != nil {
		updateBuilder = updateBuilder.SetDoneDateTime(*task.DoneDateTime)
	}

	if task.DueDateTime != nil {
		updateBuilder = updateBuilder.SetDueDateTime(*task.DueDateTime)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}

func (r *TasksRepository) GetOpenTasksByDueDate(ctx context.Context, dayStart time.Time) ([]*ent.Task, error) {
	dayEnd := dayStart.Add(24 * time.Hour)
	return r.client.Task.Query().
		Where(
			task.DoneEQ(false),
			task.DueDateTimeNotNil(),
			task.DueDateTimeGTE(dayStart),
			task.DueDateTimeLT(dayEnd),
		).
		Order(task.ByID()).
		All(ctx)
}

func (r *TasksRepository) getDoneTasksCountInRange(ctx context.Context, start time.Time, end time.Time) (int, error) {
	tasks, err := r.client.Task.Query().
		Where(
			task.DoneEQ(true),
			task.DoneDateTimeNotNil(),
			task.DoneDateTimeGTE(start),
			task.DoneDateTimeLT(end),
		).All(ctx)
	if err != nil {
		return 0, err
	}
	return len(tasks), nil
}
