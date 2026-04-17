package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/repetitivetask"
	"context"
)

type RepetitiveTasksRepository struct {
	client *ent.Client
}

func NewRepetitiveTasksRepository(client *ent.Client) *RepetitiveTasksRepository {
	return &RepetitiveTasksRepository{client: client}
}

func (r *RepetitiveTasksRepository) GetRepetitiveTask(ctx context.Context, ID int) (*ent.RepetitiveTask, error) {
	repetitiveTask, err := r.client.RepetitiveTask.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return repetitiveTask, nil
}


func (r *RepetitiveTasksRepository) GetRepetitiveTasks(ctx context.Context) ([]*ent.RepetitiveTask, error) {
	repetitiveTasks, err := r.client.RepetitiveTask.Query().Where(repetitivetask.ClosedEQ(false)).All(ctx)
	if err != nil {
		return nil, err
	}
	return repetitiveTasks, nil
}

func (r *RepetitiveTasksRepository) GetRepetitiveTaskById(ctx context.Context, ID int) (*ent.RepetitiveTask, error) {
	repetitiveTask, err := r.client.RepetitiveTask.Query().Where(repetitivetask.IDEQ(ID)).First(ctx)
	if err != nil {
		return nil, err
	}
	return repetitiveTask, nil
}

func (r *RepetitiveTasksRepository) SetRepetitiveTaskClosed(ctx context.Context, ID int, solution string) error {
	updateBuilder := r.client.RepetitiveTask.UpdateOneID(ID).
		SetClosed(true)

	_, err := updateBuilder.Save(ctx)
	return err
}

func (r *RepetitiveTasksRepository) AddRepetitiveTask(
	ctx context.Context,
	description string,
	tags []string,
	notes string,
	onceInDays *int,
	onceInWeeks *int,
	onceInMonths *int,
) (*ent.RepetitiveTask, error) {
	repetitiveTask := r.client.RepetitiveTask.Create().SetDescription(description).SetTags(tags).SetNotes(notes)
	if onceInDays != nil {
		repetitiveTask = repetitiveTask.SetOnceInDays(*onceInDays)
	}
	if onceInWeeks != nil {
		repetitiveTask = repetitiveTask.SetOnceInWeeks(*onceInWeeks)
	}
	if onceInMonths != nil {
		repetitiveTask = repetitiveTask.SetOnceInMonths(*onceInMonths)
	}
	newRepetitiveTask, err := repetitiveTask.Save(ctx)
	if err != nil {
		return nil, err
	}
	return newRepetitiveTask, nil
}
