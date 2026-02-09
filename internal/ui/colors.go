// Package ui provides terminal UI utilities including colors, menus, and banners.
package ui

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	// Colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
)

var colorsEnabled = true

// Init initializes the UI, enabling ANSI colors on Windows if needed.
func Init() {
	if runtime.GOOS == "windows" {
		// Enable virtual terminal processing on Windows
		enableWindowsVT()
	}

	// Check if we're in a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		colorsEnabled = false
	}
}

// enableWindowsVT enables virtual terminal processing on Windows.
func enableWindowsVT() {
	// This is handled automatically in Go 1.14+ for most cases
	// Additional handling would require syscalls
}

// DisableColors disables color output.
func DisableColors() {
	colorsEnabled = false
}

// EnableColors enables color output.
func EnableColors() {
	colorsEnabled = true
}

// colorize wraps text in color codes if colors are enabled.
func colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + Reset
}

// Success prints a success message in green.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(BrightGreen, "[✓] "+msg))
}

// Error prints an error message in red.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(BrightRed, "[✖] "+msg))
}

// Warning prints a warning message in yellow.
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(BrightYellow, "[!] "+msg))
}

// Info prints an info message in cyan.
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(BrightCyan, "[•] "+msg))
}

// Tip prints a tip message in magenta.
func Tip(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(BrightMagenta, "→ Tip: "+msg))
}

// Highlight prints highlighted text.
func Highlight(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(Bold+BrightWhite, msg))
}

// Print prints a regular message.
func Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Println prints a regular message with newline.
func Println(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// ColorPrint prints text with the specified color.
func ColorPrint(color, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Print(colorize(color, msg))
}

// ColorPrintln prints text with the specified color and newline.
func ColorPrintln(color, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorize(color, msg))
}

// Box prints text in a box.
func Box(title string) {
	width := len(title) + 4
	border := strings.Repeat("═", width)

	ColorPrintln(BrightCyan, "╔%s╗", border)
	ColorPrintln(BrightCyan, "║  %s  ║", title)
	ColorPrintln(BrightCyan, "╚%s╝", border)
}

// Separator prints a horizontal separator.
func Separator() {
	if colorsEnabled {
		fmt.Println(Dim + strings.Repeat("─", 50) + Reset)
	} else {
		fmt.Println(strings.Repeat("-", 50))
	}
}

// NewLine prints an empty line.
func NewLine() {
	fmt.Println()
}

// ProgressBar displays a simple progress bar.
func ProgressBar(current, total int64, width int) string {
	if total <= 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%%", bar, percent*100)
}

// Spinner characters for animation.
var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner represents an animated spinner.
type Spinner struct {
	index   int
	message string
}

// NewSpinner creates a new spinner with a message.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		index:   0,
		message: message,
	}
}

// Next advances the spinner and returns the display string.
func (s *Spinner) Next() string {
	char := spinnerChars[s.index%len(spinnerChars)]
	s.index++
	return fmt.Sprintf("%s %s", colorize(BrightCyan, char), s.message)
}

// ClearLine clears the current line.
func ClearLine() {
	fmt.Print("\r\033[K")
}

// MoveCursorUp moves the cursor up n lines.
func MoveCursorUp(n int) {
	if n > 0 {
		fmt.Printf("\033[%dA", n)
	}
}

// HideCursor hides the cursor.
func HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor shows the cursor.
func ShowCursor() {
	fmt.Print("\033[?25h")
}
