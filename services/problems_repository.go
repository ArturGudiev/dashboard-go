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

func (r *ProblemsRepository) GetProblem(ctx context.Context, ID int) (*ent.Problem, error) {
	problem, err := r.client.Problem.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return problem, nil
}

func (r *ProblemsRepository) AddProblem(ctx context.Context, description string, tags []string, notes string) (*ent.Problem, error) {
	problem, err := r.client.Problem.Create().SetDescription(description).SetTags(tags).SetNotes(notes).Save(ctx)

	if err != nil {
		return nil, err
	}
	return problem, nil
}
