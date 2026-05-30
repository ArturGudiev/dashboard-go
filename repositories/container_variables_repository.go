package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containervariables"
	"arturgudiev/dashboard/ent/schema"
	"context"
)

type ContainerVariablesRepository struct {
	client    *ent.Client
	stackRepo *VariablesStackRepository
}

func NewContainerVariablesRepository(client *ent.Client, stackRepo *VariablesStackRepository) *ContainerVariablesRepository {
	return &ContainerVariablesRepository{
		client:    client,
		stackRepo: stackRepo,
	}
}

func (r *ContainerVariablesRepository) AddVariableWithValue(
	ctx context.Context,
	containerType schema.ContainerType,
	containerID int,
	variableName string,
	variableValue string,
) (*ent.ContainerVariables, error) {
	stack, err := r.stackRepo.GetOrCreateVariablesStack(ctx, containerType, containerID)
	if err != nil {
		return nil, err
	}

	return r.client.ContainerVariables.Create().
		SetVariablesStackID(stack.ID).
		SetVariableName(variableName).
		SetVariableValue(variableValue).
		Save(ctx)
}

func (r *ContainerVariablesRepository) RemoveVariable(ctx context.Context, id int) error {
	return r.client.ContainerVariables.DeleteOneID(id).Exec(ctx)
}

func (r *ContainerVariablesRepository) UpdateVariable(
	ctx context.Context,
	id int,
	name, value *string,
) (*ent.ContainerVariables, error) {
	updateBuilder := r.client.ContainerVariables.UpdateOneID(id)
	if name != nil {
		updateBuilder = updateBuilder.SetVariableName(*name)
	}
	if value != nil {
		updateBuilder = updateBuilder.SetVariableValue(*value)
	}
	return updateBuilder.Save(ctx)
}

func (r *ContainerVariablesRepository) GetVariablesByContainer(
	ctx context.Context,
	containerType schema.ContainerType,
	containerID int,
) ([]*ent.ContainerVariables, error) {
	stack, err := r.stackRepo.GetVariablesStackByContainer(ctx, containerType, containerID)
	if err != nil {
		if ent.IsNotFound(err) {
			return []*ent.ContainerVariables{}, nil
		}
		return nil, err
	}

	return r.client.ContainerVariables.Query().
		Where(containervariables.VariablesStackIDEQ(stack.ID)).
		Order(ent.Asc(containervariables.FieldVariableName)).
		All(ctx)
}
