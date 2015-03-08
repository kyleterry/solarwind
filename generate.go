package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/extemporalgenome/slug"
	"github.com/mitchellh/cli"
	"github.com/russross/blackfriday"
)

const (
	TypeMarkdown = "md"
	TypeHTML     = "html"
)

type FileMapper struct {
	SourceFile      string
	DestinationFile string
	Filename        string
	Filetype        string
}

type Context struct {
	SiteTitle       string `json:"site_title"`
	SiteDescription string `json:"site_description"`
	Posts           []MarkdownPage
	CurrentPage     MarkdownPage
}

type Page interface {
	GetType() string
	GetFinalHTML() template.HTML
	GetTitle() string
}

type MarkdownPage struct {
	Title           string
	Slug            string
	Date            time.Time
	Category        string
	Filename        string
	DestinationFile string
	RelLink         string
	RawMarkdown     string        // This is the Markdown sans header
	FinalHTML       template.HTML // This is the final HTML after the Markdown parser
}

type HTMLPage struct {
	RawHTML   string
	Filename  string
	FinalHTML template.HTML
}

func (p MarkdownPage) GetType() string {
	return TypeMarkdown
}

func (p MarkdownPage) GetFinalHTML() template.HTML {
	return p.FinalHTML
}

func (p MarkdownPage) GetTitle() string {
	return p.Title
}

func (p HTMLPage) GetType() string {
	return TypeHTML
}

func (p HTMLPage) GetFinalHTML() template.HTML {
	return p.FinalHTML
}

func (p HTMLPage) GetTitle() string {
	return ""
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
	log.Printf("Parsing %s", filename)
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

			sl := strings.SplitN(line, ":", 2)
			sl[1] = strings.Trim(sl[1], " ")
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
	page.Slug = slug.Slug(page.Title)
	page.DestinationFile = path.Join(DefaultDestinationDir, "posts", page.Slug+".html")
	page.RelLink = "posts/" + page.Slug + ".html"
	return page
}

func NewHTMLPage(filename string, rawContent string) HTMLPage {
	return HTMLPage{RawHTML: rawContent, Filename: filename}
}

func NewContext() *Context {
	return &Context{}
}

func NewContextFromSolarwindfile(path string) *Context {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("There was a problem reading the Solarwindfile")
	}

	context := NewContext()
	if err := json.Unmarshal(content, &context); err != nil {
		log.Fatal("There was a problem parsing the Solarwindfile")
	}

	if context.SiteTitle == "" {
		context.SiteTitle = "Solarwind Site"
	}

	if context.SiteDescription == "" {
		context.SiteDescription = "This is a static site generated with Solarwind: https://github.com/kyleterry/solarwind"
	}

	return context
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
		fm.Filename = strings.Split(filepath.Base(f), ".")[0]
		switch dir {
		case DefaultContentDir:
			fm.SourceFile = f
			fm.DestinationFile = path.Join(DefaultDestinationDir, fm.Filename+".html")
		case path.Join(DefaultContentDir, "posts"):
			fm.SourceFile = f
			fm.DestinationFile = path.Join(DefaultDestinationDir, "posts", fm.Filename+".html")
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
		log.Fatal(err)
	}
}

func GenerateHTMLFromMarkdown(rawMarkdown string) string {
	return string(blackfriday.MarkdownCommon([]byte(rawMarkdown)))
}

func MakeFinalPage(htmlContent string) string {
	return ""
}

// Cowboy error handling
func CopyAssets(source, dest string) {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		log.Fatal(err)
	}
	filepath.Walk(source, func(p string, info os.FileInfo, err error) error {
		if p == source {
			return nil
		}
		if info.IsDir() {
			os.Mkdir(path.Join(dest, info.Name()), info.Mode())
			return nil
		}
		new_path := strings.Replace(p, source, dest, 1)
		r, err := os.Open(p)
		if err != nil {
			log.Fatal(err)
		}

		defer r.Close()
		w, err := os.Create(new_path)

		if err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(w, r); err != nil {
			w.Close()
			log.Fatal(err)
		}
		return nil
	})
}

// GenerateCommand code
type GenerateCommand struct {
	Ui cli.Ui
}

func (c *GenerateCommand) Help() string {
	helpText := `
usage: solarwind generate
	This command will build a solarwind project and put everything in ./public.
	`
	return helpText
}

func (c *GenerateCommand) Synopsis() string {
	return "Builds a static site from markdown content."
}

func (c *GenerateCommand) Run(args []string) int {
	if _, err := os.Stat(Solarwindfile); err != nil {
		log.Fatal("You need to create a `Solarwindfile` in the directory you'd like to serve as your site.")
	}

	log.Println("Making public directory")
	MakePublicDir(DefaultDestinationDir)
	context := NewContextFromSolarwindfile(DefaultSolarwindfilePath)
	var posts []MarkdownPage

	log.Println("Caching templates")
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

	log.Println("Collecting content")
	rootMarkdownFiles := ListFiles(DefaultContentDir, TypeMarkdown)
	rootHTMLFiles := ListFiles(DefaultContentDir, TypeHTML)
	postMarkdownFiles := ListFiles(DefaultPostsDir, TypeMarkdown)
	fileCount := len(rootMarkdownFiles) + len(rootHTMLFiles) + len(postMarkdownFiles)
	log.Printf("Found %d files", fileCount)

	// Merge root files so one loop is needed
	rootFilesToRead := append(rootMarkdownFiles, rootHTMLFiles...)

	log.Println("Parsing posts")
	for _, file := range postMarkdownFiles {
		content, err := ioutil.ReadFile(file.SourceFile)
		if err != nil {
			log.Fatalf("There was an error reading the file: %s", err)
		}

		post := NewMarkdownPage(file.Filename, string(content))
		post.FinalHTML = template.HTML(GenerateHTMLFromMarkdown(post.RawMarkdown))
		posts = append(posts, post)
	}

	context.Posts = posts

	log.Println("Parsing pages and generating site")
	for _, file := range rootFilesToRead {
		content, err := ioutil.ReadFile(file.SourceFile)
		if err != nil {
			log.Fatalf("There was an error reading the file: %s", err)
		}
		var page Page
		if file.Filetype == TypeMarkdown {
			md := NewMarkdownPage(file.Filename, string(content))
			md.FinalHTML = template.HTML(GenerateHTMLFromMarkdown(md.RawMarkdown))
			page = md
		} else {
			html := NewHTMLPage(file.Filename, string(content))
			html.FinalHTML = template.HTML(string(content))
			page = html
		}

		t := template.Must(template.New("page").Parse(string(indexCache) + string(pageCache) + string(page.GetFinalHTML())))

		// TODO: make custom io.Writer to write the template directly to a file
		b := &bytes.Buffer{}
		t.Execute(b, context)

		err = ioutil.WriteFile(file.DestinationFile, b.Bytes(), 0755)
		if err != nil {
			panic(err)
		}
	}

	for _, post := range context.Posts {
		context.CurrentPage = post
		t := template.Must(template.New("page").Parse(string(indexCache) + string(postCache)))

		b := &bytes.Buffer{}
		t.Execute(b, context)

		err = ioutil.WriteFile(post.DestinationFile, b.Bytes(), 0755)
		if err != nil {
			panic(err)
		}
	}

	log.Println("Copying static assets")
	CopyAssets(DefaultStaticDir, path.Join(DefaultDestinationDir, "static"))

	log.Println("Done!")

	return 0
}
