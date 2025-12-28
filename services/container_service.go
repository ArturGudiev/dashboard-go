package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"fmt"
)

type ContainerService struct {
	client *ent.Client
}

func NewContainerService(client *ent.Client) *ContainerService {
	return &ContainerService{client: client}
}

func (s *ContainerService) GetSubtasks(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Task, error) {
	openTasks := []*ent.Task{}
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerType),
			containerchild.ParentID(ID),
			containerchild.ChildTypeEQ(schema.ContainerTypeTask),
		).
		Order(containerchild.ByChildOrder()).
		All(ctx)

	if err == nil && len(childRelations) > 0 {
		// Filter to only open tasks - load tasks manually since edges don't exist
		for _, relation := range childRelations {
			childTask, err := s.client.Task.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			if !childTask.Done {
				openTasks = append(openTasks, childTask)
			}
		}
	}
	return openTasks, nil
}

func (s *ContainerService) GetProblems(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Problem, error) {
	openProblems := []*ent.Problem{}
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerType),
			containerchild.ParentID(ID),
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
		).
		Order(containerchild.ByChildOrder()).
		All(ctx)

	if err == nil && len(childRelations) > 0 {
		// Filter to only open problems - load problems manually since edges don't exist
		for _, relation := range childRelations {
			childProblem, err := s.client.Problem.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			// Only include problems that don't have a solution (open problems)
			if childProblem.Solution == nil {
				openProblems = append(openProblems, childProblem)
			}
		}
	}
	return openProblems, nil
}

func (s *ContainerService) GetParentCommon(ctx context.Context, containerType schema.ContainerType, ID int) (*schema.ContainerType, int, error) {
	parentRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildID(ID),
			containerchild.ChildTypeEQ(containerType),
		).
		All(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to query parent relations: %w", err)
	}
	if len(parentRelations) == 0 {
		return nil, 0, fmt.Errorf("no parent found for container type %s with ID %d", containerType, ID)
	}

	parentID := parentRelations[0].ParentID
	parentContainerType := parentRelations[0].ParentType

	return &parentContainerType, parentID, nil
}

func (s *ContainerService) GetDescription(ctx context.Context, containerType schema.ContainerType, ID int) (*string, error) {
	switch containerType {
	case schema.ContainerTypeTask:
		task, err := s.client.Task.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", containerType, ID, task.Description)
		return &result, nil
	default:
		return nil, fmt.Errorf("unsupported container type: %s", containerType)
	}
}

func (s *ContainerService) GetParentsPathDescriptions(ctx context.Context, containerType schema.ContainerType, ID int) []string {
	var items []string
	currentType := containerType
	currentID := ID

	for {
		parentType, pID, err := s.GetParentCommon(ctx, currentType, currentID)
		if err != nil {
			// No more parents found, we're done
			break
		}

		description, err := s.GetDescription(ctx, *parentType, pID)
		if err != nil {
			// If we can't get description, stop traversing
			break
		}
		items = append(items, *description)

		// Move up to the parent for next iteration
		currentType = *parentType
		currentID = pID
	}

	return items
}
