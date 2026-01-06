package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
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

func (r *ProblemsRepository) UpdateProblem(ctx context.Context, problem models.ProblemPartial) error {
	updateBuilder := r.client.Problem.UpdateOneID(problem.ID)

	if problem.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*problem.Description)
	}

	if problem.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*problem.Notes)
	}

	if problem.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*problem.Tags)
	}

	if problem.Solution != nil {
		updateBuilder = updateBuilder.SetSolution(*problem.Solution)
	}

	if problem.DoneDateTime != nil {
		updateBuilder = updateBuilder.SetDoneDateTime(*problem.DoneDateTime)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}
