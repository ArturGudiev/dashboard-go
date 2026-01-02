package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
