package services

import (
	"arturgudiev/dashboard/constants"
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/ent/task"
	"arturgudiev/dashboard/repositories"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ddddddO/gtree"
	"github.com/fatih/color"
)

type ContainerService struct {
	client                   *ent.Client
	childContainerRepository *ChildContainerRepository
	aliasesRepository        *repositories.AliasesRepository
}

func NewContainerService(client *ent.Client, childContainerRepository *ChildContainerRepository, aliasesRepository *repositories.AliasesRepository) *ContainerService {
	return &ContainerService{
		client:                   client,
		childContainerRepository: childContainerRepository,
		aliasesRepository:        aliasesRepository,
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

// func (s *ContainerService) GetAllSubtasksAsRelations(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.ContainerChild, error) {
// 	allTasksRelations := []*ent.ContainerChild{}
// 	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeTask)
// 	if err != nil {
// 		return []*ent.ContainerChild{}, err
// 	}

// 	return childRelations, nil
// }

func (s *ContainerService) ChangeTasksOrder(
	ctx context.Context,
	parentType schema.ContainerType,
	parentID int,
	tasksInNewOrder []int,
) ([]*ent.ContainerChild, error) {

	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, parentType, parentID, schema.ContainerTypeTask)
	if err != nil {
		return nil, err
	}

	allowedTaskIDs := make(map[int]struct{}, len(tasksInNewOrder))
	for _, taskID := range tasksInNewOrder {
		allowedTaskIDs[taskID] = struct{}{}
	}

	filteredRelations := make([]*ent.ContainerChild, 0, len(childRelations))
	for _, relation := range childRelations {
		if _, ok := allowedTaskIDs[relation.ChildID]; !ok {
			continue
		}
		filteredRelations = append(filteredRelations, relation)
	}

	orderIndex := make(map[int]int, len(tasksInNewOrder))
	for i, taskID := range tasksInNewOrder {
		orderIndex[taskID] = i
	}

	slices.SortFunc(filteredRelations, func(a, b *ent.ContainerChild) int {
		return orderIndex[a.ChildID] - orderIndex[b.ChildID]
	})

	for i, relation := range filteredRelations {
		relation.ChildOrder = i + 1
	}

	if err := s.childContainerRepository.UpdateChildOrders(ctx, filteredRelations); err != nil {
		return nil, err
	}

	return filteredRelations, nil
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

func (s *ContainerService) GetOpenQuestionsIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openQuestions := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeQuestion)

	if err == nil && len(childRelations) > 0 {
		// Filter to only open problems - load problems manually since edges don't exist
		for _, relation := range childRelations {
			childProblem, err := s.client.Question.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			// Only include problems that don't have a solution (open problems)
			if childProblem.Answer == nil {
				openQuestions = append(openQuestions, childProblem.ID)
			}
		}
	}
	return openQuestions, nil
}

func (s *ContainerService) GetOpenLongTasks(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.LongTask, error) {
	var openLongTasks []*ent.LongTask
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeLongTask)

	if err == nil && len(childRelations) > 0 {
		for _, relation := range childRelations {
			childLongTask, err := s.client.LongTask.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			if !childLongTask.Done {
				openLongTasks = append(openLongTasks, childLongTask)
			}
		}
	}
	return openLongTasks, nil
}

func (s *ContainerService) GetOpenLongTasksIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openLongTasks := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeLongTask)

	if err == nil && len(childRelations) > 0 {
		for _, relation := range childRelations {
			childLongTask, err := s.client.LongTask.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			if !childLongTask.Done {
				openLongTasks = append(openLongTasks, childLongTask.ID)
			}
		}
	}
	return openLongTasks, nil
}

// CollectDescendantTaskAndLongTaskIDs returns all task and long-task IDs under a container tree.
func (s *ContainerService) CollectDescendantTaskAndLongTaskIDs(
	ctx context.Context,
	parentType schema.ContainerType,
	parentID int,
) ([]int, []int, error) {
	taskIDs := []int{}
	longTaskIDs := []int{}
	if err := s.collectDescendantTaskAndLongTaskIDs(ctx, parentType, parentID, &taskIDs, &longTaskIDs); err != nil {
		return nil, nil, err
	}
	return taskIDs, longTaskIDs, nil
}

func (s *ContainerService) collectDescendantTaskAndLongTaskIDs(
	ctx context.Context,
	parentType schema.ContainerType,
	parentID int,
	taskIDs *[]int,
	longTaskIDs *[]int,
) error {
	longTaskRelations, err := s.childContainerRepository.GetChildContainers(ctx, parentType, parentID, schema.ContainerTypeLongTask)
	if err != nil {
		return err
	}
	for _, relation := range longTaskRelations {
		*longTaskIDs = append(*longTaskIDs, relation.ChildID)
	}

	taskRelations, err := s.childContainerRepository.GetChildContainers(ctx, parentType, parentID, schema.ContainerTypeTask)
	if err != nil {
		return err
	}
	for _, relation := range taskRelations {
		*taskIDs = append(*taskIDs, relation.ChildID)
		if err := s.collectDescendantTaskAndLongTaskIDs(ctx, schema.ContainerTypeTask, relation.ChildID, taskIDs, longTaskIDs); err != nil {
			return err
		}
	}

	storyRelations, err := s.childContainerRepository.GetChildContainers(ctx, parentType, parentID, schema.ContainerTypeStory)
	if err != nil {
		return err
	}
	for _, relation := range storyRelations {
		if err := s.collectDescendantTaskAndLongTaskIDs(ctx, schema.ContainerTypeStory, relation.ChildID, taskIDs, longTaskIDs); err != nil {
			return err
		}
	}

	directionRelations, err := s.childContainerRepository.GetChildContainers(ctx, parentType, parentID, schema.ContainerTypeDirection)
	if err != nil {
		return err
	}
	for _, relation := range directionRelations {
		if err := s.collectDescendantTaskAndLongTaskIDs(ctx, schema.ContainerTypeDirection, relation.ChildID, taskIDs, longTaskIDs); err != nil {
			return err
		}
	}
	return nil
}

// GetDoneTasksByIDs returns done tasks with a done_date_time among the given IDs.
func (s *ContainerService) GetDoneTasksByIDs(ctx context.Context, taskIDs []int) ([]*ent.Task, error) {
	if len(taskIDs) == 0 {
		return []*ent.Task{}, nil
	}
	return s.client.Task.Query().
		Where(
			task.DoneEQ(true),
			task.DoneDateTimeNotNil(),
			task.IDIn(taskIDs...),
		).
		All(ctx)
}

func (s *ContainerService) GetOpenDirectionsIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openDirections := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeDirection)

	if err == nil && len(childRelations) > 0 {
		for _, relation := range childRelations {
			childDirection, err := s.client.Direction.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}
			if !childDirection.Closed {
				openDirections = append(openDirections, childDirection.ID)
			}
		}
	}
	return openDirections, nil
}

func (s *ContainerService) GetOpenStoriesIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openStories := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeStory)

	if err == nil && len(childRelations) > 0 {
		for _, relation := range childRelations {
			childStory, err := s.client.Story.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}

			if !childStory.Closed {
				openStories = append(openStories, childStory.ID)
			}
		}
	}
	return openStories, nil
}

func (s *ContainerService) GetOpenEpicsIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	openEpics := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeEpic)

	if err == nil && len(childRelations) > 0 {
		for _, relation := range childRelations {
			childEpic, err := s.client.Story.Get(ctx, relation.ChildID)
			if err != nil {
				continue
			}

			if !childEpic.Closed {
				openEpics = append(openEpics, childEpic.ID)
			}
		}
	}
	return openEpics, nil
}

func (s *ContainerService) GetChildKnowledgeNodesIDs(ctx context.Context, containerType schema.ContainerType, ID int) ([]int, error) {
	knowledgeNodes := []int{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeKnowledgeNode)

	if err == nil && len(childRelations) > 0 {
		for _, relation := range childRelations {
			knowledgeNodes = append(knowledgeNodes, relation.ChildID)
		}
	}
	return knowledgeNodes, nil
}

func (s *ContainerService) GetOpenKnowledgeNodes(ctx context.Context, containerType schema.ContainerType, ID int) ([]*ent.KnowledgeNode, error) {
	knowledgeNodes := []*ent.KnowledgeNode{}
	childRelations, err := s.childContainerRepository.GetChildContainers(ctx, containerType, ID, schema.ContainerTypeKnowledgeNode)
	if err != nil {
		return nil, err
	}

	for _, relation := range childRelations {
		childKnowledgeNode, err := s.client.KnowledgeNode.Get(ctx, relation.ChildID)
		if err != nil {
			continue
		}
		knowledgeNodes = append(knowledgeNodes, childKnowledgeNode)
	}
	return knowledgeNodes, nil
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
	capitalContainerType := constants.CapitalisedContainerTypes[containerType]
	switch containerType {
	case schema.ContainerTypeTask:
		task, err := s.client.Task.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, task.Description)
		return &result, nil
	case schema.ContainerTypeProblem:
		problem, err := s.client.Problem.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, problem.Description)
		return &result, nil
	case schema.ContainerTypeQuestion:
		question, err := s.client.Question.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, question.Description)
		return &result, nil
	case schema.ContainerTypeStory:
		story, err := s.client.Story.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, story.Description)
		return &result, nil
	case schema.ContainerTypeEpic:
		epic, err := s.client.Epic.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, epic.Description)
		return &result, nil
	case schema.ContainerTypeKnowledgeNode:
		knowledgeNode, err := s.client.KnowledgeNode.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, knowledgeNode.Name)
		return &result, nil
	case schema.ContainerTypeLongTask:
		longTask, err := s.client.LongTask.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, longTask.Description)
		return &result, nil
	case schema.ContainerTypeDirection:
		direction, err := s.client.Direction.Get(ctx, ID)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("%s-%d %s", capitalContainerType, ID, direction.Description)
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
			color.RGB(15, 82, 186).Printf("  %d. [ID: %d] %s - %s\n", i+1, childEpic.ID, status, childEpic.Description)
		}
	}
}

func (s *ContainerService) PrintKnowledgeNodes(knowledgeNodes []*ent.KnowledgeNode) {
	if len(knowledgeNodes) > 0 {
		fmt.Println("\nChild Knowledge Nodes:")
		for i, childKnowledgeNode := range knowledgeNodes {
			fmt.Printf("  %d. [ID: %d] %s\n", i+1, childKnowledgeNode.ID, childKnowledgeNode.Name)
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

func (s *ContainerService) AddSubknowledgeNode(ctx context.Context, parentType schema.ContainerType, parentID int, name string) error {
	// Create the new knowledge node
	newKnowledgeNode, err := s.client.KnowledgeNode.Create().
		SetName(name).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create knowledge node: %v", err)
	}

	// Get the count of existing children to set child_order
	childCount, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(schema.ContainerTypeKnowledgeNode),
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
		SetChildType(schema.ContainerTypeKnowledgeNode).
		SetChildID(newKnowledgeNode.ID).
		SetChildOrder(childCount).
		SetParentOrder(parentCount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}

	fmt.Printf("Knowledge Node created successfully! ID: %d\n", newKnowledgeNode.ID)
	return nil
}

func (s *ContainerService) GetFilesFolder(ctx context.Context, containerType schema.ContainerType, ID int) *string {
	filesDir := s.GetFilesFolderPrefix(ctx, containerType, ID)
	if filesDir == nil {
		return nil
	}

	// Extract directory path from prefix (everything before the filename part)
	prefix := *filesDir
	dir := filepath.Dir(prefix)
	prefixBase := filepath.Base(prefix)

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Check if any directory starts with the prefix and return the full path
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefixBase) {
			fullPath := filepath.Join(dir, entry.Name())
			return &fullPath
		}
	}

	return nil
}

func (s *ContainerService) GetFilesFolderPrefix(ctx context.Context, containerType schema.ContainerType, ID int) *string {
	folderPath := "C:\\Programming\\NodeJS\\Dashboard\\files"
	var result string
	switch containerType {
	case schema.ContainerTypeTask:
		result = fmt.Sprintf("%s\\tasks\\%d_", folderPath, ID)
	case schema.ContainerTypeProblem:
		result = fmt.Sprintf("%s\\problems\\%d_", folderPath, ID)
	case schema.ContainerTypeQuestion:
		result = fmt.Sprintf("%s\\questions\\%d_", folderPath, ID)
	case schema.ContainerTypeAction:
		result = fmt.Sprintf("%s\\actions\\%d_", folderPath, ID)
	case schema.ContainerTypeDefinition:
		result = fmt.Sprintf("%s\\definitions\\%d_", folderPath, ID)
	case schema.ContainerTypeKnowledgeBit:
		result = fmt.Sprintf("%s\\knowledge-bits\\%d_", folderPath, ID)
	case schema.ContainerTypeKnowledgeNode:
		result = fmt.Sprintf("%s\\knowledge-nodes\\%d_", folderPath, ID)
	case schema.ContainerTypeStory:
		result = fmt.Sprintf("%s\\stories\\%d_", folderPath, ID)
	case schema.ContainerTypeEpic:
		result = fmt.Sprintf("%s\\epics\\%d_", folderPath, ID)
	default:
		return nil
	}
	return &result
}
