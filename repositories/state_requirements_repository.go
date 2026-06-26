package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/staterequirement"
	"arturgudiev/dashboard/models"
	"context"
)

type StateRequirementsRepository struct {
	client *ent.Client
}

func NewStateRequirementsRepository(client *ent.Client) *StateRequirementsRepository {
	return &StateRequirementsRepository{client: client}
}

func (r *StateRequirementsRepository) GetStateRequirement(ctx context.Context, ID int) (*ent.StateRequirement, error) {
	stateRequirement, err := r.client.StateRequirement.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return stateRequirement, nil
}

func (r *StateRequirementsRepository) GetStateRequirementsByIDs(ctx context.Context, IDs []int) ([]*ent.StateRequirement, error) {
	stateRequirements, err := r.client.StateRequirement.Query().Where(staterequirement.IDIn(IDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return stateRequirements, nil
}

func (r *StateRequirementsRepository) GetStateRequirementsByStateID(ctx context.Context, stateID int) ([]*ent.StateRequirement, error) {
	stateRequirements, err := r.client.StateRequirement.Query().Where(staterequirement.StateID(stateID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return stateRequirements, nil
}

func (r *StateRequirementsRepository) AddStateRequirement(ctx context.Context, description string, stateID int, onceInDays *int) (*ent.StateRequirement, error) {
	stateRequirement, err := r.client.StateRequirement.Create().
		SetDescription(description).
		SetStateID(stateID).
		SetNillableOnceInDays(onceInDays).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return stateRequirement, nil
}

func (r *StateRequirementsRepository) UpdateStateRequirement(ctx context.Context, stateRequirementID int, stateRequirement models.StateRequirementPartial) error {
	updateBuilder := r.client.StateRequirement.UpdateOneID(stateRequirementID)

	if stateRequirement.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*stateRequirement.Description)
	}

	if stateRequirement.OnceInDays != nil {
		updateBuilder = updateBuilder.SetNillableOnceInDays(stateRequirement.OnceInDays)
	}

	if stateRequirement.StateID != nil {
		updateBuilder = updateBuilder.SetStateID(*stateRequirement.StateID)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}

func (r *StateRequirementsRepository) DeleteStateRequirement(ctx context.Context, ID int) error {
	err := r.client.StateRequirement.DeleteOneID(ID).Exec(ctx)
	return err
}
