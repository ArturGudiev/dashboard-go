package services

import (
	"arturgudiev/dashboard/ent"
	"context"
)

type ProblemsRepository struct {
	client *ent.Client
}

func NewProblemsRepository(client *ent.Client) *ProblemsRepository {
	return &ProblemsRepository{client: client}
}

func (r *ProblemsRepository) AddSolution(ctx context.Context, ID int, solution string) error {
	updateBuilder := r.client.Problem.UpdateOneID(ID).
		SetSolution(solution)

	_, err := updateBuilder.Save(ctx)
	if err != nil {
		return err
	}
	return nil
}
