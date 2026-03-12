package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/karanshah229/gistsync/pkg/i18n"
)

var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	infoColor    = color.New(color.FgCyan)
	warningColor = color.New(color.FgYellow, color.Bold)
	headerColor  = color.New(color.FgHiBlue, color.Bold, color.Underline)

	output io.Writer = os.Stdout
	stdinReader *bufio.Reader
)

// GetSharedReader returns a shared buffered reader for os.Stdin
func GetSharedReader() *bufio.Reader {
	if stdinReader == nil {
		stdinReader = bufio.NewReader(os.Stdin)
	}
	return stdinReader
}

// SetOutput overrides the default output (os.Stdout)
func SetOutput(w io.Writer) {
	output = w
	successColor.SetWriter(w)
	errorColor.SetWriter(w)
	infoColor.SetWriter(w)
	warningColor.SetWriter(w)
	headerColor.SetWriter(w)
}

// Success prints a success message with a green checkmark
func Success(key string, data interface{}) {
	successColor.Fprintf(output, "✅ %s\n", i18n.T(key, data))
}

// Error prints an error message with a red cross
func Error(key string, data interface{}) {
	errorColor.Fprintf(output, "❌ %s\n", i18n.T(key, data))
}

// Info prints an info message with a blue lightbulb or info icon
func Info(key string, data interface{}) {
	infoColor.Fprintf(output, "💡 %s\n", i18n.T(key, data))
}

// Warning prints a warning message with a yellow warning triangle
func Warning(key string, data interface{}) {
	warningColor.Fprintf(output, "⚠️  %s\n", i18n.T(key, data))
}

// Header prints a header message with a blue bold underline
func Header(key string, data interface{}) {
	headerColor.Fprintf(output, "\n%s\n", i18n.T(key, data))
}

// Print prints a plain message, localized
func Print(key string, data interface{}) {
	fmt.Fprintln(output, i18n.T(key, data))
}

// Printf prints a formatted message (direct i18n wrap)
func Printf(key string, data interface{}) {
	fmt.Fprint(output, i18n.T(key, data))
}

// Confirm asks a localized yes/no question. Default is true (yes).
func Confirm(key string, data interface{}) bool {
	fmt.Fprint(output, i18n.T(key, data))
	reader := GetSharedReader()
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" || input == "y" || input == "yes" {
		return true
	}
	return false
}

