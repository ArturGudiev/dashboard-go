package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/staterequirementcheck"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
)

type StateRequirementChecksRepository struct {
	client *ent.Client
}

func NewStateRequirementChecksRepository(client *ent.Client) *StateRequirementChecksRepository {
	return &StateRequirementChecksRepository{client: client}
}

func (r *StateRequirementChecksRepository) GetStateRequirementCheck(ctx context.Context, ID int) (*ent.StateRequirementCheck, error) {
	stateRequirementCheck, err := r.client.StateRequirementCheck.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return stateRequirementCheck, nil
}

func (r *StateRequirementChecksRepository) GetStateRequirementChecks(ctx context.Context, IDs []int) ([]*ent.StateRequirementCheck, error) {
	stateRequirementChecks, err := r.client.StateRequirementCheck.Query().Where(staterequirementcheck.IDIn(IDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return stateRequirementChecks, nil
}

func (r *StateRequirementChecksRepository) GetStateRequirementChecksByStateRequirementID(ctx context.Context, stateRequirementID int) ([]*ent.StateRequirementCheck, error) {
	stateRequirementChecks, err := r.client.StateRequirementCheck.Query().
		Where(staterequirementcheck.StateRequirementID(stateRequirementID)).
		Order(staterequirementcheck.ByDateTime(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return stateRequirementChecks, nil
}

func (r *StateRequirementChecksRepository) GetLatestStateRequirementCheck(ctx context.Context, stateRequirementID int) (*ent.StateRequirementCheck, error) {
	stateRequirementCheck, err := r.client.StateRequirementCheck.Query().
		Where(staterequirementcheck.StateRequirementID(stateRequirementID)).
		Order(staterequirementcheck.ByDateTime(sql.OrderDesc())).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return stateRequirementCheck, nil
}

func (r *StateRequirementChecksRepository) AddStateRequirementCheck(ctx context.Context, stateRequirementID int, isFulfilled bool) (*ent.StateRequirementCheck, error) {
	stateRequirementCheck, err := r.client.StateRequirementCheck.Create().
		SetDateTime(time.Now()).
		SetIsFulfilled(isFulfilled).
		SetStateRequirementID(stateRequirementID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return stateRequirementCheck, nil
}

// func (r *StateRequirementChecksRepository) UpdateStateRequirementCheck(ctx context.Context, stateRequirementCheckID int, stateRequirementCheck models.StateRequirementCheckPartial) error {
// 	updateBuilder := r.client.StateRequirementCheck.UpdateOneID(stateRequirementCheckID)

// 	if stateRequirement.Description != nil {
// 		updateBuilder = updateBuilder.SetDescription(*stateRequirement.Description)
// 	}

// 	if stateRequirement.OnceInDays != nil {
// 		updateBuilder = updateBuilder.SetNillableOnceInDays(stateRequirement.OnceInDays)
// 	}

// 	if stateRequirement.StateID != nil {
// 		updateBuilder = updateBuilder.SetStateID(*stateRequirement.StateID)
// 	}

// 	_, err := updateBuilder.Save(ctx)
// 	return err
// }

func (r *StateRequirementChecksRepository) DeleteStateRequirementCheck(ctx context.Context, ID int) error {
	err := r.client.StateRequirementCheck.DeleteOneID(ID).Exec(ctx)
	return err
}