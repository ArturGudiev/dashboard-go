package main

import (
	"arturgudiev/dashboard/app"
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/task"
	"arturgudiev/dashboard/utils"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ddddddO/gtree"
	"github.com/fatih/color"
)

// runCLI starts the interactive CLI interface
func runCLI(application *app.App) {
	// Parse flags - create a new flag set to handle arguments after "cli"
	// This handles: "program cli --task 123"
	fs := flag.NewFlagSet("cli", flag.ContinueOnError)
	taskID := fs.Int("task", 0, "View task by ID interactively")

	// Parse arguments, skipping "cli" if it's the first arg
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "cli" || args[0] == "interactive") {
		args = args[1:]
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		log.Fatalf("Error parsing flags: %v", err)
	}

	ctx := context.Background()

	// If --task flag is provided, enter interactive task view
	if *taskID > 0 {
		viewTaskInteractive(ctx, application, *taskID)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Dashboard CLI ===")
	fmt.Println("Type 'help' for available commands")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "help", "h":
			printHelp()
		case "list", "ls":
			listTasks(ctx, application, args)
		case "get":
			getTask(ctx, application, args)
		case "create", "new":
			createTask(ctx, application, args)
		case "update":
			updateTask(ctx, application, args)
		case "finish":
			finishTask(ctx, application, args)
		case "delete", "rm":
			deleteTask(ctx, application, args)
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  help, h              - Show this help message")
	fmt.Println("  list, ls              - List all tasks")
	fmt.Println("  list done             - List done tasks")
	fmt.Println("  list open             - List open tasks")
	fmt.Println("  get <id>              - Get task by ID")
	fmt.Println("  create <description>  - Create a new task")
	fmt.Println("  update <id>           - Update a task (interactive)")
	fmt.Println("  finish <id>           - Finish a task recursively")
	fmt.Println("  delete <id>, rm <id> - Delete a task")
	fmt.Println("  quit, exit, q        - Exit the CLI")
	fmt.Println()
}

func listTasks(ctx context.Context, application *app.App, args []string) {
	var tasks []*ent.Task
	var err error

	if len(args) > 0 {
		filter := args[0]
		switch filter {
		case "done":
			tasks, err = application.Client.Task.Query().Where(task.DoneEQ(true)).All(ctx)
		case "open":
			tasks, err = application.Client.Task.Query().Where(task.DoneEQ(false)).All(ctx)
		default:
			fmt.Printf("Unknown filter: %s. Use 'done' or 'open'.\n", filter)
			return
		}
	} else {
		tasks, err = application.Client.Task.Query().All(ctx)
	}

	if err != nil {
		fmt.Printf("Error listing tasks: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Printf("\nFound %d task(s):\n", len(tasks))
	fmt.Println(strings.Repeat("-", 80))
	for _, t := range tasks {
		status := "Open"
		if t.Done {
			status = "Done"
		}
		fmt.Printf("ID: %d | %s | %s\n", t.ID, status, t.Description)
		if t.Notes != "" {
			fmt.Printf("    Notes: %s\n", t.Notes)
		}
		if len(t.Tags) > 0 {
			fmt.Printf("    Tags: %s\n", strings.Join(t.Tags, ", "))
		}
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()
}

// printTask prints task details (used by both flag and interactive command)
func printTask(ctx context.Context, application *app.App, id int) {
	t, err := application.Client.Task.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			fmt.Printf("Task %d not found.\n", id)
		} else {
			fmt.Printf("Error getting task: %v\n", err)
		}
		return
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Task-%d %s\n", t.ID, t.Description)
	fmt.Printf("Description: %s\n", t.Description)
	fmt.Printf("Status: %s\n", map[bool]string{true: "Done", false: "Open"}[t.Done])
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
}

// clearScreen clears the terminal screen (cross-platform)
func clearScreen() {
	// ANSI escape codes work on modern terminals (Windows 10+, Unix/Linux/Mac)
	fmt.Print("\033[2J\033[H")
}

// printTaskInfo prints task information in a formatted way
func printTaskInfo(ctx context.Context, application *app.App, t *ent.Task) {

	// Build and print parents path as a tree
	parentsPath := application.TaskService.GetParentsPath(ctx, t)
	if len(parentsPath) > 0 {
		// Reverse the path to go from root to immediate parent
		for i, j := 0, len(parentsPath)-1; i < j; i, j = i+1, j-1 {
			parentsPath[i], parentsPath[j] = parentsPath[j], parentsPath[i]
		}

		// Build tree structure
		root := gtree.NewRoot(parentsPath[0].Description)
		currentNode := root

		// Add intermediate parents
		for i := 1; i < len(parentsPath); i++ {
			currentNode = currentNode.Add(parentsPath[i].Description)
		}

		// Add current task as the final child
		currentNode.Add(t.Description)

		// Print the tree
		if err := gtree.OutputFromRoot(os.Stdout, root); err != nil {
			fmt.Printf("Error printing tree: %v\n", err)
		}
		fmt.Println()
	}

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

	// Get and display child tasks
	childRelations, err := application.Client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(containerchild.ParentTypeTask),
			containerchild.ParentID(t.ID),
			containerchild.ChildTypeEQ(containerchild.ChildTypeTask),
		).
		Order(containerchild.ByChildOrder()).
		WithChild().
		All(ctx)

	if err == nil && len(childRelations) > 0 {
		fmt.Println("\nChild Tasks:")
		for i, relation := range childRelations {
			childTask := relation.Edges.Child
			if childTask != nil {
				status := "Open"
				if childTask.Done {
					status = "Done"
				}
				fmt.Printf("  %d. [ID: %d] %s - %s\n", i+1, childTask.ID, status, childTask.Description)
			}
		}
	}

	fmt.Println("\nCommands: q/quit/exit - quit, r/refresh - refresh task, t+ <description> - add subtask")
}

// viewTaskInteractive shows task details and allows interactive commands
func viewTaskInteractive(ctx context.Context, application *app.App, id int) {
	scanner := bufio.NewScanner(os.Stdin)
	currentID := id // Track current task ID

	for {
		// Clear screen before printing
		clearScreen()

		// Get task
		t, err := application.Client.Task.Get(ctx, currentID)
		if err != nil {
			if ent.IsNotFound(err) {
				fmt.Printf("Task %d not found.\n", currentID)
			} else {
				fmt.Printf("Error getting task: %v\n", err)
			}
			return
		}

		// Print task information
		printTaskInfo(ctx, application, t)

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
			if err := addSubtask(ctx, application, currentID, description); err != nil {
				fmt.Printf("Error adding subtask: %v\n", err)
				utils.WaitForUserInput()
				continue
			}
			// Continue loop to refresh and show the new subtask
			continue
		}

		if utils.IsInteger(line) {
			index, err := strconv.Atoi(line)
			if err != nil {
				continue
			}

			// Get all child subtasks from current task
			childTasks, err := application.GetChildSubtasks(currentID)
			if err != nil {
				fmt.Printf("Error getting child tasks: %v\n", err)
				utils.WaitForUserInput()
				continue
			}

			// Validate index
			if index < 1 || index > len(childTasks) {
				fmt.Printf("Invalid index. Please enter a number between 1 and %d.\n", len(childTasks))
				utils.WaitForUserInput()
				continue
			}

			// Navigate to child task by updating currentID and continuing loop
			currentID = childTasks[index-1].ID
			continue // Loop will restart with new currentID
		}

		if line == "u" {
			parent := application.TaskService.GetParent(ctx, t)
			if parent == nil {
				fmt.Println("This task has no parent.")
				utils.WaitForUserInput()
				continue
			}
			// Navigate to parent task by updating currentID and continuing loop
			currentID = parent.ID
			continue // Loop will restart with new currentID
		}

		if strings.HasPrefix(line, "ft ") {
			indexStr := strings.TrimSpace(line[3:])
			if indexStr == "" {
				fmt.Println("Error: index required. Usage: ft <id>")
				utils.WaitForUserInput()
				continue
			}

			tasks, _ := application.TaskService.GetChildSubtasks(ctx, id)
			var taskDescriptions []string
			mapper := func(el *ent.Task) string { return el.Description }
			for _, el := range tasks {
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
			return
		case "r", "refresh":
			// Continue loop to refresh (screen will be cleared at start of loop)
			continue
		case "":
			// Empty input, refresh
			continue
		default:
			fmt.Printf("Unknown command: %s. Type 'q' to quit, 'r' to refresh, or 't+ <description>' to add subtask.\n", line)
			utils.WaitForUserInput()
		}
	}
}

func getTask(ctx context.Context, application *app.App, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: get <id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %s\n", args[0])
		return
	}

	printTask(ctx, application, id)
	fmt.Println()
}

func createTask(ctx context.Context, application *app.App, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: create <description>")
		return
	}

	description := strings.Join(args, " ")
	newTask, err := application.Client.Task.Create().
		SetDescription(description).
		SetDone(false).
		SetTags([]string{}).
		SetNotes("").
		Save(ctx)

	if err != nil {
		fmt.Printf("Error creating task: %v\n", err)
		return
	}

	fmt.Printf("Task created successfully! ID: %d\n", newTask.ID)
}

func updateTask(ctx context.Context, application *app.App, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: update <id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %s\n", args[0])
		return
	}

	// Get existing task
	t, err := application.Client.Task.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			fmt.Printf("Task %d not found.\n", id)
		} else {
			fmt.Printf("Error getting task: %v\n", err)
		}
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	updater := application.Client.Task.UpdateOneID(id)

	fmt.Printf("Current description: %s\n", t.Description)
	fmt.Print("New description (press Enter to keep current): ")
	if scanner.Scan() {
		newDesc := strings.TrimSpace(scanner.Text())
		if newDesc != "" {
			updater = updater.SetDescription(newDesc)
		}
	}

	fmt.Printf("Current notes: %s\n", t.Notes)
	fmt.Print("New notes (press Enter to keep current): ")
	if scanner.Scan() {
		newNotes := strings.TrimSpace(scanner.Text())
		if newNotes != "" {
			updater = updater.SetNotes(newNotes)
		}
	}

	fmt.Printf("Current status: %s\n", map[bool]string{true: "Done", false: "Open"}[t.Done])
	fmt.Print("Mark as done? (y/n, press Enter to keep current): ")
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer == "y" || answer == "yes" {
			updater = updater.SetDone(true)
		} else if answer == "n" || answer == "no" {
			updater = updater.SetDone(false)
		}
	}

	updatedTask, err := updater.Save(ctx)
	if err != nil {
		fmt.Printf("Error updating task: %v\n", err)
		return
	}

	fmt.Printf("Task %d updated successfully!\n", updatedTask.ID)
}

func finishTask(ctx context.Context, application *app.App, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: finish <id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %s\n", args[0])
		return
	}

	t, err := application.Client.Task.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			fmt.Printf("Task %d not found.\n", id)
		} else {
			fmt.Printf("Error getting task: %v\n", err)
		}
		return
	}

	if err := application.FinishTaskRecursively(t); err != nil {
		fmt.Printf("Error finishing task: %v\n", err)
		return
	}

	fmt.Printf("Task %d and all its descendants have been marked as done.\n", id)
}

func deleteTask(ctx context.Context, application *app.App, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: delete <id> or rm <id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid ID: %s\n", args[0])
		return
	}

	// Check if task exists
	_, err = application.Client.Task.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			fmt.Printf("Task %d not found.\n", id)
		} else {
			fmt.Printf("Error getting task: %v\n", err)
		}
		return
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete task %d? (yes/no): ", id)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "yes" && answer != "y" {
		fmt.Println("Deletion cancelled.")
		return
	}

	err = application.Client.Task.DeleteOneID(id).Exec(ctx)
	if err != nil {
		fmt.Printf("Error deleting task: %v\n", err)
		return
	}

	fmt.Printf("Task %d deleted successfully.\n", id)
}

// addSubtask creates a new subtask for the given parent task
func addSubtask(ctx context.Context, application *app.App, parentID int, description string) error {
	newTask, err := application.AddSubtask(parentID, description)
	if err != nil {
		return err
	}
	fmt.Printf("Subtask created successfully! ID: %d\n", newTask.ID)
	return nil
}
