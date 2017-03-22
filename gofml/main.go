package gofml

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var usageTemplate = `
GoFML helps organize go projects.

Usage:

    gofml [command] [arguments]

Commands:
{{ range . }}
    {{ .Name | printf "%-8s" }} {{ .Short }}{{end}}

Use "gofml help [command]" for command-specific information.
`

// Command is a command-line action.
type Command struct {
	Name    string
	Usage   string
	Short   string
	Long    string
	GetTask func([]string) (Task, error)
}

// Task is an action.
type Task interface {
	Run() error
}

// a map of command names -> commands.
var commands map[string]*Command
var usageText string

func init() {

	commands = make(map[string]*Command)
	commandList := []*Command{
		&initCommand,
		&helpCommand,
	}

	for _, cmd := range commandList {
		commands[cmd.Name] = cmd
	}
}

func usage() {

	tmpl := template.New("usage")
	tmpl, err := tmpl.Parse(strings.TrimSpace(usageTemplate) + "\n\n")

	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stderr, commands)

	if err != nil {
		panic(err)
	}
}

func Main() {

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 || len(args) == 1 && args[0] == "help" {
		flag.Usage()
		os.Exit(1)
	}

	cmd, found := commands[args[0]]

	if !found {
		fmt.Fprintf(os.Stderr, "gofml: unrecognized command %s\n", args[0])
		flag.Usage()
		os.Exit(1)
	}

	task, err := cmd.GetTask(args[1:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "gofml: failed to parse command %s\n", cmd)
		os.Exit(1)
	}

	err = task.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "gofml: error running command: %s\n", err)
		os.Exit(1)
	}
}
