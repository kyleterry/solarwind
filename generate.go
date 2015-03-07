package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/russross/blackfriday"
)

var (
	pages     []MarkdownPage
	htmlPages []HTMLPage
	posts     []MarkdownPage
)

const (
	ExtensionMD   = "md"
	ExtensionHTML = "html"
)

type MarkdownPage struct {
	Title       string
	Date        time.Time
	Category    string
	Filename    string
	RawMarkdown string // This is the Markdown sans header
	FinalHTML   string // This is the final HTML after the Markdown parser
}

type HTMLPage struct {
	RawHTML   string
	FinalHTML string
}

type FileMapper struct {
	SourceFile      string
	DestinationPath string
	Filename        string
	Filetype        string
}

type MainContext struct {
	SiteTitle       string
	SiteDescription string
	Posts           []MarkdownPage
}

// NewMarkdownPage takes a string of raw markdown content with an optional header
// in the form of:
//
// ###
// title: this is a post title
// date: 2015-03-20 15:35 PDT
// category: computers
// ###
//
// This will parse out the header and return a new MarkdownPage instance with
// the header fields and raw Markdown content sans-header.
//
// :kyleterry: TODO: thing about returning an error here if something goes wrong
// during parsing of the header. Currently `log.Fatal`s.
func NewMarkdownPage(filename string, rawContent string) MarkdownPage {
	sd := strings.Split(rawContent, "\n")
	page := MarkdownPage{}
	page.Filename = filename
	if sd[0] == "###" {
		sd = sd[1:]
		for index, line := range sd {
			if line == "###" {
				sd = sd[index+1:]
				break
			}
			sl := strings.Split(line, ": ")
			if len(sl) > 2 {
				log.Fatal("Header should be in the format of \"key: value string\".")
			}

			switch sl[0] {
			case "title":
				page.Title = sl[1]
			case "date":
				parsedTime, err := time.Parse(time.RFC822, sl[1])
				if err != nil {
					log.Fatal("Malformed date string: can't parse date.")
				}
				page.Date = parsedTime
			case "category":
				page.Category = sl[1]
			default:
				// Just ignore things that we don't know about
				continue
			}
		}
	}
	if len(sd) == 0 {
		log.Fatalf("Something went wrong parsing %s: Possible Malformed header. Reached EOF.", filename)
	}
	page.RawMarkdown = strings.Join(sd, "\n")
	return page
}

func ListFiles(dir string, extension string) []FileMapper {
	files, err := filepath.Glob(fmt.Sprintf("%s/*.%s", dir, extension))
	if err != nil {
		log.Fatal("There was an error globbing for files")
	}
	fileMaps := []FileMapper{}
	for _, f := range files {
		fm := FileMapper{}
		fm.Filetype = extension
		fm.Filename = filepath.Base(f)
		switch dir {
		case DefaultContentDir:
			fm.SourceFile = f
			fm.DestinationPath = DefaultDestinationDir
		case path.Join(DefaultContentDir, "posts"):
			fm.SourceFile = f
			fm.DestinationPath = path.Join(DefaultDestinationDir, "posts")
		}
		fileMaps = append(fileMaps, fm)
	}
	return fileMaps
}

func MakePublicDir(dir string) {
	if _, err := os.Stat(dir); err == nil {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Fatalf("Could not remove dir %s", dir)
		}
	}
	err := os.MkdirAll(path.Join(dir, "posts"), 0755)
	if err != nil {
		log.Fatalf("Could not create dir: %s", err)
	}
}

func GenerateHTMLFromMarkdown(rawMarkdown string) string {
	return string(blackfriday.MarkdownCommon([]byte(rawMarkdown)))
}

func MakeFinalPage(htmlContent string) string {
	return ""
}

// GenerateCommand code
type GenerateCommand struct {
	Ui cli.Ui
}

func (c *GenerateCommand) Help() string {
	return "help"
}

func (c *GenerateCommand) Synopsis() string {
	return "help"
}

func (c *GenerateCommand) Run(args []string) int {
	if _, err := os.Stat(Solarwindfile); err == nil {
		log.Fatal("You need to create a `Solarwindfile` in the directory you'd like to serve as your site.")
	}

	MakePublicDir(DefaultDestinationDir)

	// Cache base template content
	indexCache, err := ioutil.ReadFile(path.Join(DefaultTemplateDir, "index.html"))
	if err != nil {
		log.Fatal("There was an error reading the index.html file")
	}

	pageCache, err := ioutil.ReadFile(path.Join(DefaultTemplateDir, "page.html"))
	if err != nil {
		log.Fatal("There was an error reading the page.html file")
	}

	postCache, err := ioutil.ReadFile(path.Join(DefaultTemplateDir, "post.html"))
	if err != nil {
		log.Fatal("There was an error reading the post.html file")
	}

	rootMarkdownFiles := ListFiles(DefaultContentDir, ExtensionMD)
	rootHTMLFiles := ListFiles(DefaultContentDir, ExtensionHTML)
	postMarkdownFiles := ListFiles(DefaultPostsDir, ExtensionMD)

	// Merge root files so one loop is needed
	rootFilesToRead := append(rootMarkdownFiles, rootHTMLFiles...)

	for _, file := range rootFilesToRead {
		content, err := ioutil.ReadFile(file.SourceFile)
		if err != nil {
			log.Fatalf("There was an error reading the file: %s", err)
		}
		var html string
		if file.Filetype == ExtensionMD {
			page := NewMarkdownPage("example.md", string(content))
			html = GenerateHTMLFromMarkdown(page.RawMarkdown)
		} else {
			html = string(content)
		}

		t := template.Must(template.New("page").Parse(string(indexCache) + string(pageCache) + html))

		// TODO: make custom io.Writer to write the template directly to a file
		b := &bytes.Buffer{}
		t.Execute(b, nil)

		err = ioutil.WriteFile(DefaultDestinationDir, b.Bytes(), 0755)
		if err != nil {
			panic(err)
		}
	}

	return 0
}
