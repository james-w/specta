package specta

import "github.com/fatih/color"

// Color functions for matcher output.
// These respect NO_COLOR, FORCE_COLOR env vars and TTY detection automatically.
var (
	colorizeGreen = color.New(color.FgGreen).SprintFunc()
	colorizeRed   = color.New(color.FgRed).SprintFunc()
	colorizeGrey  = color.New(color.FgHiBlack).SprintFunc()
)

// colorize wraps text with the appropriate color.
func colorize(text string, colorFn func(...interface{}) string) string {
	return colorFn(text)
}
