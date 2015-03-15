package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/howeyc/fsnotify"
	"github.com/mitchellh/cli"
)

// Goroutine to watch for file changes and regenerate the site
// TODO: clean up the error handling in this function
func watch() {
	log.Println("Watching for changes...")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Watch(DefaultContentDir)
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Watch(path.Join(DefaultContentDir, "posts"))
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Watch(DefaultTemplateDir)
	if err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(DefaultStaticDir, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			err = watcher.Watch(p)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-watcher.Event:
			log.Println("Change detected. Regenerating site...")
			gc := GenerateCommand{nil}
			gc.Run([]string{})
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}

// ServerCommand code
type ServerCommand struct {
	Ui cli.Ui
}

func (c *ServerCommand) Help() string {
	helpText := `
Usage: solarwind server [options]
	This will watch for changes to files and regenereate the site when those
	changes are detected.

	Options:
		-bind ":8090" 
			Binds to a specific address
	`
	return helpText
}

func (c *ServerCommand) Synopsis() string {
	return "Run a development server."
}

func (c *ServerCommand) Run(args []string) int {
	var defaultBind string
	flag.StringVar(&defaultBind, "bind", "localhost:8090", "Set an address to bind to")

	log.Println("About to start development server")

	go watch()

	log.Printf("Server listening on http://%s", defaultBind)
	err := http.ListenAndServe(defaultBind, http.FileServer(http.Dir(DefaultDestinationDir)))
	if err != nil {
		log.Fatal(err)
	}
	return 0
}
