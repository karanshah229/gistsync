package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks a yes/no question. Default is true (yes).
func Confirm(message string) bool {
	fmt.Printf("%s (yes / no) ", message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" || input == "y" || input == "yes" {
		return true
	}
	return false
}
