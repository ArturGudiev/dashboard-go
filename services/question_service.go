package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/problem"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// QuestionService handles question-related business logic
type QuestionService struct {
	client                   *ent.Client
	containerService         *ContainerService
	questionsRepository      *repositories.QuestionsRepository
	childContainerRepository *ChildContainerRepository
}

// NewQuestionService creates a new QuestionService
func NewQuestionService(client *ent.Client, containerService *ContainerService, questionsRepository *repositories.QuestionsRepository, childContainerRepository *ChildContainerRepository) *QuestionService {
	return &QuestionService{client: client, containerService: containerService, questionsRepository: questionsRepository, childContainerRepository: childContainerRepository}
}

// GetOpenDescendantQuestions recursively gets all descendant questions that are not done
func (s *QuestionService) GetOpenDescendantQuestions(ctx context.Context, parentQuestion *ent.Question) []*ent.Question {
	var result []*ent.Question

	// Get all child relationships where this question is the parent
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(schema.ContainerTypeQuestion),
			containerchild.ParentID(parentQuestion.ID),
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
		).
		All(ctx)

	if err != nil {
		return result
	}

	// Process each child - manually load problems since edges don't support Problem
	for _, relation := range childRelations {
		childQuestion, err := s.client.Question.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}

		// Only include problems that are not done (solution is null)
		if childQuestion.Answer == nil {
			result = append(result, childQuestion)
		}

		// Recursively get descendants of this child
		descendants := s.GetOpenDescendantQuestions(ctx, childQuestion)
		result = append(result, descendants...)
	}

	return result
}

func (s *QuestionService) GetParent(ctx context.Context, problem *ent.Problem) *ent.Problem { // TODO Check all
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
func (s *QuestionService) FinishQuestionRecursively(ctx context.Context, question *ent.Question) error {
	// Recursively get all descendant problems that are not done
	allProblemsToFinish := s.GetOpenDescendantQuestions(ctx, question)

	// Mark all problems as done with current timestamp
	now := time.Now()
	for _, problemToFinish := range allProblemsToFinish {
		_ = s.FinishQuestionRecursively(ctx, problemToFinish)
	}

	// Mark problem as done by setting solution to empty string if not already set
	updateBuilder := s.client.Problem.UpdateOneID(question.ID).
		SetDoneDateTime(now)

	if problem.Solution == nil {
		updateBuilder = updateBuilder.SetSolution("")
	}

	_, err := updateBuilder.Save(ctx)
	return err
}

// FinishProblemById finishes a problem and all its descendants by problem ID
func (s *QuestionService) FinishQuestionById(ctx context.Context, questionID int) (*ent.Question, error) {
	// Get the problem by ID
	question, err := s.client.Question.Get(ctx, questionID)
	if err != nil {
		return nil, err
	}

	// Finish problem recursively
	if err := s.FinishQuestionRecursively(ctx, question); err != nil {
		return nil, fmt.Errorf("failed to finish problem recursively: %v", err)
	}

	// Return the updated problem
	updatedQuestion, err := s.client.Question.Get(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated problem: %v", err)
	}

	return updatedQuestion, nil
}

func (s *QuestionService) GetQuestionFull(ctx context.Context, ID int) (*models.QuestionFull, error) {
	question, errProblem := s.questionsRepository.GetQuestion(ctx, ID)
	if errProblem != nil {
		return nil, errProblem
	}
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeProblem, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeProblem, ID)
	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeProblem, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil {
		return nil, errors.New("question not found")
	}
	QuestionFull := &models.QuestionFull{
		ID:               ID,
		Description:      question.Description,
		Tags:             question.Tags,
		Notes:            question.Notes,
		Answer:           question.Answer,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        []int{},
		Actions:          []int{},
		Definitions:      []int{},
		KnowledgeBits:    []int{},
		ParentContainers: parentContainers,
		KnowledgeNodes:   []int{},
		DoneDateTime:     question.DoneDateTime,
	}

	return QuestionFull, nil
}

func (s *QuestionService) GetQuestionsFull(ctx context.Context, IDs []int) ([]*models.QuestionFull, error) {
	if len(IDs) == 0 {
		return []*models.QuestionFull{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]*models.QuestionFull, 0, len(IDs))
	var firstErr error

	for _, id := range IDs {
		wg.Add(1)
		go func(problemID int) {
			defer wg.Done()

			questionFull, err := s.GetQuestionFull(ctx, problemID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			results = append(results, questionFull)
		}(id)
	}

	wg.Wait()

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *QuestionService) AddProblem(ctx context.Context, problem models.QuestionShort, parent *models.ContainerDescription) (*models.QuestionFull, error) {
	newProblem, err := s.questionsRepository.AddQuestion(ctx, problem.Description, problem.Tags, problem.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeProblem, newProblem.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetQuestionFull(ctx, newProblem.ID)
}

func (s *QuestionService) SolveProblem(ctx context.Context, problemID int, solution string) (*models.QuestionFull, error) {
	err := s.questionsRepository.AddSolution(ctx, problemID, solution)
	if err != nil {
		return nil, err
	}
	return s.GetQuestionFull(ctx, problemID)
}

func (s *QuestionService) UpdateProblem(ctx context.Context, questionPartial models.QuestionPartial) (*models.QuestionFull, error) {
	err := s.questionsRepository.UpdateQuestion(ctx, questionPartial)
	if err != nil {
		return nil, err
	}
	return s.GetQuestionFull(ctx, questionPartial.ID)
}
