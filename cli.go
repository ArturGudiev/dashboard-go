package main

import (
	"arturgudiev/dashboard/app"
	"arturgudiev/dashboard/ent"
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
)

func runCLI(application *app.App) {
	args := os.Args[1:]

	// Check if first arg is "epics" or second arg (after "cli") is "epics"
	if len(args) > 0 && args[0] == "epics" {
		ctx := context.Background()
		application.CLIService.ViewEpicsInteractive(ctx)
		return
	}
	// if len(args) == 2 && args[0] == "knowledge-node" {
	// 	if id, err := strconv.Atoi(args[1]); err == nil {
	// 		ctx := context.Background()
	// 		application.CLIService.ViewKnowledgeNodeInteractive(ctx, id)
	// 		return
	// 	}
	// }
	if len(args) > 1 && (args[0] == "cli" || args[0] == "interactive") && args[1] == "epics" {
		ctx := context.Background()
		application.CLIService.ViewEpicsInteractive(ctx)
		return
	}

	fs := flag.NewFlagSet("cli", flag.ContinueOnError)
	taskID := fs.Int("task", 0, "View task by ID interactively")
	knowledgeNodeID := fs.Int("knowledge-node", 0, "View knowledge node by ID interactively")
	epicInteractiveID := fs.Int("epic-interactive", 0, "View epic by ID interactively")
	tempFlag := fs.Bool("temp", false, "Run temporary method")
	alias := fs.String("alias", "", "Open item interactive by alias")

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

	// If --temp flag is provided, run temporary method
	if *tempFlag {
		tempMethod(ctx, application)
		return
	}

	if *alias != "" {
		application.CLIService.ViewContainerByAlias(ctx, *alias)
	}

	// If --epic-interactive flag is provided, enter interactive epic view
	if *epicInteractiveID > 0 {
		application.CLIService.ViewEpicInteractive(ctx, *epicInteractiveID)
		return
	}

	// If --knowledge-node flag is provided, enter interactive knowledge node view
	if *knowledgeNodeID > 0 {
		application.CLIService.ViewKnowledgeNodeInteractive(ctx, *knowledgeNodeID)
		return
	}

	// If --task flag is provided, enter interactive task view
	if *taskID > 0 {
		application.CLIService.ViewTaskInteractive(ctx, *taskID)
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
	fmt.Printf("Task-%d %s\n", t.ID, t.Description)
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

// tempMethod is a temporary method for testing
func tempMethod(ctx context.Context, a *app.App) {
	arr := []string{"яблоко", "банан", "апельсин", "груша"}
	selectedIndexes, err := utils.SelectIndexesFromList(arr)
	if err != nil {
		fmt.Printf("Error selecting item: %v\n", err)
		return
	}
	fmt.Printf("Selected indexes: %v\n", selectedIndexes)
	return
}
