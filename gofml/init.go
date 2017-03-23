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
export GOPATH={{.GoFmlPath}}
export PATH="$GOPATH/bin:$PATH"
`

func getGofmlRoot() string {
	root := os.Getenv("GOFMLROOT")

	if root == "" {
		root = "~/gofml"
	}

	return root
}

var GOFMLROOT string = getGofmlRoot()

var initCommand = Command{
	Name:  "init",
	Short: "initialize a gofml env",
	Usage: "init [-g][-n] [import path]",
	Long: fmt.Sprintf(`
Init supports the following options:

    -g
         the gofml root, where all projects are created defaulting to %s (uses env $GOFMLROOT if possible)

    -n
         the name of the environment, defaulting to the basename of the import path.

`, GOFMLROOT),
	GetTask: NewInitTask,
}

// InitTask initializes a gofml.
type InitTask struct {
	GoFmlRoot   string // the gofml root to create envs in, default "~/.gofml" or uses env $GOFMLROOT
	ImportPath  string // the import path of the project, e.g. "github.com/rubensayshi/gofml"
	ProjectName string // the name of the project, e.g. "gofml".

	GoFmlPath   string // GoFmlRoot + ProjectName
	ProjectPath string // the path to the project
}

// NewInitTask returns a new InitTask created with the specified command-line args.
func NewInitTask(args []string) (Task, error) {

	flags := flag.NewFlagSet("init", flag.ExitOnError)

	gofmlRoot := flags.String("g", getGofmlRoot(), "the gofml root")
	projectName := flags.String("n", "", "the project name")

	flags.Parse(args)
	args = flags.Args()

	if len(args) < 1 {
		return nil, errors.New("no import path specified")
	}

	task := InitTask{
		ImportPath:  args[0],
		GoFmlRoot:   *gofmlRoot,
		ProjectName: *projectName,
		ProjectPath: "",
		GoFmlPath:   "",
	}

	if task.ProjectName == "" {
		task.ProjectName = filepath.Base(task.ImportPath)
	}

	if task.GoFmlPath == "" {
		task.GoFmlPath = filepath.Join(task.GoFmlRoot, task.ProjectName)
	}

	if task.ProjectPath == "" {
		task.ProjectPath = filepath.Join(task.GoFmlPath, "src", task.ImportPath)
	}

	return &task, nil
}

// Run exeuctes the InitTask
func (task *InitTask) Run() error {

	fmt.Println("gofml: initializing...")

	fmt.Printf("GoFmlRoot=%s \n", task.GoFmlRoot)
	fmt.Printf("ImportPath=%s \n", task.ImportPath)
	fmt.Printf("ProjectName=%s \n", task.ProjectName)
	fmt.Printf("GoFmlPath=%s \n", task.GoFmlPath)
	fmt.Printf("ProjectPath=%s \n", task.ProjectPath)

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

	_ = os.Remove(filepath.Join(task.GoFmlPath, ".envrc"))
	err = ioutil.WriteFile(filepath.Join(task.GoFmlPath, ".envrc"), []byte(buf.String()), 0664)

	fmt.Printf("gofml: wrote activation script at %s\n", filepath.Join(task.GoFmlPath, ".envrc"))

	return err
}

// writeScript writes the gofml activate script.
func (task *InitTask) printHints() error {

	fmt.Printf("gofml: now do `cd %s` to goto your project and then `direnv allow` to activate the direnv file. \n", task.ProjectPath)

	return nil
}
