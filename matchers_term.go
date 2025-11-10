package specta

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	colorReset = 0
	colorRed   = 31
	colorGreen = 32
	colorGrey  = 90
)

// isTerminal checks if stdout is connected to a terminal (TTY).
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	// Check if it's a character device (terminal)
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// colorize wraps text with ANSI color codes if stdout is a terminal.
func colorize(text string, colorCode int) string {
	if !isTerminal() {
		return text
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[%dm", colorCode, text, colorReset)
}
