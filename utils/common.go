package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/niemeyer/pretty"
)

// isInteger checks if a string represents a valid integer
func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func WaitForUserInput() {
	fmt.Println("Press Enter to continue...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

func PrintAndWait(message string) {
	fmt.Println(message)
	WaitForUserInput()
}

// ClearScreen clears the terminal screen (cross-platform)
func ClearScreen() {
	// ANSI escape codes work on modern terminals (Windows 10+, Unix/Linux/Mac)
	fmt.Print("\033[2J\033[H")
}

func GetUserInput(scanner *bufio.Scanner) string {
	fmt.Print("> ")
	if !scanner.Scan() {
		return ""
	}
	line := strings.TrimSpace(scanner.Text())
	return line
}

// PrintObj prints an object structure with all keys and values using JSON formatting.
// Falls back to pretty print if JSON marshaling fails.
func PrintObj(obj interface{}, label string) {
	if label != "" {
		fmt.Printf("%s:\n", label)
	}
	jsonData, jsonErr := json.MarshalIndent(obj, "", "  ")
	if jsonErr == nil {
		fmt.Println(string(jsonData))
	} else {
		// Fallback to pretty print if JSON fails
		pretty.Printf("%+v\n", obj)
	}
}

func PrettyPrint(data any) {
	// data can be a struct, slice, or map
	b, err := json.MarshalIndent(data, "", "    ")
	if err == nil {
		fmt.Println(string(b))
	}
}

// OpenFile opens a file using the system's default application.
// On Windows: uses PowerShell to open the file
// On Linux: uses xdg-open
// On Mac OS: uses open
func OpenFile(filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Use PowerShell to open the file (Invoke-Item handles the file path)
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Invoke-Item '%s'", filePath))
	case "linux":
		// Use xdg-open for Linux
		cmd = exec.Command("xdg-open", filePath)
	case "darwin":
		// Use open for Mac OS
		cmd = exec.Command("open", filePath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Run()
}

// OpenDirectory opens a folder in the default file manager (Explorer / Finder / xdg-open).
func OpenDirectory(dirPath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dirPath)
	case "darwin":
		cmd = exec.Command("open", dirPath)
	case "linux":
		cmd = exec.Command("xdg-open", dirPath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return cmd.Run()
}

func SelectItemFromList(list []string) (*string, error) {
	if len(list) == 0 {
		return nil, errors.New("empty list")
	}
	items := list

	// Find: itemFunc maps index to display string. WithAlignTop pins the UI to the top of the terminal
	// (upstream go-fuzzyfinder anchors to the bottom by default).
	idx, err := fuzzyfinder.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithAlignTop())

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Выбор отменен")
			return nil, errors.New("selection aborted")
		}
		return nil, err
	}

	// 3. Получение результата по выбранному индексу
	fmt.Printf("Вы выбрали: %s\n", items[idx])
	return &items[idx], nil
}

func SelectIndexesFromList(list []string) ([]int, error) {
	if len(list) == 0 {
		return []int{}, nil
	}
	items := list

	// Find: itemFunc maps index to display string. WithAlignTop pins the UI to the top of the terminal
	// (upstream go-fuzzyfinder anchors to the bottom by default).
	indexes, err := fuzzyfinder.FindMulti(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithAlignTop())

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Выбор отменен")
			return nil, errors.New("selection aborted")
		}
		return nil, err
	}

	return indexes, nil
}
