package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.0.1"

type (
	// model implements tea.Model
	model struct {
		// choices is the list of choices that the user can choose from.
		choices []string
		// cmds maps indices in choices to executables to run.
		cmds        map[int]command
		quitOnError bool
		// output contains the combined standard streams of the command and any error that occurred
		output struct {
			std string
			err error
		}
		// cursor is the index of the current cursor position in choices.
		cursor int16
		// selected is the index of selected item in choices. No items are selected if
		// negative
		selected int16
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
	// errorMsg indicates that the command did not successfully complete. Implements error
	errorMsg struct{ output string }
)

func (e errorMsg) Error() string { return e.output }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ranMsg:
		m.output.err = nil
		m.output.std = msg.output
		return m, nil
	case errorMsg:
		m.output.err = msg
		var cmd tea.Cmd
		if m.quitOnError {
			cmd = tea.Quit
		}
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor == 0 {
				m.cursor = int16(len(m.choices) - 1)
			} else {
				m.cursor--
			}
		case "down", "j":
			if m.cursor == int16(len(m.choices)-1) {
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
			if c, ok := m.cmds[int(m.selected)]; ok {
				cmd = exec.Command(c.name, c.args...)
			}
			return m, runShCmd(cmd)
		}
	}
	return m, nil
}

func (m model) View() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "untitled-dm v%s\n\nWhat should we do?\nPress space to select. Press enter to confirm selection.\n\n", version)
	for i, choice := range m.choices {
		ii := int16(i)
		cursor := " "
		if m.cursor == ii {
			cursor = ">"
		}
		checked := " "
		if m.selected == ii {
			checked = "x"
		}
		fmt.Fprintf(b, "%s [%s] %s\n", cursor, checked, choice)
	}
	tail := m.output.std
	if m.output.err != nil {
		tail = fmt.Sprintf("\nCould not start selected option: %v\n", m.output.err)
	}
	fmt.Fprintf(b, "\nPress q to quit.\n\n%s", tail)
	return b.String()
}

func runShCmd(cmd *exec.Cmd) tea.Cmd {
	return func() tea.Msg {
		if cmd == nil {
			return ranMsg{"No command to run\n"}
		}
		o, err := cmd.CombinedOutput()
		if err != nil {
			return errorMsg{fmt.Sprintf("%v\n\n%s", err, o)}
		}
		return ranMsg{string(o)}
	}
}

func parseTomlConfig(f string) (conf config) {
	if _, err := os.Stat(f); err != nil {
		return conf
	}
	if _, err := toml.DecodeFile(f, &conf); err != nil {
		log.Fatal("Could not decode config file: ", err)
	}
	return conf
}

type int16Value int16

func newInt16Value(val int16, p *int16) *int16Value {
	*p = val
	return (*int16Value)(p)
}

func (i *int16Value) Set(s string) error {
	x, err := strconv.ParseInt(s, 0, 16)
	*i = int16Value(x)
	return err
}

func (i *int16Value) String() string {
	return strconv.FormatInt(int64(*i), 10)
}

func getFlags() (flags struct {
	version          bool
	quitOnError      bool
	confFile         string
	defaultSelection int16
}) {
	flag.BoolVar(&flags.version, "V", false, "Print program version and exit")
	flag.BoolVar(&flags.quitOnError, "q", false, "Quit on command error")
	flag.StringVar(&flags.confFile, "c", "config.toml", "Path to configuration file")
	flag.Var(newInt16Value(-1, &flags.defaultSelection), "d", "Index to default selection")
	flag.Parse()
	return flags
}

func main() {
	flags := getFlags()
	if flags.version {
		fmt.Printf("%s v%s\n", os.Args[0], version)
		os.Exit(0)
	}
	conf := parseTomlConfig(flags.confFile)
	m := model{
		choices:     make([]string, len(conf.Commands)),
		cmds:        make(map[int]command, len(conf.Commands)),
		quitOnError: flags.quitOnError,
		selected:    flags.defaultSelection,
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
