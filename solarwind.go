package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/cli"
)

var (
	CurrentPath           string
	DefaultContentDir     string
	DefaultPostsDir       string
	DefaultDestinationDir string
	DefaultTemplateDir    string
)

const Solarwindfile = "Solarwindfile"

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	CurrentPath = dir
	DefaultContentDir := path.Join(dir, "content")
	DefaultPostsDir := path.Join(DefaultContentDir, "posts")
	DefaultDestinationDir := path.Join(dir, "public")
	DefaultTemplateDir := path.Join(dir, "templates")

	// Sanity check. Make sure a couple of these things exist
	for _, dir := range []string{DefaultContentDir, DefaultPostsDir, DefaultTemplateDir} {
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("%s does not exist. It must exist to continue generating a site.", dir)
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
	}

	exitCode, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitCode)
}
