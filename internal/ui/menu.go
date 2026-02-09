package ui

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Menu represents an interactive menu.
type Menu struct {
	Title   string
	Items   []MenuItem
	ShowNum bool // Whether to show numbers
}

// MenuItem represents a single menu item.
type MenuItem struct {
	Label       string
	Description string
	Action      func() error
	SubMenu     *Menu
}

// NewMenu creates a new menu with the given title.
func NewMenu(title string) *Menu {
	return &Menu{
		Title:   title,
		Items:   make([]MenuItem, 0),
		ShowNum: true,
	}
}

// Add adds a menu item with optional action.
func (m *Menu) Add(label string, action ...func() error) *Menu {
	item := MenuItem{Label: label}
	if len(action) > 0 {
		item.Action = action[0]
	}
	m.Items = append(m.Items, item)
	return m
}

// AddWithDesc adds a menu item with description and optional action.
func (m *Menu) AddWithDesc(label, desc string, action ...func() error) *Menu {
	item := MenuItem{
		Label:       label,
		Description: desc,
	}
	if len(action) > 0 {
		item.Action = action[0]
	}
	m.Items = append(m.Items, item)
	return m
}

// AddItem adds a complete menu item.
func (m *Menu) AddItem(item MenuItem) *Menu {
	m.Items = append(m.Items, item)
	return m
}

// AddSubMenu adds a submenu.
func (m *Menu) AddSubMenu(label string, subMenu *Menu) *Menu {
	m.Items = append(m.Items, MenuItem{
		Label:   label,
		SubMenu: subMenu,
	})
	return m
}

// Display shows the menu and handles input. Returns the selected index or -1 for back/exit.
func (m *Menu) Display() int {
	ClearScreen()
	DisplayCompactBanner()

	Box(m.Title)
	NewLine()

	ColorPrintln(BrightCyan, "Options:")
	for i, item := range m.Items {
		if item.Description != "" {
			PrintMenuItemWithDesc(i+1, item.Label, item.Description)
		} else {
			PrintMenuItem(i+1, item.Label)
		}
	}

	// Add back/exit option
	PrintMenuBack(len(m.Items) + 1)

	choice := m.getChoice()
	return choice
}

// DisplayInline shows the menu without clearing screen. Use when caller handles screen setup.
func (m *Menu) DisplayInline() int {
	Box(m.Title)
	NewLine()

	ColorPrintln(BrightCyan, "Options:")
	for i, item := range m.Items {
		if item.Description != "" {
			PrintMenuItemWithDesc(i+1, item.Label, item.Description)
		} else {
			PrintMenuItem(i+1, item.Label)
		}
	}

	// Add back/exit option
	PrintMenuBack(len(m.Items) + 1)

	choice := m.getChoice()
	return choice
}

// Run displays the menu and executes the selected action. Loops until exit.
func (m *Menu) Run() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		ClearScreen()
		DisplayCompactBanner()

		Box(m.Title)
		NewLine()

		ColorPrintln(BrightCyan, "Options:")
		for i, item := range m.Items {
			if item.Description != "" {
				PrintMenuItemWithDesc(i+1, item.Label, item.Description)
			} else {
				PrintMenuItem(i+1, item.Label)
			}
		}

		// Add back/exit option
		exitNum := len(m.Items) + 1
		ColorPrint(BrightWhite, "  %d. ", exitNum)
		Println("Back/Exit")
		NewLine()

		ColorPrint(BrightCyan, "→ Choose: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil {
			Error("Invalid choice. Please enter a number.")
			WaitForEnter()
			continue
		}

		if choice == exitNum {
			return nil // Exit menu
		}

		if choice < 1 || choice > len(m.Items) {
			Error("Invalid choice. Please select 1-%d", exitNum)
			WaitForEnter()
			continue
		}

		item := m.Items[choice-1]

		// Handle submenu
		if item.SubMenu != nil {
			if err := item.SubMenu.Run(); err != nil {
				return err
			}
			continue
		}

		// Handle action
		if item.Action != nil {
			ClearScreen()
			DisplayCompactBanner()

			if err := item.Action(); err != nil {
				Error("Error: %v", err)
			}
			WaitForEnter()
		}
	}
}

// getChoice reads user input and returns the choice.
func (m *Menu) getChoice() int {
	reader := bufio.NewReader(os.Stdin)
	ColorPrint(BrightCyan, "→ Choose: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}

	exitNum := len(m.Items) + 1
	if choice == exitNum {
		return -1
	}

	if choice < 1 || choice > len(m.Items) {
		return -1
	}

	return choice - 1
}
