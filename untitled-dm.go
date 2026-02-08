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

const version = "1.0.0"

type (
	// model implements [tea.Model]
	model struct {
		// choices is the list of choices that the user can choose from.
		choices []string
		// cmds maps indices in choices to commands to run.
		cmds        map[int]command
		quitOnError bool
		// output contains the combined standard streams of the command and any error that occurred
		output struct {
			std string
			err error
		}
		// cursor is the index of the current cursor position in choices.
		cursor int
		// selected is the index of selected item in choices. No items are selected if negative
		selected int
	}
	command struct {
		// Name is the string that is presented in the menu
		Name string
		// Command is the name of the program to be executed
		Command string
		// Args are passed to Command
		Args []string
	}
	// config is the TOML structure of the configuration file
	config struct {
		Commands []command
	}
	// ranMsg indicates that the command successfully completed
	ranMsg struct{ output string }
	// errorMsg indicates that the command did not successfully complete. Implements [error]
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
				cmd = exec.Command(c.Command, c.Args...)
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
	tail := m.output.std
	if m.output.err != nil {
		tail = fmt.Sprintf("\nCould not start selected option: %v\n", m.output.err)
	}
	fmt.Fprintf(b, "\nPress q to quit.\n\n%s", tail)
	return b.String()
}

func (m model) AddCommands(commands []command, offset int) model {
	for i, c := range commands {
		m.choices[i+offset] = c.Name
		if c.Command != "" {
			m.cmds[i+offset] = c
		}
	}
	return m
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

func getShellArgs(argString string) (args []string) {
	for i := 0; i < len(argString)-1; {
		c := argString[i]
		if c == '"' {
			nextQuote := strings.Index(argString[i+1:], `"`) + i + 1
			args = append(args, argString[i+1:nextQuote])
			i = nextQuote + 1
		} else if c != ' ' {
			nextSpace := strings.Index(argString[i:], " ") + i
			if nextSpace < i { // strings.Index returned -1, i.e. this is last arg
				nextSpace = len(argString)
			}
			args = append(args, argString[i:nextSpace])
			i = nextSpace
		} else {
			i++
		}
	}
	return args
}

type commandsValue []command

func (i *commandsValue) String() string { return fmt.Sprint(*i) }

func (i *commandsValue) Set(s string) error {
	name, commandAndArgs, _ := strings.Cut(s, "=")
	cmd, argString, _ := strings.Cut(commandAndArgs, " ")
	if strings.Count(argString, `"`)%2 != 0 {
		return errorMsg{"Must match all quotations in command arguments"}
	}
	args := getShellArgs(argString)
	*i = append(*i, command{Name: name, Command: cmd, Args: args})
	return nil
}

func getFlags() (flags struct {
	version          bool
	quitOnError      bool
	confFile         string
	defaultSelection int
	extraCommands    commandsValue
}) {
	flag.BoolVar(&flags.version, "V", false, "Print program version and exit")
	flag.BoolVar(&flags.quitOnError, "q", false, "Quit when a command fails")
	flag.StringVar(&flags.confFile, "c", "config.toml", "Path to configuration file")
	flag.IntVar(&flags.defaultSelection, "d", -1, "Default `index` for selection. No selection if negative")
	flag.Var(&flags.extraCommands, "e", "Extra `command`s to include of the format:\nNAME=[COMMAND] [ARGS]...\nSurround arguments with spaces with double quotes.\nCan be provided multiple times")
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
		choices:     make([]string, len(conf.Commands)+len(flags.extraCommands)),
		cmds:        make(map[int]command, len(conf.Commands)+len(flags.extraCommands)),
		quitOnError: flags.quitOnError,
		selected:    flags.defaultSelection,
	}.AddCommands(flags.extraCommands, 0).AddCommands(conf.Commands, len(flags.extraCommands))
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
}
