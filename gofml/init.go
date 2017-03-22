package gofml

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const script = `
export GOFML={{.ImportPath}}
export GOPATH={{.GoPath}}
export PATH="$GOPATH/bin:$PATH"
`

var initCommand = Command{
	Name:  "init",
	Short: "initialize a gofml",
	Usage: "init [-g][-p][-n] [import path]",
	Long: `
Init initializes a gofml and creates an initialization script that
activates it.  This script creates, if needed, a GOPATH directory
structure, symlinks the project into that structure at the specified
input path, and then alters the current session's GOPATH environment
variable to point to it.

The gofml can be deactivated with 'deactivate'.

Init supports the following options:

    -n
         the name of the environment, defaulting to the name
         of the current working directory.

    -g
         the GOPATH to create, defaulting to ~/.gofml/<name>

    -p
         the project path, defaulting to the current working
         directory.
`,
	GetTask: NewInitTask,
}

// InitTask initializes a gofml.
type InitTask struct {
	GoPath      string // the GOPATH to create, default "~/.gofml/<project name>"
	ImportPath  string // the import path of the project, e.g. "github.com/crsmithdev/gofml"
	ProjectName string // the name of the project, e.g. "gofml".
	ProjectPath string // the path to the project, default "./"
}

// NewInitTask returns a new InitTask created with the specified command-line args.
func NewInitTask(args []string) (Task, error) {

	flags := flag.NewFlagSet("init", flag.ExitOnError)

	goPath := flags.String("g", "", "the GOPATH to create")
	projectName := flags.String("n", "", "the project name")
	projectPath := flags.String("p", "", "the project path")

	flags.Parse(args)
	args = flags.Args()

	if len(args) < 1 {
		return nil, errors.New("no import path specified")
	}

	task := InitTask{
		ImportPath:  args[0],
		GoPath:      *goPath,
		ProjectName: *projectName,
		ProjectPath: *projectPath,
	}

	if task.ProjectName == "" {
		task.ProjectName = filepath.Base(task.ImportPath)
	}

	if task.GoPath == "" {
		task.GoPath = filepath.Join("/work/gofml", task.ProjectName)
	}

	if task.ProjectPath == "" {
		task.ProjectPath = filepath.Join(task.GoPath, "src", task.ImportPath)
	}

	return &task, nil
}

// Run exeuctes the InitTask
func (task *InitTask) Run() error {

	fmt.Println("gofml: initializing...")

	fmt.Printf("ImportPath=%s \n", task.ImportPath)
	fmt.Printf("ProjectName=%s \n", task.ProjectName)
	fmt.Printf("ProjectPath=%s \n", task.ProjectPath)
	fmt.Printf("GoPath=%s \n", task.GoPath)

	if err := task.makeDir(); err != nil {
		return err
	}

	if err := task.writeEnvrc(); err != nil {
		return err
	}

	if err := task.printHints(); err != nil {
		return err
	}

	fmt.Println("gofml: done")

	return nil
}

// writeScript writes the gofml activate script.
func (task *InitTask) makeDir() error {
	err := os.MkdirAll(task.ProjectPath, os.ModeDir|0775)

	fmt.Printf("gofml: create project directory %s\n", task.ProjectPath)

	return err
}

// writeScript writes the gofml activate script.
func (task *InitTask) writeEnvrc() error {

	scriptTemplate := template.New("test")
	scriptTemplate, err := scriptTemplate.Parse(script)

	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = scriptTemplate.Execute(&buf, task)

	if err != nil {
		return err
	}

	_ = os.Remove(filepath.Join(task.GoPath, ".envrc"))
	err = ioutil.WriteFile(filepath.Join(task.GoPath, ".envrc"), []byte(buf.String()), 0664)

	fmt.Printf("gofml: wrote activation script at %s\n", filepath.Join(task.GoPath, ".envrc"))

	return err
}

// writeScript writes the gofml activate script.
func (task *InitTask) printHints() error {

	fmt.Printf("gofml: now do `cd %s` to goto your project and then `direnv allow` to activate the direnv file. \n", task.ProjectPath)

	return nil
}
