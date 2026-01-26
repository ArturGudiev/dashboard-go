package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"errors"
	"sync"
)

type EpicsService struct {
	client                   *ent.Client
	containerService         *ContainerService
	epicsRepository          *repositories.EpicsRepository
	childContainerRepository *ChildContainerRepository
}

func NewEpicsService(client *ent.Client, containerService *ContainerService, epicsRepository *repositories.EpicsRepository, childContainerRepository *ChildContainerRepository) *EpicsService {
	return &EpicsService{client: client, containerService: containerService, epicsRepository: epicsRepository, childContainerRepository: childContainerRepository}
}

func (s *EpicsService) GetEpicFull(ctx context.Context, ID int) (*models.EpicFull, error) {
	epic, errProblem := s.epicsRepository.GetEpic(ctx, ID)
	if errProblem != nil {
		return nil, errProblem
	}
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeEpic, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeEpic, ID)
	stories, errStories := s.containerService.GetOpenStoriesIDs(ctx, schema.ContainerTypeEpic, ID)

	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeEpic, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil || errStories != nil {
		return nil, errors.New("epic not found")
	}
	EpicFull := &models.EpicFull{
		ID:               ID,
		Description:      epic.Description,
		Tags:             epic.Tags,
		Notes:            epic.Notes,
		Closed:           epic.Closed,
		Epics:            []int{},
		Stories:          stories,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        []int{},
		Actions:          []int{},
		Definitions:      []int{},
		KnowledgeBits:    []int{},
		ParentContainers: parentContainers,
		KnowledgeNodes:   []int{},
		DoneDateTime:     epic.DoneDateTime,
	}
	return EpicFull, nil
}

func (s *EpicsService) GetEpicsFull(ctx context.Context, IDs []int) ([]*models.EpicFull, error) {
	if len(IDs) == 0 {
		return []*models.EpicFull{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]*models.EpicFull, 0, len(IDs))
	var firstErr error

	for _, id := range IDs {
		wg.Add(1)
		go func(problemID int) {
			defer wg.Done()

			epicFull, err := s.GetEpicFull(ctx, problemID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			results = append(results, epicFull)
		}(id)
	}

	wg.Wait()

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *EpicsService) GetAllOpenEpicsFull(ctx context.Context) ([]*models.EpicFull, error) {

	allOpenEpics, err := s.epicsRepository.GetAllOpenEpics(ctx)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]*models.EpicFull, 0, len(allOpenEpics))
	var firstErr error

	for _, epic := range allOpenEpics {
		wg.Add(1)
		go func(epic *ent.Epic) {
			defer wg.Done()
			epicFull, err := s.GetEpicFull(ctx, epic.ID)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			mu.Lock()
			defer mu.Unlock()
			results = append(results, epicFull)
		}(epic)
	}

	wg.Wait()

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *EpicsService) AddEpic(ctx context.Context, problem models.EpicShort, parent *models.ContainerDescription) (*models.EpicFull, error) {
	newEpic, err := s.epicsRepository.AddEpic(ctx, problem.Description, problem.Tags, problem.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeEpic, newEpic.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetEpicFull(ctx, newEpic.ID)
}

func (s *EpicsService) UpdateEpic(ctx context.Context, epicPartial models.EpicPartial) (*models.EpicFull, error) {
	err := s.epicsRepository.UpdateEpic(ctx, epicPartial)
	if err != nil {
		return nil, err
	}
	return s.GetEpicFull(ctx, epicPartial.ID)
}
