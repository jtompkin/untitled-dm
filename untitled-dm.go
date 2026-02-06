package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// model implements tea.Model
type model struct {
	// choices is the list of choices that the user can choose from.
	choices []string
	// cmds maps indices in choices to executables to run.
	cmds map[int]struct {
		name string
		args []string
	}
	// err is any error that occurred while running a command from cmds.
	err error
	// cursor is the index of the current cursor position in choices.
	cursor int
	// selected is the index of selected item in choices. No items are selected if
	// negative
	selected int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ranMsg:
		return m, tea.Quit
	case errorMsg:
		m.err = msg
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor == 0 {
				m.cursor = len(m.choices) - 1
			} else {
				m.cursor--
			}
		case "down", "j":
			if m.cursor == len(m.choices)-1 {
				m.cursor = 0
			} else {
				m.cursor++
			}
		case " ":
			if m.cursor == m.selected {
				m.selected = -1
			} else {
				m.selected = m.cursor
			}
		case "enter":
			var cmd *exec.Cmd
			if c, ok := m.cmds[m.selected]; ok {
				cmd = exec.Command(c.name, c.args...)
			}
			return m, runShCmd(cmd)
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nCould not start selected option: %v\n\n", m.err)
	}
	var b strings.Builder
	b.WriteString("Where should we go?\n\n")
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if m.selected == i {
			checked = "x"
		}
		fmt.Fprintf(&b, "%s [%s] %s\n", cursor, checked, choice)
	}
	b.WriteString("\nPress q to quit.\n")
	return b.String()
}

func runShCmd(cmd *exec.Cmd) tea.Cmd {
	return func() tea.Msg {
		if cmd == nil {
			return ranMsg{}
		}
		if err := cmd.Run(); err != nil {
			return errorMsg{err}
		}
		return ranMsg{}
	}
}

// ranMsg indicates that the command successfully completed
type ranMsg struct{}

// errorMsg indicates that the command did not successfully complete
type errorMsg struct{ err error }

func (e errorMsg) Error() string { return e.err.Error() }

func initialModel() model {
	return model{
		choices: []string{"Xinit", "TTY"},
		cmds: map[int]struct {
			name string
			args []string
		}{
			0: {name: "startx"},
		},
		selected: -1,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
}
