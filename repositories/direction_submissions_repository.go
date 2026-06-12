package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/directionsubmission"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
)

type DirectionSubmissionsRepository struct {
	client *ent.Client
}

func NewDirectionSubmissionsRepository(client *ent.Client) *DirectionSubmissionsRepository {
	return &DirectionSubmissionsRepository{client: client}
}

func (r *DirectionSubmissionsRepository) AddDirectionSubmission(
	ctx context.Context,
	directionID int,
	text *string,
) (*ent.DirectionSubmission, error) {
	createBuilder := r.client.DirectionSubmission.Create().
		SetDirectionID(directionID).
		SetExecutionDate(time.Now())
	if text != nil {
		createBuilder = createBuilder.SetNillableText(text)
	}
	return createBuilder.Save(ctx)
}

func (r *DirectionSubmissionsRepository) GetDirectionSubmissions(
	ctx context.Context,
	directionID int,
) ([]*ent.DirectionSubmission, error) {
	return r.client.DirectionSubmission.Query().
		Where(directionsubmission.DirectionIDEQ(directionID)).
		Order(directionsubmission.ByExecutionDate(sql.OrderDesc())).
		All(ctx)
}
