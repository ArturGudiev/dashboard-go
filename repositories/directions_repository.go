package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/direction"
	"arturgudiev/dashboard/models"
	"context"

	"entgo.io/ent/dialect/sql"
)

type DirectionsRepository struct {
	client *ent.Client
}

func NewDirectionsRepository(client *ent.Client) *DirectionsRepository {
	return &DirectionsRepository{client: client}
}

func (r *DirectionsRepository) GetDirection(ctx context.Context, id int) (*ent.Direction, error) {
	return r.client.Direction.Get(ctx, id)
}

func (r *DirectionsRepository) GetDirections(ctx context.Context, open *bool) ([]*ent.Direction, error) {
	query := r.client.Direction.Query().Order(direction.ByID(sql.OrderDesc()))
	if open != nil && *open {
		query = query.Where(direction.ClosedEQ(false))
	}
	return query.All(ctx)
}

func (r *DirectionsRepository) AddDirection(
	ctx context.Context,
	description string,
	tags []string,
	notes string,
) (*ent.Direction, error) {
	return r.client.Direction.Create().
		SetDescription(description).
		SetTags(tags).
		SetNotes(notes).
		Save(ctx)
}

func (r *DirectionsRepository) UpdateDirection(ctx context.Context, partial models.DirectionPartial) error {
	updateBuilder := r.client.Direction.UpdateOneID(partial.ID)
	if partial.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*partial.Description)
	}
	if partial.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*partial.Notes)
	}
	if partial.Closed != nil {
		updateBuilder = updateBuilder.SetClosed(*partial.Closed)
	}
	_, err := updateBuilder.Save(ctx)
	return err
}
