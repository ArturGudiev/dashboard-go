package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"context"
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

func (r *TasksRepository) AddTask(ctx context.Context, description string, tags []string, notes string) (*ent.Task, error) {
	task, err := r.client.Task.Create().SetDescription(description).SetTags(tags).SetNotes(notes).Save(ctx)

	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TasksRepository) UpdateTask(ctx context.Context, problem models.TaskPartial) error {
	updateBuilder := r.client.Task.UpdateOneID(problem.ID)

	if problem.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*problem.Description)
	}

	if problem.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*problem.Notes)
	}

	if problem.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*problem.Tags)
	}

	if problem.Done != nil {
		updateBuilder = updateBuilder.SetDone(*problem.Done)
	}

	if problem.DoneDateTime != nil {
		updateBuilder = updateBuilder.SetDoneDateTime(*problem.DoneDateTime)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}
