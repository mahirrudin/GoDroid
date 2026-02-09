package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Banner ASCII art
const banner = `
    ██████╗  ██████╗ ██████╗ ██████╗  ██████╗ ██╗██████╗ 
   ██╔════╝ ██╔═══██╗██╔══██╗██╔══██╗██╔═══██╗██║██╔══██╗
   ██║  ███╗██║   ██║██║  ██║██████╔╝██║   ██║██║██║  ██║
   ██║   ██║██║   ██║██║  ██║██╔══██╗██║   ██║██║██║  ██║
   ╚██████╔╝╚██████╔╝██████╔╝██║  ██║╚██████╔╝██║██████╔╝
    ╚═════╝  ╚═════╝ ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═════╝ 
`

// Tagline displayed under the banner.
const tagline = "Android Security Automation | v2.0.0"

// ClearScreen clears the terminal screen.
func ClearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// DisplayBanner shows the GoDroid banner.
func DisplayBanner() {
	ColorPrintln(BrightRed, banner)
	ColorPrintln(BrightYellow+Bold, "          "+tagline)
	NewLine()
}

// DisplayStartupAnimation shows an animated startup sequence.
func DisplayStartupAnimation() {
	frames := []struct {
		color   string
		message string
	}{
		{BrightRed, "Initializing..."},
		{BrightCyan, "Loading Systems..."},
		{BrightGreen, "Systems Online!"},
	}

	for _, frame := range frames {
		ClearScreen()
		ColorPrintln(frame.color, banner)
		NewLine()
		ColorPrintln(frame.color+Bold, "        %s", frame.message)
		time.Sleep(600 * time.Millisecond)
	}

	time.Sleep(400 * time.Millisecond)
}

// DisplayCompactBanner shows a smaller banner for menus.
func DisplayCompactBanner() {
	ColorPrintln(BrightRed+Bold, "╔═══════════════════════════════════════════════════╗")
	ColorPrintln(BrightRed+Bold, "║              🤖 GoDroid v2.0.0                    ║")
	ColorPrintln(BrightRed+Bold, "╚═══════════════════════════════════════════════════╝")
	NewLine()
}

// StatusIcon represents different status icons
type StatusIcon string

const (
	IconCheck    StatusIcon = "✓"
	IconCross    StatusIcon = "✖"
	IconWarning  StatusIcon = "⚠"
	IconInfo     StatusIcon = "ℹ"
	IconArrow    StatusIcon = "→"
	IconDot      StatusIcon = "•"
	IconRocket   StatusIcon = "🚀"
	IconLock     StatusIcon = "🔐"
	IconUnlock   StatusIcon = "🔓"
	IconAndroid  StatusIcon = "🤖"
	IconPhone    StatusIcon = "📱"
	IconDownload StatusIcon = "⬇"
	IconUpload   StatusIcon = "⬆"
	IconFolder   StatusIcon = "📁"
	IconFile     StatusIcon = "📄"
	IconGear     StatusIcon = "⚙"
	IconSearch   StatusIcon = "🔍"
)

// PrintStatus prints a status line with icon.
func PrintStatus(icon StatusIcon, color, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ColorPrintln(color, "[%s] %s", icon, msg)
}

// PrintMenuItem prints a menu item.
func PrintMenuItem(number int, text string) {
	ColorPrint(BrightWhite, "  %d. ", number)
	Println(text)
}

// PrintMenuItemWithDesc prints a menu item with description.
func PrintMenuItemWithDesc(number int, text, desc string) {
	ColorPrint(BrightWhite, "  %d. ", number)
	Println(text)
	ColorPrintln(BrightMagenta, "     - %s", desc)
}

// PrintMenuBack prints the "Back" menu option.
func PrintMenuBack(number int) {
	ColorPrint(BrightWhite, "  %d. ", number)
	Println("Back")
	NewLine()
}

// PrintMenuExit prints the "Exit" menu option.
func PrintMenuExit(number int) {
	ColorPrint(BrightWhite, "  %d. ", number)
	Println("Exit")
	NewLine()
}

// WaitForEnter waits for user to press Enter.
func WaitForEnter(message ...string) {
	msg := "Press Enter to continue..."
	if len(message) > 0 {
		msg = message[0]
	}
	ColorPrint(BrightCyan, "→ %s", msg)
	fmt.Scanln()
}

// GetInput prompts for user input and returns the value.
func GetInput(prompt string) string {
	ColorPrint(BrightCyan, "→ %s: ", prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

// GetInputDefault prompts for user input with a default value.
func GetInputDefault(prompt, defaultVal string) string {
	ColorPrint(BrightCyan, "→ %s [%s]: ", prompt, defaultVal)
	var input string
	fmt.Scanln(&input)
	if input == "" {
		return defaultVal
	}
	return input
}

// Confirm prompts for yes/no confirmation.
func Confirm(prompt string) bool {
	ColorPrint(BrightCyan, "→ %s (y/n): ", prompt)
	var input string
	fmt.Scanln(&input)
	return input == "y" || input == "Y" || input == "yes" || input == "Yes"
}

// SelectFromList displays a list and returns the selected index (0-based).
func SelectFromList(title string, items []string) int {
	Box(title)
	NewLine()

	for i, item := range items {
		PrintMenuItem(i+1, item)
	}
	NewLine()

	for {
		input := GetInput("Choose")
		var choice int
		if _, err := fmt.Sscanf(input, "%d", &choice); err == nil {
			if choice >= 1 && choice <= len(items) {
				return choice - 1
			}
		}
		Error("Invalid choice. Please select 1-%d", len(items))
	}
}
