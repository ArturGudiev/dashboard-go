package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"fmt"
	"time"
)

// ProblemService handles problem-related business logic
type ProblemService struct {
	client           *ent.Client
	containerService *ContainerService
}

// NewProblemService creates a new ProblemService
func NewProblemService(client *ent.Client, containerService *ContainerService) *ProblemService {
	return &ProblemService{client: client, containerService: containerService}
}

// GetOpenDescendantProblems recursively gets all descendant problems that are not done
func (s *ProblemService) GetOpenDescendantProblems(ctx context.Context, parentProblem *ent.Problem) []*ent.Problem {
	var result []*ent.Problem

	// Get all child relationships where this problem is the parent
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(schema.ContainerTypeProblem),
			containerchild.ParentID(parentProblem.ID),
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
		).
		All(ctx)

	if err != nil {
		return result
	}

	// Process each child - manually load problems since edges don't support Problem
	for _, relation := range childRelations {
		childProblem, err := s.client.Problem.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}

		// Only include problems that are not done (solution is null)
		if childProblem.Solution == nil {
			result = append(result, childProblem)
		}

		// Recursively get descendants of this child
		descendants := s.GetOpenDescendantProblems(ctx, childProblem)
		result = append(result, descendants...)
	}

	return result
}

func (s *ProblemService) GetParent(ctx context.Context, problem *ent.Problem) *ent.Problem {
	// Get all parent relationships where this problem is the child
	parentRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
			containerchild.ChildID(problem.ID),
		).
		All(ctx)

	if err != nil || len(parentRelations) == 0 {
		return nil
	}

	// Get the parent problem from the first relation
	// Since edges don't support Problem, we need to check parent type and load accordingly
	parentRelation := parentRelations[0]
	var parentProblem *ent.Problem

	if parentRelation.ParentType == schema.ContainerTypeProblem {
		parentProblem, err = s.client.Problem.Get(ctx, parentRelation.ParentID)
		if err != nil {
			return nil
		}
		return parentProblem
	}

	// If parent is not a problem, return nil (could be a task or other type)
	return nil
}

// FinishProblemRecursively finishes all open descendant problems of the given problem
func (s *ProblemService) FinishProblemRecursively(ctx context.Context, problem *ent.Problem) error {
	// Recursively get all descendant problems that are not done
	allProblemsToFinish := s.GetOpenDescendantProblems(ctx, problem)

	// Mark all problems as done with current timestamp
	now := time.Now()
	for _, problemToFinish := range allProblemsToFinish {
		_ = s.FinishProblemRecursively(ctx, problemToFinish)
	}

	// Mark problem as done by setting solution to empty string if not already set
	updateBuilder := s.client.Problem.UpdateOneID(problem.ID).
		SetDoneDateTime(now)

	if problem.Solution == nil {
		updateBuilder = updateBuilder.SetSolution("")
	}

	updateBuilder.Save(ctx)

	return nil
}

// FinishProblemById finishes a problem and all its descendants by problem ID
func (s *ProblemService) FinishProblemById(ctx context.Context, problemID int) (*ent.Problem, error) {
	// Get the problem by ID
	problem, err := s.client.Problem.Get(ctx, problemID)
	if err != nil {
		return nil, err
	}

	// Finish problem recursively
	if err := s.FinishProblemRecursively(ctx, problem); err != nil {
		return nil, fmt.Errorf("failed to finish problem recursively: %v", err)
	}

	// Return the updated problem
	updatedProblem, err := s.client.Problem.Get(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated problem: %v", err)
	}

	return updatedProblem, nil
}

// GetChildSubproblems returns all child problems for a given parent problem ID
func (s *ProblemService) GetChildSubproblems(ctx context.Context, parentID int) ([]*ent.Problem, error) {
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(schema.ContainerTypeProblem),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
		).
		Order(containerchild.ByChildOrder()).
		All(ctx)

	if err != nil {
		return nil, err
	}

	var childProblems []*ent.Problem
	for _, relation := range childRelations {
		childProblem, err := s.client.Problem.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}
		childProblems = append(childProblems, childProblem)
	}

	return childProblems, nil
}

// AddSubproblem creates a new subproblem for the given parent problem
func (s *ProblemService) AddSubproblem(ctx context.Context, parentID int, description string) (*ent.Problem, error) {
	// Create the new problem (not done - solution is null by default)
	newProblem, err := s.client.Problem.Create().
		SetDescription(description).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create problem: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(schema.ContainerTypeProblem),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count children: %v", err)
	}

	// Get the count of existing parents to set parent_order
	parentCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
			containerchild.ChildID(newProblem.ID),
			containerchild.ParentTypeEQ(schema.ContainerTypeProblem),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count parents: %v", err)
	}

	// Create the parent-child relationship
	_, err = s.client.ContainerChild.Create().
		SetParentType(schema.ContainerTypeProblem).
		SetParentID(parentID).
		SetChildType(schema.ContainerTypeProblem).
		SetChildID(newProblem.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %v", err)
	}

	return newProblem, nil
}
