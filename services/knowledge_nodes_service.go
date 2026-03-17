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

type KnowledgeNodesService struct {
	client                   *ent.Client
	containerService         *ContainerService
	knowledgeNodesRepository *repositories.KnowledgeNodesRepository
	childContainerRepository *ChildContainerRepository
}

func NewKnowledgeNodesService(client *ent.Client, containerService *ContainerService, knowledgeNodesRepository *repositories.KnowledgeNodesRepository, childContainerRepository *ChildContainerRepository) *KnowledgeNodesService {
	return &KnowledgeNodesService{client: client, containerService: containerService, knowledgeNodesRepository: knowledgeNodesRepository, childContainerRepository: childContainerRepository}
}

func (s *KnowledgeNodesService) GetKnowledgeNodeFull(ctx context.Context, ID int) (*models.KnowledgeNodeFull, error) {
	knowledgeNode, errProblem := s.knowledgeNodesRepository.GetKnowledgeNode(ctx, ID)
	if errProblem != nil {
		return nil, errProblem
	}
	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeKnowledgeNode, ID)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeKnowledgeNode, ID)
	knowledgeNodes, errKnowledgeNodes := s.containerService.GetChildKnowledgeNodesIDs(ctx, schema.ContainerTypeKnowledgeNode, ID)
	parentContainers, errParentContainers := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeKnowledgeNode, ID)
	if errSubtasks != nil || errParentContainers != nil || errSubproblems != nil || errKnowledgeNodes != nil {
		return nil, errors.New("knowledgeNode not found")
	}
	KnowledgeNodeFull := &models.KnowledgeNodeFull{
		ID:               ID,
		Name:             knowledgeNode.Name,
		Tags:             knowledgeNode.Tags,
		Notes:            knowledgeNode.Notes,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        []int{},
		Actions:          []int{},
		Definitions:      []int{},
		KnowledgeBits:    []int{},
		ParentContainers: parentContainers,
		KnowledgeNodes:   knowledgeNodes,
	}
	return KnowledgeNodeFull, nil
}

func (s *KnowledgeNodesService) GetKnowledgeNodesFull(ctx context.Context, IDs []int) ([]*models.KnowledgeNodeFull, error) {
	if len(IDs) == 0 {
		return []*models.KnowledgeNodeFull{}, nil
	}

	var wg sync.WaitGroup
	var firstErr error
	var firstErrOnce sync.Once
	byIndex := make([]*models.KnowledgeNodeFull, len(IDs))

	for i, id := range IDs {
		wg.Add(1)
		go func(idx int, problemID int) {
			defer wg.Done()

			knowledgeNodeFull, err := s.GetKnowledgeNodeFull(ctx, problemID)

			if err != nil {
				firstErrOnce.Do(func() { firstErr = err })
				return
			}

			byIndex[idx] = knowledgeNodeFull
		}(i, id)
	}

	wg.Wait()

	results := make([]*models.KnowledgeNodeFull, 0, len(IDs))
	for _, knowledgeNodeFull := range byIndex {
		if knowledgeNodeFull != nil {
			results = append(results, knowledgeNodeFull)
		}
	}

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

func (s *KnowledgeNodesService) AddKnowledgeNode(ctx context.Context, knowledgeNode models.KnowledgeNodeShort, parent *models.ContainerDescription) (*models.KnowledgeNodeFull, error) {
	newKnowledgeNode, err := s.knowledgeNodesRepository.AddKnowledgeNode(ctx, knowledgeNode.Name, knowledgeNode.Tags, knowledgeNode.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(ctx, parent.Type, parent.ID, schema.ContainerTypeKnowledgeNode, newKnowledgeNode.ID)
		if err != nil {
			return nil, err
		}
	}
	return s.GetKnowledgeNodeFull(ctx, newKnowledgeNode.ID)
}

func (s *KnowledgeNodesService) UpdateKnowledgeNode(ctx context.Context, knowledgeNodePartial models.KnowledgeNodePartial) (*models.KnowledgeNodeFull, error) {
	err := s.knowledgeNodesRepository.UpdateKnowledgeNode(ctx, knowledgeNodePartial)
	if err != nil {
		return nil, err
	}
	return s.GetKnowledgeNodeFull(ctx, knowledgeNodePartial.ID)
}
