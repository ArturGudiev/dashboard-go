package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
