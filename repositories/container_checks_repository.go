package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containercheck"
	"arturgudiev/dashboard/ent/schema"
	"context"
)

type ContainerChecksRepository struct {
	client *ent.Client
}

func NewContainerChecksRepository(client *ent.Client) *ContainerChecksRepository {
	return &ContainerChecksRepository{client: client}
}

func (r *ContainerChecksRepository) AddCheck(
	ctx context.Context,
	description string,
	containerType schema.ContainerType,
	containerID int,
) (*ent.ContainerCheck, error) {
	return r.client.ContainerCheck.Create().
		SetDescription(description).
		SetContainerType(containerType).
		SetContainerID(containerID).
		Save(ctx)
}

func (r *ContainerChecksRepository) RemoveCheck(ctx context.Context, id int) error {
	return r.client.ContainerCheck.DeleteOneID(id).Exec(ctx)
}

func (r *ContainerChecksRepository) UpdateCheck(
	ctx context.Context,
	id int,
	description string,
) (*ent.ContainerCheck, error) {
	return r.client.ContainerCheck.UpdateOneID(id).
		SetDescription(description).
		Save(ctx)
}

func (r *ContainerChecksRepository) GetCheck(ctx context.Context, id int) (*ent.ContainerCheck, error) {
	return r.client.ContainerCheck.Get(ctx, id)
}

func (r *ContainerChecksRepository) GetChecksByContainer(
	ctx context.Context,
	containerType schema.ContainerType,
	containerID int,
) ([]*ent.ContainerCheck, error) {
	return r.client.ContainerCheck.Query().
		Where(
			containercheck.ContainerTypeEQ(containerType),
			containercheck.ContainerIDEQ(containerID),
		).
		Order(ent.Asc(containercheck.FieldID)).
		All(ctx)
}
