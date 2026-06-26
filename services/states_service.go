package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"time"

	"arturgudiev/dashboard/ent/schema"
)

type StatesService struct {
	client *ent.Client
	containerService *ContainerService
	statesRepository *repositories.StatesRepository
	stateRequirementsRepository *repositories.StateRequirementsRepository
	stateRequirementChecksRepository *repositories.StateRequirementChecksRepository
	childContainerRepository *ChildContainerRepository
}

func NewStatesService(client *ent.Client, containerService *ContainerService, statesRepository *repositories.StatesRepository, stateRequirementsRepository *repositories.StateRequirementsRepository, stateRequirementChecksRepository *repositories.StateRequirementChecksRepository, childContainerRepository *ChildContainerRepository) *StatesService {
	return &StatesService{client: client, containerService: containerService, statesRepository: statesRepository, stateRequirementsRepository: stateRequirementsRepository, stateRequirementChecksRepository: stateRequirementChecksRepository, childContainerRepository: childContainerRepository}
}

func (s *StatesService) GetStateFull(ctx context.Context, ID int) (*models.StateFull, error) {
	state, err := s.statesRepository.GetState(ctx, ID)
	if err != nil {
		return nil, err
	}

	stateRequirements, err := s.stateRequirementsRepository.GetStateRequirementsByStateID(ctx, state.ID)
	if err != nil {
		return nil, err
	}

	requirementIDs := make([]int, len(stateRequirements))
	requirementStatuses := make([]*bool, len(stateRequirements))
	now := time.Now()

	for i, requirement := range stateRequirements {
		requirementIDs[i] = requirement.ID
		requirementStatuses[i], err = s.getRequirementIsFulfilled(ctx, requirement, now)
		if err != nil {
			return nil, err
		}
	}

	childStates, err := s.containerService.GetOpenStatesIDs(ctx, schema.ContainerTypeState, state.ID)
	if err != nil {
		return nil, err
	}

	stateFull := &models.StateFull{
		ID:                state.ID,
		Description:       state.Description,
		Tags:              state.Tags,
		Notes:             state.Notes,
		Closed:            state.Closed,
		States:            childStates,
		StateRequirements: requirementIDs,
		IsFulfilled:       ComputeStateIsFulfilled(requirementStatuses),
	}

	return stateFull, nil
}


func (s *StatesService) GetAllStatesFull(ctx context.Context) ([]*models.StateFull, error) {
	states, err := s.statesRepository.GetAllStates(ctx)
	if err != nil {
		return nil, err
	}

	statesFull := make([]*models.StateFull, len(states))
	for i, state := range states {
		stateFull, err := s.GetStateFull(ctx, state.ID)
		if err != nil {
			return nil, err
		}
		statesFull[i] = stateFull
	}
	return statesFull, nil
}

func (s *StatesService) AddState(ctx context.Context, state models.StateShort, parent *models.ContainerDescription) (*models.StateFull, error) {
	newState, err := s.statesRepository.AddState(ctx, state.Description, state.Tags, state.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeState, newState.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetStateFull(ctx, newState.ID)
}

func (s *StatesService) AddStateRequirement(ctx context.Context, stateID int, description string, onceInDays *int) (*models.StateRequirementFull, error) {
	newStateRequirement, err := s.stateRequirementsRepository.AddStateRequirement(ctx, description, stateID, onceInDays)
	if err != nil {
		return nil, err
	}

	return s.buildStateRequirementFull(ctx, newStateRequirement, time.Now())
}

func (s *StatesService) GetStateRequirementsFullByStateID(ctx context.Context, stateID int) ([]*models.StateRequirementFull, error) {
	stateRequirements, err := s.stateRequirementsRepository.GetStateRequirementsByStateID(ctx, stateID)
	if err != nil {
		return nil, err
	}

	return s.buildStateRequirementsFull(ctx, stateRequirements)
}

func (s *StatesService) GetStateRequirementsFullByIDs(ctx context.Context, ids []int) ([]*models.StateRequirementFull, error) {
	if len(ids) == 0 {
		return []*models.StateRequirementFull{}, nil
	}

	stateRequirements, err := s.stateRequirementsRepository.GetStateRequirementsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return s.buildStateRequirementsFull(ctx, stateRequirements)
}

func (s *StatesService) buildStateRequirementsFull(ctx context.Context, stateRequirements []*ent.StateRequirement) ([]*models.StateRequirementFull, error) {
	now := time.Now()
	result := make([]*models.StateRequirementFull, len(stateRequirements))

	for i, requirement := range stateRequirements {
		full, err := s.buildStateRequirementFull(ctx, requirement, now)
		if err != nil {
			return nil, err
		}
		result[i] = full
	}

	return result, nil
}

func (s *StatesService) buildStateRequirementFull(ctx context.Context, requirement *ent.StateRequirement, now time.Time) (*models.StateRequirementFull, error) {
	isFulfilled, err := s.getRequirementIsFulfilled(ctx, requirement, now)
	if err != nil {
		return nil, err
	}

	return &models.StateRequirementFull{
		ID:          requirement.ID,
		Description: requirement.Description,
		StateID:     requirement.StateID,
		OnceInDays:  requirement.OnceInDays,
		IsFulfilled: isFulfilled,
	}, nil
}

func (s *StatesService) getRequirementIsFulfilled(ctx context.Context, requirement *ent.StateRequirement, now time.Time) (*bool, error) {
	latestCheck, err := s.stateRequirementChecksRepository.GetLatestStateRequirementCheck(ctx, requirement.ID)
	if err != nil {
		return nil, err
	}

	return ComputeRequirementIsFulfilled(requirement, latestCheck, now), nil
}


// func (s *EpicsService) GetEpicFull(ctx context.Context, ID int) (*models.EpicFull, error) {
// 	epic, errProblem := s.epicsRepository.GetEpic(ctx, ID)
// 	if errProblem != nil {
// 		return nil, errProblem
// 	}
// 	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeEpic, ID)
// 	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeEpic, ID)
// 	subquestions, errSubquestions := s.containerService.GetOpenQuestionsIDs(ctx, schema.ContainerTypeEpic, ID)
// 	longTasks, errLongTasks := s.containerService.GetOpenLongTasksIDs(ctx, schema.ContainerTypeEpic, ID)
// 	stories, errStories := s.containerService.GetOpenStoriesIDs(ctx, schema.ContainerTypeEpic, ID)
// 	epics, errEpics := s.containerService.GetOpenEpicsIDs(ctx, schema.ContainerTypeEpic, ID)

// 	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeEpic, ID)
// 	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil || errStories != nil || errSubquestions != nil || errEpics != nil || errLongTasks != nil {
// 		return nil, errors.New("epic not found")
// 	}
// 	EpicFull := &models.EpicFull{
// 		ID:               ID,
// 		Description:      epic.Description,
// 		Tags:             epic.Tags,
// 		Notes:            epic.Notes,
// 		Closed:           epic.Closed,
// 		Epics:            epics,
// 		Stories:          stories,
// 		Tasks:            subtasks,
// 		Problems:         subproblems,
// 		Questions:        subquestions,
// 		LongTasks:        longTasks,
// 		Actions:          []int{},
// 		Definitions:      []int{},
// 		KnowledgeBits:    []int{},
// 		ParentContainers: parentContainers,
// 		KnowledgeNodes:   []int{},
// 		DoneDateTime:     epic.DoneDateTime,
// 	}
// 	return EpicFull, nil
// }

// func (s *EpicsService) GetEpicsFull(ctx context.Context, IDs []int) ([]*models.EpicFull, error) {
// 	if len(IDs) == 0 {
// 		return []*models.EpicFull{}, nil
// 	}

// 	var wg sync.WaitGroup
// 	var firstErr error
// 	var firstErrOnce sync.Once
// 	byIndex := make([]*models.EpicFull, len(IDs))

// 	for i, id := range IDs {
// 		wg.Add(1)
// 		go func(idx int, problemID int) {
// 			defer wg.Done()

// 			epicFull, err := s.GetEpicFull(ctx, problemID)

// 			if err != nil {
// 				firstErrOnce.Do(func() { firstErr = err })
// 				return
// 			}

// 			byIndex[idx] = epicFull
// 		}(i, id)
// 	}

// 	wg.Wait()

// 	results := make([]*models.EpicFull, 0, len(IDs))
// 	for _, epicFull := range byIndex {
// 		if epicFull != nil {
// 			results = append(results, epicFull)
// 		}
// 	}

// 	if firstErr != nil && len(results) == 0 {
// 		return nil, firstErr
// 	}

// 	return results, nil
// }

// func (s *EpicsService) GetAllOpenEpicsFull(ctx context.Context) ([]*models.EpicFull, error) {

// 	allOpenEpics, err := s.epicsRepository.GetAllOpenEpics(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var wg sync.WaitGroup
// 	var firstErr error
// 	var firstErrOnce sync.Once
// 	byIndex := make([]*models.EpicFull, len(allOpenEpics))

// 	for i, epic := range allOpenEpics {
// 		wg.Add(1)
// 		go func(idx int, epic *ent.Epic) {
// 			defer wg.Done()
// 			epicFull, err := s.GetEpicFull(ctx, epic.ID)
// 			if err != nil {
// 				firstErrOnce.Do(func() { firstErr = err })
// 				return
// 			}
// 			byIndex[idx] = epicFull
// 		}(i, epic)
// 	}

// 	wg.Wait()

// 	results := make([]*models.EpicFull, 0, len(allOpenEpics))
// 	for _, epicFull := range byIndex {
// 		if epicFull != nil {
// 			results = append(results, epicFull)
// 		}
// 	}

// 	if firstErr != nil && len(results) == 0 {
// 		return nil, firstErr
// 	}

// 	return results, firstErr
// }

// func (s *EpicsService) AddEpic(ctx context.Context, problem models.EpicShort, parent *models.ContainerDescription) (*models.EpicFull, error) {
// 	newEpic, err := s.epicsRepository.AddEpic(ctx, problem.Description, problem.Tags, problem.Notes)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if parent != nil {
// 		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeEpic, newEpic.ID)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return s.GetEpicFull(ctx, newEpic.ID)
// }

// func (s *EpicsService) UpdateEpic(ctx context.Context, epicPartial models.EpicPartial) (*models.EpicFull, error) {
// 	err := s.epicsRepository.UpdateEpic(ctx, epicPartial)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s.GetEpicFull(ctx, epicPartial.ID)
// }
