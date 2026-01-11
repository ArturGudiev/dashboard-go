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
	client             *ent.Client
	containerService   *ContainerService
	problemsRepository *ProblemsRepository
}

func NewCLIService(client *ent.Client, containerService *ContainerService, problemsRepository *ProblemsRepository) *CLIService {
	return &CLIService{
		client:             client,
		containerService:   containerService,
		problemsRepository: problemsRepository,
	}
}

func (s *CLIService) printTaskInfo(ctx context.Context, t *ent.Task, subtasks []*ent.Task, problems []*ent.Problem, questions []*ent.Question) {
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
	s.containerService.PrintQuestions(questions)
	fmt.Println()
}

func (s *CLIService) ViewTaskInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id // Track current task ID

	for {
		// Clear screen before printing
		utils.ClearScreen()

		// Get task
		t, err := s.client.Task.Get(ctx, currentID)
		subtasks, _ := s.containerService.GetOpenSubtasks(ctx, schema.ContainerTypeTask, currentID)
		problems, _ := s.containerService.GetOpenProblems(ctx, schema.ContainerTypeTask, currentID)
		questions, _ := s.containerService.GetOpenQuestions(ctx, schema.ContainerTypeTask, currentID)

		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Task %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting task: %v\n", err)
			}
			return
		}

		// Print task information
		s.printTaskInfo(ctx, t, subtasks, problems, questions)

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

		if goToNextIteration := s.checkAddTaskCommand(ctx, line, schema.ContainerTypeTask, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddProblemCommand(ctx, line, schema.ContainerTypeTask, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddQuestionCommand(ctx, line, schema.ContainerTypeTask, id); goToNextIteration {
			continue
		}

		if wasIt := s.checkSelectTaskCommand(ctx, line, subtasks); wasIt {
			continue
		}

		if wasIt := s.checkSelectProblemCommand(ctx, line, problems); wasIt {
			continue
		}

		if wasIt := s.checkSelectQuestionCommand(ctx, line, questions); wasIt {
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

func (s *CLIService) checkAddTaskCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "t+ ") {
		description := strings.TrimSpace(line[3:])
		if description == "" {
			fmt.Println("Error: Description required. Usage: t+ <description>")
			utils.WaitForUserInput()
			return true
		}
		if _, err := s.containerService.AddSubtask(ctx, containerType, id, description); err != nil {
			fmt.Printf("Error adding subtask: %v\n", err)
			utils.WaitForUserInput()
			return true
		}
		// Continue loop to refresh and show the new subtask
		return true
	}
	return false
}

func (s *CLIService) checkAddProblemCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "p+ ") {
		description := strings.TrimSpace(line[3:])
		if description == "" {
			fmt.Println("Error: Description required. Usage: p+ <description>")
			utils.WaitForUserInput()
			return true
		}
		if err := s.containerService.AddSubproblem(ctx, containerType, id, description); err != nil {
			fmt.Printf("Error adding problem: %v\n", err)
			utils.WaitForUserInput()
			return true
		}
		// Continue loop to refresh and show the new problem
		return true
	}
	return false
}

func (s *CLIService) checkAddQuestionCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "q+ ") {
		description := strings.TrimSpace(line[3:])
		if description == "" {
			fmt.Println("Error: Description required. Usage: q+ <description>")
			utils.WaitForUserInput()
			return true
		}
		if err := s.containerService.AddSubquestion(ctx, containerType, id, description); err != nil {
			fmt.Printf("Error adding Question: %v\n", err)
			utils.WaitForUserInput()
			return true
		}
		// Continue loop to refresh and show the new Question
		return true
	}
	return false
}

func (s *CLIService) checkSelectTaskCommand(ctx context.Context, line string, tasks []*ent.Task) bool {
	if index, err := strconv.Atoi(line); err == nil {
		// Validate index
		if index < 1 || index > len(tasks) {
			utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(tasks)))
			return false
		}

		childTaskID := tasks[index-1].ID
		s.ViewTaskInteractive(ctx, childTaskID)
		return true
	}
	return false
}

func (s *CLIService) checkSelectProblemCommand(ctx context.Context, line string, problems []*ent.Problem) bool {
	if strings.HasPrefix(line, "p ") {
		indexPart := strings.TrimSpace(line[2:])
		if index, err := strconv.Atoi(indexPart); err == nil {
			selectedProblem := problems[index-1]
			s.ViewProblemInteractive(ctx, selectedProblem.ID)
			return true
		}
	}
	return false
}

func (s *CLIService) checkSelectQuestionCommand(ctx context.Context, line string, questions []*ent.Question) bool {
	if strings.HasPrefix(line, "q ") {
		indexPart := strings.TrimSpace(line[2:])
		if index, err := strconv.Atoi(indexPart); err == nil {
			// Validate index
			if index < 1 || index > len(questions) {
				utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(questions)))
				return false
			}
			selectedQuestion := questions[index-1]
			s.ViewQuestionInteractive(ctx, selectedQuestion.ID)
			return true
		}
	}
	return false
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

func (s *CLIService) PrintQuestionInfo(ctx context.Context, q *ent.Question, subtasks []*ent.Task, problems []*ent.Problem, questions []*ent.Question) {
	s.containerService.PrintParentsPath(ctx, schema.ContainerTypeQuestion, q.ID)

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tQuestion ID: %d\n", q.ID)
	fmt.Printf("\tDescription: %s\n", q.Description)
	fmt.Printf("\tStatus: %s\n", map[bool]string{true: "Answered", false: "Open"}[q.Answer != nil])
	if q.Notes != "" {
		fmt.Printf("Notes: %s\n", q.Notes)
	}
	if len(q.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(q.Tags, ", "))
	}
	if q.DoneDateTime != nil {
		fmt.Printf("Done Date: %s\n", q.DoneDateTime.Format("2006-01-02 15:04:05"))
	}
	if q.Answer != nil {
		fmt.Printf("Answer: %s\n", *q.Answer)
	}
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	s.containerService.PrintQuestions(questions)
	fmt.Println()
}

func (s *CLIService) ViewProblemInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id

	for {
		// Clear screen before printing
		utils.ClearScreen()
		problem, _ := s.client.Problem.Get(ctx, currentID)
		subtasks, _ := s.containerService.GetOpenSubtasks(ctx, schema.ContainerTypeProblem, currentID)
		problems, _ := s.containerService.GetOpenProblems(ctx, schema.ContainerTypeProblem, currentID)

		s.PrintProblemInfo(ctx, problem, subtasks, problems)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		if goToNextIteration := s.checkAddTaskCommand(ctx, line, schema.ContainerTypeProblem, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddProblemCommand(ctx, line, schema.ContainerTypeProblem, id); goToNextIteration {
			continue
		}

		if wasIt := s.checkSelectTaskCommand(ctx, line, subtasks); wasIt {
			continue
		}

		if wasIt := s.checkSelectProblemCommand(ctx, line, problems); wasIt {
			continue
		}

		if line == "res" {
			fmt.Print("Enter solution> ")
			solution := utils.GetUserInput(scanner)
			err := s.problemsRepository.AddSolution(ctx, currentID, solution)
			if err != nil {
				fmt.Printf("Error adding solution: %v\n", err)
			}
			navigationError := s.NavigateToParent(ctx, schema.ContainerTypeProblem, currentID)
			if navigationError != nil {
				fmt.Printf("Error navigating to parent: %v\n", navigationError)
			}
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

func (s *CLIService) ViewQuestionInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id

	for {
		// Clear screen before printing
		utils.ClearScreen()
		question, err := s.client.Question.Get(ctx, currentID)
		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Question %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting question: %v\n", err)
			}
			return
		}
		subtasks, _ := s.containerService.GetOpenSubtasks(ctx, schema.ContainerTypeQuestion, currentID)
		problems, _ := s.containerService.GetOpenProblems(ctx, schema.ContainerTypeQuestion, currentID)
		questions, _ := s.containerService.GetOpenQuestions(ctx, schema.ContainerTypeQuestion, currentID)

		s.PrintQuestionInfo(ctx, question, subtasks, problems, questions)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		if goToNextIteration := s.checkAddTaskCommand(ctx, line, schema.ContainerTypeQuestion, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddProblemCommand(ctx, line, schema.ContainerTypeQuestion, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddQuestionCommand(ctx, line, schema.ContainerTypeQuestion, id); goToNextIteration {
			continue
		}

		if wasIt := s.checkSelectTaskCommand(ctx, line, subtasks); wasIt {
			continue
		}

		if wasIt := s.checkSelectProblemCommand(ctx, line, problems); wasIt {
			continue
		}

		if wasIt := s.checkSelectQuestionCommand(ctx, line, questions); wasIt {
			continue
		}

		if line == "ans" {
			fmt.Print("Enter answer> ")
			answer := utils.GetUserInput(scanner)
			_, err := s.client.Question.UpdateOneID(currentID).SetAnswer(answer).Save(ctx)
			if err != nil {
				fmt.Printf("Error adding answer: %v\n", err)
			}
			navigationError := s.NavigateToParent(ctx, schema.ContainerTypeQuestion, currentID)
			if navigationError != nil {
				fmt.Printf("Error navigating to parent: %v\n", navigationError)
			}
		}

		if line == "u" {
			err := s.NavigateToParent(ctx, schema.ContainerTypeQuestion, currentID)
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
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, 'p+ <description>' to add problem, or 'q+ <description>' to add question.\n", line)
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
	case schema.ContainerTypeQuestion:
		s.ViewQuestionInteractive(ctx, ID)
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
