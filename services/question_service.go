package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
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

func NewQuestionService(client *ent.Client, containerService *ContainerService, questionsRepository *repositories.QuestionsRepository, childContainerRepository *ChildContainerRepository) *QuestionService {
	return &QuestionService{client: client, containerService: containerService, questionsRepository: questionsRepository, childContainerRepository: childContainerRepository}
}

func (s *QuestionService) GetOpenDescendantQuestions(ctx context.Context, parentQuestion *ent.Question) []*ent.Question {
	var result []*ent.Question

	// Get all child relationships where this question is the parent
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(schema.ContainerTypeQuestion),
			containerchild.ParentID(parentQuestion.ID),
			containerchild.ChildTypeEQ(schema.ContainerTypeQuestion),
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

func (s *QuestionService) FinishQuestionRecursively(ctx context.Context, question *ent.Question) error {
	// Recursively get all descendant problems that are not done
	allProblemsToFinish := s.GetOpenDescendantQuestions(ctx, question)

	// Mark all problems as done with current timestamp
	now := time.Now()
	for _, problemToFinish := range allProblemsToFinish {
		_ = s.FinishQuestionRecursively(ctx, problemToFinish)
	}

	updateBuilder := s.client.Question.UpdateOneID(question.ID).
		SetDoneDateTime(now)

	if question.Answer == nil {
		updateBuilder = updateBuilder.SetAnswer("")
	}

	_, err := updateBuilder.Save(ctx)
	return err
}

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
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeQuestion, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeQuestion, ID)
	subquestions, errSubquestions := s.containerService.GetOpenQuestionsIDs(ctx, schema.ContainerTypeQuestion, ID)
	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeQuestion, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil || errSubquestions != nil {
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
		Questions:        subquestions,
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
	var firstErr error
	var firstErrOnce sync.Once
	byIndex := make([]*models.QuestionFull, len(IDs))

	for i, id := range IDs {
		wg.Add(1)
		go func(idx int, problemID int) {
			defer wg.Done()

			questionFull, err := s.GetQuestionFull(ctx, problemID)

			if err != nil {
				firstErrOnce.Do(func() { firstErr = err })
				return
			}

			byIndex[idx] = questionFull
		}(i, id)
	}

	wg.Wait()

	results := make([]*models.QuestionFull, 0, len(IDs))
	for _, questionFull := range byIndex {
		if questionFull != nil {
			results = append(results, questionFull)
		}
	}

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *QuestionService) AddQuestion(ctx context.Context, problem models.QuestionShort, parent *models.ContainerDescription) (*models.QuestionFull, error) {
	newQuestion, err := s.questionsRepository.AddQuestion(ctx, problem.Description, problem.Tags, problem.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeQuestion, newQuestion.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetQuestionFull(ctx, newQuestion.ID)
}

func (s *QuestionService) AnswerQuestion(ctx context.Context, questionID int, answer string) (*models.QuestionFull, error) {
	err := s.questionsRepository.AddAnswer(ctx, questionID, answer)
	if err != nil {
		return nil, err
	}
	return s.GetQuestionFull(ctx, questionID)
}

func (s *QuestionService) UpdateQuestion(ctx context.Context, questionPartial models.QuestionPartial) (*models.QuestionFull, error) {
	err := s.questionsRepository.UpdateQuestion(ctx, questionPartial)
	if err != nil {
		return nil, err
	}
	return s.GetQuestionFull(ctx, questionPartial.ID)
}
