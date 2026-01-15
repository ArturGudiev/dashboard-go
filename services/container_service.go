package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"fmt"
	"os"

	"github.com/ddddddO/gtree"
	"github.com/fatih/color"
)

type ContainerService struct {
	client                   *ent.Client
	childContainerRepository *ChildContainerRepository
}

func NewContainerService(client *ent.Client, childContainerRepository *ChildContainerRepository) *ContainerService {
	return &ContainerService{
		client:                   client,
		childContainerRepository: childContainerRepository,
	}
}

func (s *ContainerService) GetOpenSubtasksIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openTasksIDs := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeTask)
	if err != nil {
		return []int{}, err
	}

	for _, relation := range childRelations {
		childTask, err := s.client.Task.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}
		if !childTask.Done {
			openTasksIDs = append(openTasksIDs, childTask.ID)
		}
	}
	return openTasksIDs, nil
}

func (s *ContainerService) GetOpenSubtasks(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Task, error) {
	openTasks := []*ent.Task{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeTask)
	if err != nil {
		return nil, err
	}

	for _, relation := range childRelations {
		childTask, err := s.client.Task.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}
		if !childTask.Done {
			openTasks = append(openTasks, childTask)
		}
	}
	return openTasks, nil
}

func (s *ContainerService) GetOpenStories(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Story, error) {
	openStories := []*ent.Story{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeStory)
	if err != nil {
		return nil, err
	}

	for _, relation := range childRelations {
		childStory, err := s.client.Story.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}
		if !childStory.Closed {
			openStories = append(openStories, childStory)
		}
	}
	return openStories, nil
}

func (s *ContainerService) GetOpenEpics(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Epic, error) {
	openEpics := []*ent.Epic{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeEpic)
	if err != nil {
		return nil, err
	}

	for _, relation := range childRelations {
		childEpic, err := s.client.Epic.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}
		if !childEpic.Closed {
			openEpics = append(openEpics, childEpic)
		}
	}
	return openEpics, nil
}

func (s *ContainerService) GetOpenProblems(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Problem, error) {
	var openProblems []*ent.Problem
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeProblem)

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

func (s *ContainerService) GetOpenQuestions(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.Question, error) {
	var openQuestions []*ent.Question
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeQuestion)

	if err == nil && len(childRelations) > 0 {
		// Filter to only open Questions - load Questions manually since edges don't exist
		for _, relation := range childRelations {
			childQuestion, err := s.client.Question.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			// Only include Questions that don't have a solution (open Questions)
			if childQuestion.Answer == nil {
				openQuestions = append(openQuestions, childQuestion)
			}
		}
	}
	return openQuestions, nil
}

func (s *ContainerService) GetOpenProblemsIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openProblems := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeProblem)

	if err == nil && len(childRelations) > 0 {
		// Filter to only open problems - load problems manually since edges don't exist
		for _, relation := range childRelations {
			childProblem, err := s.client.Problem.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			// Only include problems that don't have a solution (open problems)
			if childProblem.Solution == nil {
				openProblems = append(openProblems, childProblem.ID)
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
	case schema.ContainerTypeQuestion:
		question, err := s.client.Question.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", containerType, ID, question.Description)
		return &result, nil
	case schema.ContainerTypeStory:
		story, err := s.client.Story.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", containerType, ID, story.Description)
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

func (s *ContainerService) PrintQuestions(questions []*ent.Question) {
	if len(questions) > 0 {
		fmt.Println("\nChild Questions:")
		for i, childProblem := range questions {
			status := "Open"
			if childProblem.Answer != nil {
				status = "Solved"
			}
			fmt.Printf("  %d. [ID: %d] %s - %s\n", i+1, childProblem.ID, status, childProblem.Description)
		}
	}
}

func (s *ContainerService) PrintStories(stories []*ent.Story) {
	if len(stories) > 0 {
		fmt.Println("\nChild Stories:")
		for i, childStory := range stories {
			status := "Open"
			if childStory.Closed {
				status = "Closed"
			}
			fmt.Printf("  %d. [ID: %d] %s - %s\n", i+1, childStory.ID, status, childStory.Description)
		}
	}
}

func (s *ContainerService) PrintEpics(epics []*ent.Epic) {

	if len(epics) > 0 {
		color.RGB(15, 82, 186).Println("foreground orange")
		fmt.Println("\nChild Epics:")
		for i, childEpic := range epics {
			status := "Open"
			if childEpic.Closed {
				status = "Closed"
			}
			color.RGB(15, 82, 186).Printf(fmt.Sprintf("  %d. [ID: %d] %s - %s\n", i+1, childEpic.ID, status, childEpic.Description))
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

func (s *ContainerService) AddSubquestion(ctx context.Context, parentType schema.ContainerType, parentID int, description string) error {
	// Create the new question (not done - answer is null by default)
	newQuestion, err := s.client.Question.Create().
		SetDescription(description).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create question: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeQuestion),
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
		SetChildType(schema.ContainerTypeQuestion).
		SetChildID(newQuestion.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}

	fmt.Printf("Question created successfully! ID: %d\n", newQuestion.ID)
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

func (s *ContainerService) AddSubstory(ctx context.Context, parentType schema.ContainerType, parentID int, description string) error {
	// Create the new story (not closed by default)
	newStory, err := s.client.Story.Create().
		SetDescription(description).
		SetTags([]string{}).
		SetNotes("").
		SetClosed(false).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create story: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeStory),
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
		SetChildType(schema.ContainerTypeStory).
		SetChildID(newStory.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}

	fmt.Printf("Story created successfully! ID: %d\n", newStory.ID)
	return nil
}

func (s *ContainerService) AddSubepic(ctx context.Context, parentType schema.ContainerType, parentID int, description string) error {
	// Create the new epic (not closed by default)
	newEpic, err := s.client.Epic.Create().
		SetDescription(description).
		SetTags([]string{}).
		SetNotes("").
		SetClosed(false).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create epic: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeEpic),
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
		SetChildType(schema.ContainerTypeEpic).
		SetChildID(newEpic.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}

	fmt.Printf("Epic created successfully! ID: %d\n", newEpic.ID)
	return nil
}
