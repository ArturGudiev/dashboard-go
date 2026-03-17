package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ProblemService handles problem-related business logic
type ProblemService struct {
	client                   *ent.Client
	containerService         *ContainerService
	problemsRepository       *ProblemsRepository
	childContainerRepository *ChildContainerRepository
}

// NewProblemService creates a new ProblemService
func NewProblemService(client *ent.Client, containerService *ContainerService, problemsRepository *ProblemsRepository, childContainerRepository *ChildContainerRepository) *ProblemService {
	return &ProblemService{client: client, containerService: containerService, problemsRepository: problemsRepository, childContainerRepository: childContainerRepository}
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

	_, err := updateBuilder.Save(ctx)
	return err
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

func (s *ProblemService) GetProblemFull(ctx context.Context, ID int) (*models.ProblemFull, error) {
	problem, errProblem := s.problemsRepository.GetProblem(ctx, ID)
	if errProblem != nil {
		return nil, errProblem
	}
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeProblem, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeProblem, ID)
	subquestions, errSubquestions := s.containerService.GetOpenQuestionsIDs(ctx, schema.ContainerTypeProblem, ID)
	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeProblem, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil || errSubquestions != nil {
		return nil, errors.New("problem not found")
	}
	ProblemFull := &models.ProblemFull{
		ID:               ID,
		Description:      problem.Description,
		Tags:             problem.Tags,
		Notes:            problem.Notes,
		Solution:         problem.Solution,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        subquestions,
		Actions:          []int{},
		Definitions:      []int{},
		KnowledgeBits:    []int{},
		ParentContainers: parentContainers,
		KnowledgeNodes:   []int{},
		DoneDateTime:     problem.DoneDateTime,
	}

	return ProblemFull, nil
}

func (s *ProblemService) GetProblemsFull(ctx context.Context, IDs []int) ([]*models.ProblemFull, error) {
	if len(IDs) == 0 {
		return []*models.ProblemFull{}, nil
	}

	var wg sync.WaitGroup
	var firstErr error
	var firstErrOnce sync.Once
	byIndex := make([]*models.ProblemFull, len(IDs))

	for i, id := range IDs {
		wg.Add(1)
		go func(idx int, problemID int) {
			defer wg.Done()

			problemFull, err := s.GetProblemFull(ctx, problemID)

			if err != nil {
				firstErrOnce.Do(func() { firstErr = err })
				return
			}

			byIndex[idx] = problemFull
		}(i, id)
	}

	wg.Wait()

	results := make([]*models.ProblemFull, 0, len(IDs))
	for _, problemFull := range byIndex {
		if problemFull != nil {
			results = append(results, problemFull)
		}
	}

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *ProblemService) AddProblem(ctx context.Context, problem models.ProblemShort, parent *models.ContainerDescription) (*models.ProblemFull, error) {
	newProblem, err := s.problemsRepository.AddProblem(ctx, problem.Description, problem.Tags, problem.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeProblem, newProblem.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetProblemFull(ctx, newProblem.ID)
}

func (s *ProblemService) SolveProblem(ctx context.Context, problemID int, solution string) (*models.ProblemFull, error) {
	err := s.problemsRepository.AddSolution(ctx, problemID, solution)
	if err != nil {
		return nil, err
	}
	return s.GetProblemFull(ctx, problemID)
}

func (s *ProblemService) UpdateProblem(ctx context.Context, problemPartial models.ProblemPartial) (*models.ProblemFull, error) {
	err := s.problemsRepository.UpdateProblem(ctx, problemPartial)
	if err != nil {
		return nil, err
	}
	return s.GetProblemFull(ctx, problemPartial.ID)
}
