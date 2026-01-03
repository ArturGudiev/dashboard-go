package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"fmt"
	"os"

	"github.com/ddddddO/gtree"
)

type ContainerService struct {
	client *ent.Client
}

func NewContainerService(client *ent.Client) *ContainerService {
	return &ContainerService{
		client: client,
	}
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
	case schema.ContainerTypeProblem:
		problem, err := s.client.Problem.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", containerType, ID, problem.Description)
		return &result, nil
	default:
		return nil, fmt.Errorf("unsupported container type: %s", containerType)
	}
}

func (s *ContainerService) GetParentsPathDescriptions(ctx context.Context, containerType schema.ContainerType, ID int) []string {
	var items []string
	currentType := containerType
	currentID := ID
	if description, err := s.GetDescription(ctx, containerType, ID); err == nil {
		items = append(items, *description)
	}

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

func (s *ContainerService) PrintParentsPath(ctx context.Context, containerType schema.ContainerType, ID int) {
	parentsPath := s.GetParentsPathDescriptions(ctx, containerType, ID)
	if len(parentsPath) > 0 {
		for i, j := 0, len(parentsPath)-1; i < j; i, j = i+1, j-1 {
			parentsPath[i], parentsPath[j] = parentsPath[j], parentsPath[i]
		}

		// Build tree structure
		root := gtree.NewRoot(parentsPath[0])
		currentNode := root

		// Add intermediate parents
		for i := 1; i < len(parentsPath); i++ {
			currentNode = currentNode.Add(parentsPath[i])
		}

		res, err := s.GetDescription(ctx, containerType, ID)
		if err != nil && res != nil {
			currentNode.Add(*res)
		}

		// Print the tree
		if err := gtree.OutputFromRoot(os.Stdout, root); err != nil {
			fmt.Printf("Error printing tree: %v\n", err)
		}
		fmt.Println()
	}
}

func (s *ContainerService) PrintSubtasks(subtasks []*ent.Task) {
	if len(subtasks) > 0 {
		fmt.Println("\nChild Tasks:")
		for i, childTask := range subtasks {
			fmt.Printf("  %d. [ID: %d] Open - %s\n", i+1, childTask.ID, childTask.Description)
		}
	}
}

func (s *ContainerService) PrintProblems(problems []*ent.Problem) {
	if len(problems) > 0 {
		fmt.Println("\nChild Problems:")
		for i, childProblem := range problems {
			status := "Open"
			if childProblem.Solution != nil {
				status = "Solved"
			}
			fmt.Printf("  %d. [ID: %d] %s - %s\n", i+1, childProblem.ID, status, childProblem.Description)
		}
	}
}

func (s *ContainerService) AddSubproblem(ctx context.Context, parentType schema.ContainerType, parentID int, description string) error {
	// Create the new problem (not done - solution is null by default)
	newProblem, err := s.client.Problem.Create().
		SetDescription(description).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create problem: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeProblem),
		).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count children: %v", err)
	}

	// Get the count of existing parents to set parent_order
	parentCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(parentType),
			containerchild.ChildID(parentID),
			containerchild.ParentTypeEQ(schema.ContainerTypeTask),
		).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count parents: %v", err)
	}

	// Create the parent-child relationship
	_, err = s.client.ContainerChild.Create().
		SetParentType(parentType).
		SetParentID(parentID).
		SetChildType(schema.ContainerTypeProblem).
		SetChildID(newProblem.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}

	fmt.Printf("Problem created successfully! ID: %d\n", newProblem.ID)
	return nil
}

func (s *ContainerService) AddSubtask(ctx context.Context, parentType schema.ContainerType, parentID int, description string) (*ent.Task, error) {
	// Create the new task
	newTask, err := s.client.Task.Create().
		SetDescription(description).
		SetDone(false).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeTask),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count children: %v", err)
	}

	// Get the count of existing parents to set parent_order
	parentCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(parentType),
			containerchild.ChildID(newTask.ID),
			containerchild.ParentTypeEQ(schema.ContainerTypeTask),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count parents: %v", err)
	}

	// Create the parent-child relationship
	_, err = s.client.ContainerChild.Create().
		SetParentType(parentType).
		SetParentID(parentID).
		SetChildType(schema.ContainerTypeTask).
		SetChildID(newTask.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %v", err)
	}

	return newTask, nil
}
