package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/epic"
	"arturgudiev/dashboard/models"
	"context"

	"entgo.io/ent/dialect/sql"
)

type EpicsRepository struct {
	client *ent.Client
}

func NewEpicsRepository(client *ent.Client) *EpicsRepository {
	return &EpicsRepository{client: client}
}

func (r *EpicsRepository) GetAllEpics(ctx context.Context) ([]*ent.Epic, error) {
	epics, err := r.client.Epic.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return epics, nil
}

func (r *EpicsRepository) GetAllOpenEpics(ctx context.Context) ([]*ent.Epic, error) {
	epics, err := r.client.Epic.Query().Where(epic.ClosedEQ(false)).Order(epic.ByID(sql.OrderAsc())).All(ctx)

	if err != nil {
		return nil, err
	}

	return epics, nil
}

func (r *EpicsRepository) GetEpic(ctx context.Context, ID int) (*ent.Epic, error) {
	epic, err := r.client.Epic.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return epic, nil
}

func (r *EpicsRepository) AddEpic(ctx context.Context, description string, tags []string, notes string) (*ent.Epic, error) {
	epic, err := r.client.Epic.Create().SetDescription(description).SetTags(tags).SetNotes(notes).Save(ctx)

	if err != nil {
		return nil, err
	}
	return epic, nil
}

func (r *EpicsRepository) UpdateEpic(ctx context.Context, epic models.EpicPartial) error {
	updateBuilder := r.client.Epic.UpdateOneID(epic.ID)

	if epic.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*epic.Description)
	}

	if epic.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*epic.Notes)
	}

	if epic.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*epic.Tags)
	}

	if epic.DoneDateTime != nil {
		updateBuilder = updateBuilder.SetDoneDateTime(*epic.DoneDateTime)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}
