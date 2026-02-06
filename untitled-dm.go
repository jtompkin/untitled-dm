package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	// model implements tea.Model
	model struct {
		// choices is the list of choices that the user can choose from.
		choices []string
		// cmds maps indices in choices to executables to run.
		cmds map[int]command
		// output contains the combined standard streams of the command and any error that occurred
		output struct {
			std string
			err error
		}
		// cursor is the index of the current cursor position in choices.
		cursor int
		// selected is the index of selected item in choices. No items are selected if
		// negative
		selected int
	}
	command struct {
		// name is the file name of the command
		name string
		args []string
	}
	// config is the TOML structure of the configuration file
	config struct {
		Commands []struct {
			// Name is the string that is presented in the menu
			Name string
			// Command is the file name to be run
			Command string
			// Args are passed to Command
			Args []string
		}
	}
	// ranMsg indicates that the command successfully completed
	ranMsg struct{ output string }
	// errorMsg indicates that the command did not successfully complete
	errorMsg struct {
		output string
		err    error
	}
)

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ranMsg:
		m.output.std = msg.output
		return m, tea.Quit
	case errorMsg:
		m.output.std = msg.output
		m.output.err = msg.err
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
	if m.output.err != nil {
		return fmt.Sprintf("\nCould not start selected option: %v\n\n%s", m.output.err, m.output.std)
	}
	b := new(strings.Builder)
	b.WriteString("Where should we go?\nPress enter to confirm selection.\n\n")
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if m.selected == i {
			checked = "x"
		}
		fmt.Fprintf(b, "%s [%s] %s\n", cursor, checked, choice)
	}
	fmt.Fprintf(b, "\nPress q to quit.\n%s", m.output.std)
	return b.String()
}

func runShCmd(cmd *exec.Cmd) tea.Cmd {
	return func() tea.Msg {
		if cmd == nil {
			return ranMsg{"No command to run\n"}
		}
		o, err := cmd.CombinedOutput()
		if err != nil {
			return errorMsg{output: string(o), err: err}
		}
		return ranMsg{string(o)}
	}
}

func parseTomlConfig(f string) (conf config) {
	if _, err := os.Stat(f); err != nil {
		return conf
	}
	if _, err := toml.DecodeFile(f, &conf); err != nil {
		log.Fatal(err)
	}
	return conf
}

func main() {
	var confFlag = flag.String("c", "config.toml", "Path to configuration file")
	flag.Parse()
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.txt", "debug")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}
	conf := parseTomlConfig(*confFlag)
	m := model{
		choices:  make([]string, len(conf.Commands)),
		cmds:     make(map[int]command),
		selected: -1,
	}
	for i, c := range conf.Commands {
		m.choices[i] = c.Name
		if c.Command != "" {
			m.cmds[i] = command{
				name: c.Command,
				args: c.Args,
			}
		}
	}
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
}
