package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/utils"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type CLIService struct {
	client           *ent.Client
	containerService *ContainerService
}

func NewCLIService(client *ent.Client, containerService *ContainerService) *CLIService {
	return &CLIService{
		client:           client,
		containerService: containerService,
	}
}

func (s *CLIService) printTaskInfo(ctx context.Context, t *ent.Task, subtasks []*ent.Task, problems []*ent.Problem) {
	s.containerService.PrintParentsPath(ctx, schema.ContainerTypeTask, t.ID)

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tTask ID: %d\n", t.ID)
	fmt.Printf("\tDescription: %s\n", t.Description)
	fmt.Printf("\tStatus: %s\n", map[bool]string{true: "Done", false: "Open"}[t.Done])
	if t.Notes != "" {
		fmt.Printf("Notes: %s\n", t.Notes)
	}
	if len(t.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(t.Tags, ", "))
	}
	if t.DoneDateTime != nil {
		fmt.Printf("Done Date: %s\n", t.DoneDateTime.Format("2006-01-02 15:04:05"))
	}
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	fmt.Println()
}

func (s *CLIService) ViewTaskInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id // Track current task ID
	subtasks, _ := s.containerService.GetSubtasks(ctx, schema.ContainerTypeTask, currentID)
	problems, _ := s.containerService.GetProblems(ctx, schema.ContainerTypeTask, currentID)

	for {
		// Clear screen before printing
		utils.ClearScreen()

		// Get task
		t, err := s.client.Task.Get(ctx, currentID)
		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Task %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting task: %v\n", err)
			}
			return
		}

		// Print task information
		s.printTaskInfo(ctx, t, subtasks, problems)

		// Wait for user input
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		// Check for t+ command (case-sensitive for the +)
		if strings.HasPrefix(line, "t+ ") {
			description := strings.TrimSpace(line[3:])
			if description == "" {
				fmt.Println("Error: Description required. Usage: t+ <description>")
				utils.WaitForUserInput()
				continue
			}
			if _, err := s.containerService.AddSubtask(ctx, schema.ContainerTypeTask, id, description); err != nil {
				fmt.Printf("Error adding subtask: %v\n", err)
				utils.WaitForUserInput()
				continue
			}
			// Continue loop to refresh and show the new subtask
			continue
		}

		// Check for p+ command (case-sensitive for the +)
		if strings.HasPrefix(line, "p+ ") {
			description := strings.TrimSpace(line[3:])
			if description == "" {
				fmt.Println("Error: Description required. Usage: p+ <description>")
				utils.WaitForUserInput()
				continue
			}
			if err := s.containerService.AddSubproblem(ctx, currentID, description); err != nil {
				fmt.Printf("Error adding problem: %v\n", err)
				utils.WaitForUserInput()
				continue
			}
			// Continue loop to refresh and show the new problem
			continue
		}

		if strings.HasPrefix(line, "p ") {
			indexPart := strings.TrimSpace(line[2:])
			if index, err := strconv.Atoi(indexPart); err == nil {
				selectedProblem := problems[index-1]
				s.ViewProblemInteractive(ctx, selectedProblem.ID)
				continue
			}
		}

		if index, err := strconv.Atoi(line); err == nil {
			// Validate index
			if index < 1 || index > len(subtasks) {
				utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(subtasks)))
				continue
			}

			childTaskID := subtasks[index-1].ID
			s.ViewTaskInteractive(ctx, childTaskID)
			continue
		}

		if line == "u" {
			err := s.NavigateToParent(ctx, schema.ContainerTypeTask, t.ID)
			if err != nil {
				fmt.Printf("Error navigating to parent: %v\n", err)
			}
			continue
		}

		if strings.HasPrefix(line, "ft ") {
			indexStr := strings.TrimSpace(line[3:])
			if indexStr == "" {
				fmt.Println("Error: index required. Usage: ft <id>")
				utils.WaitForUserInput()
				continue
			}

			var taskDescriptions []string
			mapper := func(el *ent.Task) string { return el.Description }
			for _, el := range subtasks {
				taskDescriptions = append(taskDescriptions, mapper(el))
			}

			fmt.Println(strings.Join(taskDescriptions, ", "))

			utils.WaitForUserInput()
			// Continue loop to refresh and show the new subtask
			continue
		}

		lineLower := strings.ToLower(line)
		switch lineLower {
		case "q", "quit", "exit":
			os.Exit(0)
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, or 'p+ <description>' to add problem.\n", line)
			utils.WaitForUserInput()
		}
	}
}

func (s *CLIService) PrintProblemInfo(ctx context.Context, p *ent.Problem, subtasks []*ent.Task, problems []*ent.Problem) {
	s.containerService.PrintParentsPath(ctx, schema.ContainerTypeProblem, p.ID)

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tProblem ID: %d\n", p.ID)
	fmt.Printf("\tDescription: %s\n", p.Description)
	fmt.Printf("\tStatus: %s\n", map[bool]string{true: "Solved", false: "Open"}[p.Solution !=
		nil])
	if p.Notes != "" {
		fmt.Printf("Notes: %s\n", p.Notes)
	}
	if len(p.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(p.Tags, ", "))
	}
	if p.DoneDateTime != nil {
		fmt.Printf("Done Date: %s\n", p.DoneDateTime.Format("2006-01-02 15:04:05"))
	}
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	fmt.Println()
}

func (s *CLIService) ViewProblemInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id
	problem, _ := s.client.Problem.Get(ctx, currentID)
	subtasks, _ := s.containerService.GetSubtasks(ctx, schema.ContainerTypeProblem, currentID)
	problems, _ := s.containerService.GetProblems(ctx, schema.ContainerTypeProblem, currentID)

	for {
		// Clear screen before printing
		utils.ClearScreen()
		s.PrintProblemInfo(ctx, problem, subtasks, problems)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		// Check for t+ command (case-sensitive for the +)
		if strings.HasPrefix(line, "t+ ") {
			description := strings.TrimSpace(line[3:])
			if description == "" {
				fmt.Println("Error: Description required. Usage: t+ <description>")
				utils.WaitForUserInput()
				continue
			}
			if _, err := s.containerService.AddSubtask(ctx, schema.ContainerTypeProblem, currentID, description); err != nil {
				fmt.Printf("Error adding subtask: %v\n", err)
				utils.WaitForUserInput()
				continue
			}
			// Continue loop to refresh and show the new subtask
			continue
		}

		// Check for p+ command (case-sensitive for the +)
		if strings.HasPrefix(line, "p+ ") {
			//description := strings.TrimSpace(line[3:])
			//if description == "" {
			//	fmt.Println("Error: Description required. Usage: p+ <description>")
			//	utils.WaitForUserInput()
			//	continue
			//}
			//if err := s.containerService.addSubproblem(ctx, application, currentID, description); err != nil {
			//	fmt.Printf("Error adding problem: %v\n", err)
			//	utils.WaitForUserInput()
			//	continue
			//}
			// Continue loop to refresh and show the new problem
			continue
		}

		if strings.HasPrefix(line, "p ") {
			if index, err := strconv.Atoi(line); err == nil {
				selectedProblem := problems[index]
				s.ViewProblemInteractive(ctx, selectedProblem.ID)
			}
		}

		if index, err := strconv.Atoi(line); err == nil {
			// Validate index
			if index < 1 || index > len(subtasks) {
				utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(subtasks)))
				continue
			}

			childTaskID := subtasks[index-1].ID
			s.ViewTaskInteractive(ctx, childTaskID)
			continue
		}

		if line == "u" {
			err := s.NavigateToParent(ctx, schema.ContainerTypeProblem, currentID)
			if err != nil {
			}
			continue
		}

		if strings.HasPrefix(line, "ft ") {
			indexStr := strings.TrimSpace(line[3:])
			if indexStr == "" {
				fmt.Println("Error: index required. Usage: ft <id>")
				utils.WaitForUserInput()
				continue
			}

			var taskDescriptions []string
			mapper := func(el *ent.Task) string { return el.Description }
			for _, el := range subtasks {
				taskDescriptions = append(taskDescriptions, mapper(el))
			}

			fmt.Println(strings.Join(taskDescriptions, ", "))

			utils.WaitForUserInput()
			// Continue loop to refresh and show the new subtask
			continue
		}

		lineLower := strings.ToLower(line)
		switch lineLower {
		case "q", "quit", "exit":
			os.Exit(0)
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:

			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, or 'p+ <description>' to add problem.\n", line)
			utils.WaitForUserInput()
		}
	}
}

func (s *CLIService) ViewContainerInteractive(ctx context.Context, containerType schema.ContainerType, ID int) {
	switch containerType {
	case schema.ContainerTypeTask:
		s.ViewTaskInteractive(ctx, ID)
	case schema.ContainerTypeProblem:
		s.ViewProblemInteractive(ctx, ID)
	}

}

func (s *CLIService) NavigateToParent(ctx context.Context, containerType schema.ContainerType, ID int) error {
	parentType, parentID, err := s.containerService.GetParentCommon(ctx, containerType, ID)
	if parentType == nil || err != nil {
		fmt.Println("This task has no parent.")
		utils.WaitForUserInput()
		return errors.New("Error, cant get parent")
	}
	s.ViewContainerInteractive(ctx, *parentType, parentID)
	return nil
}
