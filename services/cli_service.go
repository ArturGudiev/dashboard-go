package services

import (
	"arturgudiev/dashboard/assets"
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/repositories"
	"arturgudiev/dashboard/utils"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"arturgudiev/dashboard/models"

	"github.com/fatih/color"
)

type CLIService struct {
	client             *ent.Client
	containerService   *ContainerService
	aliasesService     *AliasesService
	problemsRepository *ProblemsRepository
	epicsRepository    *repositories.EpicsRepository
	aliasesRepository  *repositories.AliasesRepository
}

func NewCLIService(client *ent.Client, containerService *ContainerService,
	aliasesService *AliasesService,
	problemsRepository *ProblemsRepository,
	epicsRepository *repositories.EpicsRepository, aliasesRepository *repositories.AliasesRepository) *CLIService {
	return &CLIService{
		client:             client,
		containerService:   containerService,
		aliasesService:     aliasesService,
		problemsRepository: problemsRepository,
		epicsRepository:    epicsRepository,
		aliasesRepository:  aliasesRepository,
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
	filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeTask, t.ID)
	if filesDir != nil {
		fmt.Printf("\tFiles Directory: %s\n", *filesDir)
	}
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	s.containerService.PrintQuestions(questions)
	if filesDir != nil {
		s.printDirectoryContents(*filesDir)
	}
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

		if wasIt := s.checkOpenDirCommand(ctx, line, schema.ContainerTypeTask, id); wasIt {
			continue
		}

		if wasIt := s.checkAddMindMapTemplateCommand(ctx, line, schema.ContainerTypeTask, id); wasIt {
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
		case "q", "quit", "exit", "x":
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

func (s *CLIService) checkSelectFileCommand(ctx context.Context, line string, filesDir *string, files []os.DirEntry) bool {
	if strings.HasPrefix(line, "fs ") {
		if filesDir == nil {
			return false
		}
		indexPart := strings.TrimSpace(line[3:])
		if index, err := strconv.Atoi(indexPart); err == nil {
			selectedFile := files[index-1]
			fullPath := filepath.Join(*filesDir, selectedFile.Name())
			if selectedFile.IsDir() {
				utils.OpenDirectory(fullPath)
			} else {
				s.ViewFileInteractive(ctx, fullPath)
				// utils.OpenFile(fullPath)
				// utils.OpenFile(fullPath)
			}
			return true
		}
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

func (s *CLIService) checkAddStoryCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "s+ ") {
		description := strings.TrimSpace(line[3:])
		if description == "" {
			fmt.Println("Error: Description required. Usage: s+ <description>")
			utils.WaitForUserInput()
			return true
		}
		if err := s.containerService.AddSubstory(ctx, containerType, id, description); err != nil {
			fmt.Printf("Error adding story: %v\n", err)
			utils.WaitForUserInput()
			return true
		}
		// Continue loop to refresh and show the new story
		return true
	}
	return false
}

func (s *CLIService) checkAddEpicCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "e+ ") {
		description := strings.TrimSpace(line[3:])
		if description == "" {
			fmt.Println("Error: Description required. Usage: e+ <description>")
			utils.WaitForUserInput()
			return true
		}
		if err := s.containerService.AddSubepic(ctx, containerType, id, description); err != nil {
			fmt.Printf("Error adding epic: %v\n", err)
			utils.WaitForUserInput()
			return true
		}
		// Continue loop to refresh and show the new epic
		return true
	}
	return false
}

func (s *CLIService) checkSelectStoryCommand(ctx context.Context, line string, stories []*ent.Story) bool {
	if strings.HasPrefix(line, "s ") {
		indexPart := strings.TrimSpace(line[2:])
		if index, err := strconv.Atoi(indexPart); err == nil {
			// Validate index
			if index < 1 || index > len(stories) {
				utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(stories)))
				return false
			}
			selectedStory := stories[index-1]
			s.ViewStoryInteractive(ctx, selectedStory.ID)
			return true
		}
	}
	return false
}

func (s *CLIService) checkSelectEpicCommand(ctx context.Context, line string, epics []*ent.Epic) bool {
	if strings.HasPrefix(line, "e ") {
		indexPart := strings.TrimSpace(line[2:])
		if index, err := strconv.Atoi(indexPart); err == nil {
			// Validate index
			if index < 1 || index > len(epics) {
				utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(epics)))
				return false
			}
			selectedEpic := epics[index-1]
			s.ViewEpicInteractive(ctx, selectedEpic.ID)
			return true
		}
	}
	return false
}

func (s *CLIService) checkSelectKnowledgeNodeCommand(ctx context.Context, line string, knowledgeNodes []*ent.KnowledgeNode) bool {
	var indexPart string
	switch {
	case strings.HasPrefix(line, "kn "):
		indexPart = strings.TrimSpace(line[3:])
	case strings.HasPrefix(line, "n "):
		indexPart = strings.TrimSpace(line[2:])
	default:
		// Bare number selects nth child knowledge node (only used in knowledge-node view).
		if _, err := strconv.Atoi(line); err != nil {
			return false
		}
		indexPart = line
	}

	index, err := strconv.Atoi(indexPart)
	if err != nil {
		return false
	}
	if index < 1 || index > len(knowledgeNodes) {
		utils.PrintAndWait(fmt.Sprintf("Invalid index. Please enter a number between 1 and %d.\n", len(knowledgeNodes)))
		return true
	}
	selectedKnowledgeNode := knowledgeNodes[index-1]
	s.ViewKnowledgeNodeInteractive(ctx, selectedKnowledgeNode.ID)
	return true
}

func (s *CLIService) checkAddKnowledgeNodeCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "kn+ ") {
		name := strings.TrimSpace(line[4:])
		if name == "" {
			fmt.Println("Error: Name required. Usage: kn+ <name>")
			utils.WaitForUserInput()
			return true
		}
		if err := s.containerService.AddSubknowledgeNode(ctx, containerType, id, name); err != nil {
			fmt.Printf("Error adding knowledge node: %v\n", err)
			utils.WaitForUserInput()
			return true
		}
		// Continue loop to refresh and show the new knowledge node
		return true
	}
	return false
}

func (s *CLIService) checkOpenDirCommand(ctx context.Context, line string, containerType schema.ContainerType, ID int) bool {
	if line == "o" || line == "open" || line == "dir" {
		filesDir, err := s.containerService.GetOrCreateFilesFolder(ctx, containerType, ID)
		if err != nil {
			fmt.Printf("Error resolving files directory: %v\n", err)
			utils.WaitForUserInput()
			return true
		}

		err = utils.OpenDirectory(*filesDir)
		if err != nil {
			fmt.Printf("Error opening directory: %v\n", err)
			utils.WaitForUserInput()
			return true
		}

		return true
	}
	return false
}

// checkAddMindMapTemplateCommand handles "mm+" and "mm+ <name>".
// Copies assets/template.mm into the container files folder as template.mm,
// or as <name>.mm when a name argument is provided.
func (s *CLIService) checkAddMindMapTemplateCommand(ctx context.Context, line string, containerType schema.ContainerType, ID int) bool {
	fileName := assets.TemplateMindMapFileName
	switch {
	case line == "mm+":
		// default name
	case strings.HasPrefix(line, "mm+ "):
		name := strings.TrimSpace(line[4:])
		if name == "" {
			fmt.Println("Error: name required. Usage: mm+ [name]")
			utils.WaitForUserInput()
			return true
		}
		name = sanitizeMindMapFileName(name)
		if name == "" {
			fmt.Println("Error: invalid name after sanitizing")
			utils.WaitForUserInput()
			return true
		}
		if !strings.HasSuffix(strings.ToLower(name), ".mm") {
			name += ".mm"
		}
		fileName = name
	default:
		return false
	}

	filesDir, err := s.containerService.GetOrCreateFilesFolder(ctx, containerType, ID)
	if err != nil {
		fmt.Printf("Error resolving files directory: %v\n", err)
		utils.WaitForUserInput()
		return true
	}

	dest := filepath.Join(*filesDir, fileName)
	if _, err := os.Stat(dest); err == nil {
		fmt.Printf("File already exists: %s\n", dest)
		utils.WaitForUserInput()
		return true
	} else if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error checking destination: %v\n", err)
		utils.WaitForUserInput()
		return true
	}

	if err := os.WriteFile(dest, assets.TemplateMindMap, 0o644); err != nil {
		fmt.Printf("Error copying mind map template: %v\n", err)
		utils.WaitForUserInput()
		return true
	}

	fmt.Printf("Created: %s\n", dest)
	return true
}

func sanitizeMindMapFileName(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-", "?", "-",
		"\"", "", "<", "", ">", "", "|", "-",
	)
	return strings.TrimSpace(replacer.Replace(name))
}

func (s *CLIService) checkAppendAliasToFileCommand(ctx context.Context, line string, filePath string) bool {
	if strings.HasPrefix(line, "apal ") {
		alias := strings.TrimSpace(line[5:])
		_, err := s.aliasesService.AddFileAlias(ctx, filePath, alias)
		if err != nil {
			fmt.Printf("Error adding alias: %v\n", err)
			utils.WaitForUserInput()
			return false
		}
		return true
	}
	return false
}

func (s *CLIService) checkRemoveAliasFromFileCommand(ctx context.Context, line string, filePath string) bool {
	if strings.HasPrefix(line, "ral ") {
		aliasStr := strings.TrimSpace(line[4:])
		_, err := s.aliasesService.RemoveFileAlias(ctx, filePath, aliasStr)
		if err != nil {
			fmt.Printf("Error removing alias: %v\n", err)
			utils.WaitForUserInput()
			return false
		}
		return true
	}
	return false
}

func (s *CLIService) checkRemoveAliasFromContainerCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "ral ") {
		aliasStr := strings.TrimSpace(line[4:])
		_, err := s.aliasesService.RemoveContainerAlias(ctx, containerType, id, aliasStr)
		if err != nil {
			fmt.Printf("Error removing alias: %v\n", err)
			utils.WaitForUserInput()
			return false
		}
		return true
	}
	return false
}

func (s *CLIService) checkAppendAliasToContainerCommand(ctx context.Context, line string, containerType schema.ContainerType, id int) bool {
	if strings.HasPrefix(line, "apal ") {
		alias := strings.TrimSpace(line[5:])
		_, err := s.aliasesService.AddContainerAlias(ctx, containerType, id, alias)
		if err != nil {
			fmt.Printf("Error adding alias: %v\n", err)
			utils.WaitForUserInput()
			return false
		}
		return true
	}
	return false
}

func getDirectoryEntries(dirPath *string) ([]os.DirEntry, error) {
	if dirPath == nil {
		return []os.DirEntry{}, nil
	}

	entries, err := os.ReadDir(*dirPath)
	if err != nil {
		return []os.DirEntry{}, err
	}

	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di
		}
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

func (s *CLIService) printDirectoryContents(dirPath string) {
	entries, err := getDirectoryEntries(&dirPath)
	if err != nil {
		fmt.Printf("Error getting directory entries: %v\n", err)
		return
	}

	fmt.Println("\nFiles:")
	for index, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// mode := info.Mode()
		size := info.Size()
		modTime := info.ModTime()

		isDir := entry.IsDir()

		line := fmt.Sprintf("%3d %-30s %8d %s",
			index+1,
			entry.Name(),
			size,
			modTime.Format("2006-01-02 15:04:05"),
		)
		if isDir {
			color.New(color.FgWhite, color.BgBlue).Println(line)
		} else {
			fmt.Println(line)
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
	filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeProblem, p.ID)
	if filesDir != nil {
		fmt.Printf("\tFiles Directory: %s\n", *filesDir)
	}
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	if filesDir != nil {
		s.printDirectoryContents(*filesDir)
	}
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
	filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeQuestion, q.ID)
	if filesDir != nil {
		fmt.Printf("\tFiles Directory: %s\n", *filesDir)
	}
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	s.containerService.PrintQuestions(questions)
	if filesDir != nil {
		s.printDirectoryContents(*filesDir)
	}
	fmt.Println()
}

func (s *CLIService) PrintStoryInfo(
	ctx context.Context,
	st *ent.Story,
	subtasks []*ent.Task,
	problems []*ent.Problem,
	questions []*ent.Question,
	stories []*ent.Story,
	epics []*ent.Epic,
	aliases []*models.AliasModel,
) {
	s.containerService.PrintParentsPath(ctx, schema.ContainerTypeStory, st.ID)

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tStory ID: %d\n", st.ID)
	fmt.Printf("\tDescription: %s\n", st.Description)
	fmt.Printf("\tStatus: %s\n", map[bool]string{true: "Closed", false: "Open"}[st.Closed])
	if st.Notes != "" {
		fmt.Printf("Notes: %s\n", st.Notes)
	}
	if len(st.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(st.Tags, ", "))
	}
	if st.DoneDateTime != nil {
		fmt.Printf("Done Date: %s\n", st.DoneDateTime.Format("2006-01-02 15:04:05"))
	}
	filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeStory, st.ID)
	if filesDir != nil {
		fmt.Printf("\tFiles Directory: %s\n", *filesDir)
	}
	s.aliasesService.PrintAliases(aliases)
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	s.containerService.PrintQuestions(questions)
	s.containerService.PrintStories(stories)
	s.containerService.PrintEpics(epics)
	if filesDir != nil {
		s.printDirectoryContents(*filesDir)
	}
	fmt.Println()
}

func (s *CLIService) PrintEpicsInfo(epics []*ent.Epic) {

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tEpics All: %d\n")
	fmt.Println(strings.Repeat("=", 80))
	s.containerService.PrintEpics(epics)
	fmt.Println()
}

func (s *CLIService) PrintEpicInfo(
	ctx context.Context,
	e *ent.Epic,
	subtasks []*ent.Task,
	problems []*ent.Problem,
	questions []*ent.Question,
	stories []*ent.Story,
	epics []*ent.Epic,
	aliases []*models.AliasModel,
) {
	s.containerService.PrintParentsPath(ctx, schema.ContainerTypeEpic, e.ID)

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tEpic ID: %d\n", e.ID)
	fmt.Printf("\tDescription: %s\n", e.Description)
	fmt.Printf("\tStatus: %s\n", map[bool]string{true: "Closed", false: "Open"}[e.Closed])
	if e.Notes != "" {
		fmt.Printf("\tNotes: %s\n", e.Notes)
	}
	if len(e.Tags) > 0 {
		fmt.Printf("\tTags: %s\n", strings.Join(e.Tags, ", "))
	}
	if e.DoneDateTime != nil {
		fmt.Printf("\tDone Date: %s\n", e.DoneDateTime.Format("2006-01-02 15:04:05"))
	}
	filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeEpic, e.ID)
	if filesDir != nil {
		fmt.Printf("\tFiles Directory: %s\n", *filesDir)
	}
	s.aliasesService.PrintAliases(aliases)
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	s.containerService.PrintQuestions(questions)
	s.containerService.PrintStories(stories)
	s.containerService.PrintEpics(epics)

	if filesDir != nil {
		s.printDirectoryContents(*filesDir)
	}
	fmt.Println()
}

func (s *CLIService) PrintFileInfo(ctx context.Context, filePath string, aliases []*models.AliasModel) {

	fmt.Println(strings.Repeat("=", 100))
	color.Cyan("\tFile Path: " + filePath)
	s.aliasesService.PrintAliases(aliases)
	fmt.Println(strings.Repeat("=", 100))

	fmt.Println()
}

func (s *CLIService) printKnowledgeNodeInfo(
	ctx context.Context,
	kn *ent.KnowledgeNode,
	subtasks []*ent.Task,
	problems []*ent.Problem,
	questions []*ent.Question,
	stories []*ent.Story,
	epics []*ent.Epic,
	knowledgeNodes []*ent.KnowledgeNode,
	files []os.DirEntry,
	aliases []*models.AliasModel,
) {
	s.containerService.PrintParentsPath(ctx, schema.ContainerTypeKnowledgeNode, kn.ID)

	fmt.Println(strings.Repeat("=", 80))
	color.Cyan("\tKnowledge Node ID: %d\n", kn.ID)
	fmt.Printf("\tName: %s\n", kn.Name)
	if kn.Notes != "" {
		fmt.Printf("Notes: %s\n", kn.Notes)
	}
	if len(kn.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(kn.Tags, ", "))
	}
	filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeKnowledgeNode, kn.ID)
	s.aliasesService.PrintAliases(aliases)
	fmt.Println(strings.Repeat("=", 80))

	s.containerService.PrintSubtasks(subtasks)
	s.containerService.PrintProblems(problems)
	s.containerService.PrintQuestions(questions)
	s.containerService.PrintStories(stories)
	s.containerService.PrintEpics(epics)
	s.containerService.PrintKnowledgeNodes(knowledgeNodes)
	if filesDir != nil {
		s.printDirectoryContents(*filesDir)
	}

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

		if wasIt := s.checkOpenDirCommand(ctx, line, schema.ContainerTypeProblem, id); wasIt {
			continue
		}

		if wasIt := s.checkAddMindMapTemplateCommand(ctx, line, schema.ContainerTypeProblem, id); wasIt {
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
		case "q", "quit", "exit", "x":
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

		if wasIt := s.checkOpenDirCommand(ctx, line, schema.ContainerTypeQuestion, id); wasIt {
			continue
		}

		if wasIt := s.checkAddMindMapTemplateCommand(ctx, line, schema.ContainerTypeQuestion, id); wasIt {
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
		case "q", "quit", "exit", "x":
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

func (s *CLIService) ViewStoryInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id

	for {
		// Clear screen before printing
		utils.ClearScreen()
		story, err := s.client.Story.Get(ctx, currentID)
		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Story %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting story: %v\n", err)
			}
			return
		}
		subtasks, _ := s.containerService.GetOpenSubtasks(ctx, schema.ContainerTypeStory, currentID)
		problems, _ := s.containerService.GetOpenProblems(ctx, schema.ContainerTypeStory, currentID)
		questions, _ := s.containerService.GetOpenQuestions(ctx, schema.ContainerTypeStory, currentID)
		stories, _ := s.containerService.GetOpenStories(ctx, schema.ContainerTypeStory, currentID)
		epics, _ := s.containerService.GetOpenEpics(ctx, schema.ContainerTypeStory, currentID)
		aliases, _ := s.aliasesService.GetAliasesByTaskContainer(ctx, schema.ContainerTypeStory, currentID)

		s.PrintStoryInfo(ctx, story, subtasks, problems, questions, stories, epics, aliases)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		if goToNextIteration := s.checkAddTaskCommand(ctx, line, schema.ContainerTypeStory, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddProblemCommand(ctx, line, schema.ContainerTypeStory, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddQuestionCommand(ctx, line, schema.ContainerTypeStory, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddStoryCommand(ctx, line, schema.ContainerTypeStory, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddEpicCommand(ctx, line, schema.ContainerTypeStory, id); goToNextIteration {
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
		if wasIt := s.checkSelectStoryCommand(ctx, line, stories); wasIt {
			continue
		}
		if wasIt := s.checkSelectEpicCommand(ctx, line, epics); wasIt {
			continue
		}
		if wasIt := s.checkOpenDirCommand(ctx, line, schema.ContainerTypeStory, id); wasIt {
			continue
		}

		if wasIt := s.checkAddMindMapTemplateCommand(ctx, line, schema.ContainerTypeStory, id); wasIt {
			continue
		}

		if wasIt := s.checkAppendAliasToContainerCommand(ctx, line, schema.ContainerTypeStory, id); wasIt {
			continue
		}

		if wasIt := s.checkRemoveAliasFromContainerCommand(ctx, line, schema.ContainerTypeStory, id); wasIt {
			continue
		}

		if line == "u" {
			err := s.NavigateToParent(ctx, schema.ContainerTypeStory, currentID)
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
		case "q", "quit", "exit", "x":
			os.Exit(0)
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, 'p+ <description>' to add problem, 'q+ <description>' to add question, 's+ <description>' to add story, or 'e+ <description>' to add epic.\n", line)
			utils.WaitForUserInput()
		}
	}
}

func (s *CLIService) ViewEpicsInteractive(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		utils.ClearScreen()

		epics, err := s.epicsRepository.GetAllEpics(ctx)
		s.PrintEpicsInfo(epics)
		if err != nil {
			fmt.Printf("Error getting epics: %v\n", err)
			return
		}

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		if wasIt := s.checkSelectEpicCommand(ctx, line, epics); wasIt {
			continue
		}

		lineLower := strings.ToLower(line)
		switch lineLower {
		case "q", "quit", "exit", "x":
			os.Exit(0)
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, 'p+ <description>' to add problem, 'q+ <description>' to add question, 's+ <description>' to add story, or 'e+ <description>' to add epic.\n", line)
			utils.WaitForUserInput()
		}
	}
}

func (s *CLIService) ViewEpicInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id

	for {
		// Clear screen before printing
		utils.ClearScreen()
		epic, err := s.client.Epic.Get(ctx, currentID)
		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Epic %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting epic: %v\n", err)
			}
			return
		}
		subtasks, _ := s.containerService.GetOpenSubtasks(ctx, schema.ContainerTypeEpic, currentID)
		problems, _ := s.containerService.GetOpenProblems(ctx, schema.ContainerTypeEpic, currentID)
		questions, _ := s.containerService.GetOpenQuestions(ctx, schema.ContainerTypeEpic, currentID)
		stories, _ := s.containerService.GetOpenStories(ctx, schema.ContainerTypeEpic, currentID)
		epics, _ := s.containerService.GetOpenEpics(ctx, schema.ContainerTypeEpic, currentID)
		aliases, _ := s.aliasesService.GetAliasesByTaskContainer(ctx, schema.ContainerTypeEpic, currentID)

		s.PrintEpicInfo(ctx, epic, subtasks, problems, questions, stories, epics, aliases)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		if goToNextIteration := s.checkAddTaskCommand(ctx, line, schema.ContainerTypeEpic, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddProblemCommand(ctx, line, schema.ContainerTypeEpic, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddQuestionCommand(ctx, line, schema.ContainerTypeEpic, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddStoryCommand(ctx, line, schema.ContainerTypeEpic, id); goToNextIteration {
			continue
		}

		if goToNextIteration := s.checkAddEpicCommand(ctx, line, schema.ContainerTypeEpic, id); goToNextIteration {
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

		if wasIt := s.checkSelectStoryCommand(ctx, line, stories); wasIt {
			continue
		}

		if wasIt := s.checkSelectEpicCommand(ctx, line, epics); wasIt {
			continue
		}

		if wasIt := s.checkOpenDirCommand(ctx, line, schema.ContainerTypeEpic, id); wasIt {
			continue
		}

		if wasIt := s.checkAddMindMapTemplateCommand(ctx, line, schema.ContainerTypeEpic, id); wasIt {
			continue
		}

		if wasIt := s.checkAppendAliasToContainerCommand(ctx, line, schema.ContainerTypeEpic, id); wasIt {
			continue
		}

		if wasIt := s.checkRemoveAliasFromContainerCommand(ctx, line, schema.ContainerTypeEpic, id); wasIt {
			continue
		}

		if line == "u" {
			err := s.NavigateToParent(ctx, schema.ContainerTypeEpic, currentID)
			if err != nil {
				s.ViewEpicsInteractive(ctx)
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
		case "q", "quit", "exit", "x":
			os.Exit(0)
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, 'p+ <description>' to add problem, 'q+ <description>' to add question, 's+ <description>' to add story, or 'e+ <description>' to add epic.\n", line)
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
	case schema.ContainerTypeStory:
		s.ViewStoryInteractive(ctx, ID)
	case schema.ContainerTypeEpic:
		s.ViewEpicInteractive(ctx, ID)
	case schema.ContainerTypeKnowledgeNode:
		s.ViewKnowledgeNodeInteractive(ctx, ID)
	}
}

func (s *CLIService) ViewContainerByAlias(ctx context.Context, alias string) {
	aliasModel, err := s.aliasesRepository.GetAliasByAliasString(ctx, alias)

	if err != nil {
		print(err)
		return
	}

	// Handle file type aliases
	if aliasModel.Type == schema.AliasTypeFile {
		if aliasModel.FilePath != nil {
			fmt.Printf("File alias '%s' points to: %s\n", alias, *aliasModel.FilePath)
			utils.OpenFile(*aliasModel.FilePath)
			os.Exit(0)
			// TODO: Open file or show file contents
		} else {
			fmt.Printf("File alias '%s' has no file path specified\n", alias)
		}
		return
	}

	// Convert AliasType to ContainerType
	containerType, ok := aliasModel.Type.ToContainerType()
	if !ok {
		fmt.Printf("Invalid alias type: %s\n", aliasModel.Type)
		return
	}

	// Check if ItemID is provided
	if aliasModel.ItemID == nil {
		fmt.Printf("Alias '%s' has no item ID specified\n", alias)
		return
	}

	s.ViewContainerInteractive(ctx, containerType, *aliasModel.ItemID)
}

func (s *CLIService) ViewKnowledgeNodeInteractive(ctx context.Context, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id

	for {
		// Clear screen before printing
		utils.ClearScreen()
		knowledgeNode, err := s.client.KnowledgeNode.Get(ctx, currentID)
		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Knowledge Node %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting knowledge node: %v\n", err)
			}
			return
		}
		subtasks, _ := s.containerService.GetOpenSubtasks(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		problems, _ := s.containerService.GetOpenProblems(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		questions, _ := s.containerService.GetOpenQuestions(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		stories, _ := s.containerService.GetOpenStories(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		epics, _ := s.containerService.GetOpenEpics(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		knowledgeNodes, _ := s.containerService.GetOpenKnowledgeNodes(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		filesDir := s.containerService.GetFilesFolder(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		files, _ := getDirectoryEntries(filesDir)
		if err != nil {
			fmt.Printf("Error getting directory entries: %v\n", err)
			return
		}
		aliases, _ := s.aliasesService.GetAliasesByTaskContainer(ctx, schema.ContainerTypeKnowledgeNode, currentID)
		s.printKnowledgeNodeInfo(ctx, knowledgeNode, subtasks, problems, questions, stories, epics, knowledgeNodes, files, aliases)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}
		if goToNextIteration := s.checkAddTaskCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddProblemCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddQuestionCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddStoryCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddEpicCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); goToNextIteration {
			continue
		}
		if goToNextIteration := s.checkAddKnowledgeNodeCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); goToNextIteration {
			continue
		}
		// Prefer child knowledge-node selection (kn / n / bare number) over task index.
		if wasIt := s.checkSelectKnowledgeNodeCommand(ctx, line, knowledgeNodes); wasIt {
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

		if wasIt := s.checkSelectStoryCommand(ctx, line, stories); wasIt {
			continue
		}
		if wasIt := s.checkSelectEpicCommand(ctx, line, epics); wasIt {
			continue
		}

		if wasIt := s.checkOpenDirCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); wasIt {
			continue
		}

		if wasIt := s.checkAddMindMapTemplateCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); wasIt {
			continue
		}

		if wasIt := s.checkSelectFileCommand(ctx, line, filesDir, files); wasIt {
			continue
		}
		if wasIt := s.checkAppendAliasToContainerCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); wasIt {
			continue
		}
		if wasIt := s.checkRemoveAliasFromContainerCommand(ctx, line, schema.ContainerTypeKnowledgeNode, id); wasIt {
			continue
		}

		if line == "u" {
			err := s.NavigateToParent(ctx, schema.ContainerTypeKnowledgeNode, currentID)
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
		case "q", "quit", "exit", "x":
			os.Exit(0)
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, 't+ <description>' to add subtask, 'p+ <description>' to add problem, 'q+ <description>' to add question, 's+ <description>' to add story, 'e+ <description>' to add epic, or 'kn+ <name>' to add knowledge node.\n", line)
			utils.WaitForUserInput()
		}
	}
}

func (s *CLIService) ViewFileInteractive(ctx context.Context, filePath string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Clear screen before printing
		utils.ClearScreen()

		// Get task
		aliasModels, _ := s.aliasesService.GetAliasesByFilePath(ctx, filePath)
		s.PrintFileInfo(ctx, filePath, aliasModels)

		line := utils.GetUserInput(scanner)
		if line == "" {
			continue
		}

		if line == "u" {
			return
		}

		if line == "open" {
			utils.OpenFile(filePath)
			continue
		}

		if wasIt := s.checkAppendAliasToFileCommand(ctx, line, filePath); wasIt {
			continue
		}

		if wasIt := s.checkRemoveAliasFromFileCommand(ctx, line, filePath); wasIt {
			continue
		}

		lineLower := strings.ToLower(line)
		switch lineLower {
		case "q", "quit", "exit", "x":
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

func (s *CLIService) NavigateToParent(ctx context.Context, containerType schema.ContainerType, ID int) error {
	parentType, parentID, err := s.containerService.GetParentCommon(ctx, containerType, ID)
	if parentType == nil || err != nil {
		return errors.New("Error, cant get parent")
	}
	s.ViewContainerInteractive(ctx, *parentType, parentID)
	return nil
}
