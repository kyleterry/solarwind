package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/cli"
)

var (
	CurrentPath              string
	DefaultContentDir        string
	DefaultPostsDir          string
	DefaultDestinationDir    string
	DefaultTemplateDir       string
	DefaultSolarwindfilePath string
	DefaultStaticDir         string
)

const Solarwindfile = "Solarwindfile"

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	CurrentPath = dir
	DefaultContentDir = path.Join(dir, "content")
	DefaultPostsDir = path.Join(DefaultContentDir, "posts")
	DefaultDestinationDir = path.Join(CurrentPath, "public")
	DefaultTemplateDir = path.Join(dir, "templates")
	DefaultSolarwindfilePath = path.Join(CurrentPath, Solarwindfile)
	DefaultStaticDir = path.Join(CurrentPath, "static")

	// Sanity check. Make sure a couple of these things exist
	for _, node := range []string{DefaultSolarwindfilePath, DefaultContentDir, DefaultPostsDir, DefaultTemplateDir} {
		if _, err := os.Stat(node); err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("%s does not exist. It must exist to continue generating a site.", node)
			} else {
				panic(err)
			}
		}
	}
}

func main() {
	c := cli.NewCLI("nbsssg", "0.1.0")
	ui := &cli.BasicUi{Writer: os.Stdout}
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"generate": func() (cli.Command, error) {
			return &GenerateCommand{
				Ui: ui,
			}, nil
		},
		"server": func() (cli.Command, error) {
			return &ServerCommand{
				Ui: ui,
			}, nil
		},
	}

	exitCode, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitCode)
}
