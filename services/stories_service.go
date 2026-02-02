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

type StoriesService struct {
	client                   *ent.Client
	containerService         *ContainerService
	storiesRepository        *repositories.StoriesRepository
	childContainerRepository *ChildContainerRepository
}

func NewStoriesService(client *ent.Client, containerService *ContainerService, storiesRepository *repositories.StoriesRepository, childContainerRepository *ChildContainerRepository) *StoriesService {
	return &StoriesService{client: client, containerService: containerService, storiesRepository: storiesRepository, childContainerRepository: childContainerRepository}
}

func (s *StoriesService) GetStoryFull(ctx context.Context, ID int) (*models.StoryFull, error) {
	story, errProblem := s.storiesRepository.GetStory(ctx, ID)
	if errProblem != nil {
		return nil, errProblem
	}
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeStory, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeStory, ID)
	subquestions, errSubquestions := s.containerService.GetOpenQuestionsIDs(ctx, schema.ContainerTypeQuestion, ID)
	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeStory, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil || errSubquestions != nil {
		return nil, errors.New("story not found")
	}
	StoryFull := &models.StoryFull{
		ID:               ID,
		Description:      story.Description,
		Tags:             story.Tags,
		Notes:            story.Notes,
		Closed:           story.Closed,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        subquestions,
		Actions:          []int{},
		Definitions:      []int{},
		KnowledgeBits:    []int{},
		ParentContainers: parentContainers,
		KnowledgeNodes:   []int{},
		DoneDateTime:     story.DoneDateTime,
	}
	return StoryFull, nil
}

func (s *StoriesService) GetStoriesFull(ctx context.Context, IDs []int) ([]*models.StoryFull, error) {
	if len(IDs) == 0 {
		return []*models.StoryFull{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]*models.StoryFull, 0, len(IDs))
	var firstErr error

	for _, id := range IDs {
		wg.Add(1)
		go func(problemID int) {
			defer wg.Done()

			storyFull, err := s.GetStoryFull(ctx, problemID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			results = append(results, storyFull)
		}(id)
	}

	wg.Wait()

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *StoriesService) AddStory(ctx context.Context, problem models.StoryShort, parent *models.ContainerDescription) (*models.StoryFull, error) {
	newStory, err := s.storiesRepository.AddStory(ctx, problem.Description, problem.Tags, problem.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeStory, newStory.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetStoryFull(ctx, newStory.ID)
}

func (s *StoriesService) UpdateStory(ctx context.Context, storyPartial models.StoryPartial) (*models.StoryFull, error) {
	err := s.storiesRepository.UpdateStory(ctx, storyPartial)
	if err != nil {
		return nil, err
	}
	return s.GetStoryFull(ctx, storyPartial.ID)
}
