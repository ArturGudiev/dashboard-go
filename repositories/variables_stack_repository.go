package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/ent/variablesstack"
	"context"
)

type VariablesStackRepository struct {
	client *ent.Client
}

func NewVariablesStackRepository(client *ent.Client) *VariablesStackRepository {
	return &VariablesStackRepository{client: client}
}

func (r *VariablesStackRepository) CreateVariablesStack(ctx context.Context, containerType schema.ContainerType, containerID int) (*ent.VariablesStack, error) {
	return r.client.VariablesStack.Create().
		SetContainerType(containerType).
		SetContainerID(containerID).
		Save(ctx)
}

func (r *VariablesStackRepository) GetVariablesStackByContainer(ctx context.Context, containerType schema.ContainerType, containerID int) (*ent.VariablesStack, error) {
	return r.client.VariablesStack.Query().
		Where(
			variablesstack.ContainerTypeEQ(containerType),
			variablesstack.ContainerIDEQ(containerID),
		).
		Only(ctx)
}

func (r *VariablesStackRepository) GetOrCreateVariablesStack(ctx context.Context, containerType schema.ContainerType, containerID int) (*ent.VariablesStack, error) {
	stack, err := r.GetVariablesStackByContainer(ctx, containerType, containerID)
	if err == nil {
		return stack, nil
	}
	if ent.IsNotFound(err) {
		return r.CreateVariablesStack(ctx, containerType, containerID)
	}
	return nil, err
}
