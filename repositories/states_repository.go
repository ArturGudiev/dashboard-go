package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/state"
	"arturgudiev/dashboard/models"
	"context"

	"entgo.io/ent/dialect/sql"
)

type StatesRepository struct {
	client *ent.Client
}

func NewStatesRepository(client *ent.Client) *StatesRepository {
	return &StatesRepository{client: client}
}

func (r *StatesRepository) GetState(ctx context.Context, ID int) (*ent.State, error) {
	state, err := r.client.State.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (r *StatesRepository) AddState(ctx context.Context, description string, tags []string, notes string) (*ent.State, error) {
	state, err := r.client.State.Create().
		SetDescription(description).
		SetTags(tags).
		SetNotes(notes).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (r *StatesRepository) UpdateState(ctx context.Context, state models.StatePartial) error {
	updateBuilder := r.client.State.UpdateOneID(state.ID)

	if state.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*state.Description)
	}

	if state.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*state.Notes)
	}

	if state.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*state.Tags)
	}

	if state.Closed != nil {
		updateBuilder = updateBuilder.SetClosed(*state.Closed)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}

func (r *StatesRepository) GetAllStates(ctx context.Context) ([]*ent.State, error) {
	states, err := r.client.State.Query().Where(state.ClosedEQ(false)).Order(state.ByID(sql.OrderDesc())).All(ctx)
	if err != nil {
		return nil, err
	}
	return states, nil
}

func (r *StatesRepository) DeleteState(ctx context.Context, ID int) error {
	err := r.client.State.DeleteOneID(ID).Exec(ctx)
	return err
}