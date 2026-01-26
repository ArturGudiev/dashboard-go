package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

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
