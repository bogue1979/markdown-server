package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/docopt/docopt-go"
	"github.com/russross/blackfriday"
)

const version = `markdown-server 1.0`
const usage = `Usage: markdown-server [-v] [--root=DIR] [ADDR]

Options:
  -h --help        Show this screen.
     --version     Show version.
  -v --verbose     Show more information.
     --root=DIR    Document root. [Default: .]
     --assets=DIR  Serve assets from a directory.
`

var (
	verbose   bool
	httpAddr  string
	rootDir   string
	assetDir  string
	templates *template.Template
)

func init() {
	opts, _ := docopt.Parse(usage, nil, true, version, false)

	log.SetFlags(0)

	var err error

	verbose = opts["--verbose"].(bool)

	if opts["ADDR"] == nil {
		opts["ADDR"] = "127.0.0.1:8080"
	}
	httpAddr = opts["ADDR"].(string)

	rootDir, err = filepath.Abs(opts["--root"].(string))
	if err != nil {
		log.Fatalf("Fatal: invalid document root %q: %v", opts["--root"], err)
	}

	if opts["--assets"] == nil {
		opts["--assets"] = "./assets"
	}
	assetDir, err = filepath.Abs(opts["--assets"].(string))
	if err != nil {
		log.Fatalf("Fatal: invalid asset directory %q: %v", opts["--assets"], err)
	}

	templates, err = template.ParseGlob(filepath.Join(assetDir, "*.html"))
	if err != nil {
		log.Fatalf("Fatal: template error: %v", err)
	}
}

func main() {
	log.Printf("httpAddr=%v rootDir=%v assetDir=%v", httpAddr, rootDir, assetDir)

	http.Handle("/assets/", Log(http.StripPrefix("/assets/", http.FileServer(http.Dir(assetDir)))))
	http.Handle("/markdown/", Log(http.StripPrefix("/markdown/", http.HandlerFunc(markdown))))
	http.Handle("/favicon.ico", Log(http.HandlerFunc(favicon)))
	http.Handle("/", Log(http.HandlerFunc(index)))

	log.Printf("starting server at %v", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

func favicon(w http.ResponseWriter, r *http.Request) {
	// TODO: serve favicon.ico
}

func index(w http.ResponseWriter, r *http.Request) {
	matches, err := filepath.Glob(filepath.Join(rootDir, "*.md"))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	files := []string{}
	for _, m := range matches {
		dir, _ := filepath.Rel(rootDir, m)
		files = append(files, dir)
	}

	if err := templates.ExecuteTemplate(w, "index.html", files); err != nil {
		log.Printf("Error: %v", err)
		return
	}
}

type Markdown struct {
	Filename string
	Content  string
}

func markdown(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile(filepath.Join(rootDir, r.URL.Path))
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	flags := blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	extensions := blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_FOOTNOTES |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_TITLEBLOCK |
		blackfriday.EXTENSION_AUTO_HEADER_IDS
	renderer := blackfriday.HtmlRenderer(flags, "", "")

	markdown := &Markdown{
		Filename: r.URL.Path,
		Content:  string(blackfriday.Markdown(b, renderer, extensions)),
	}

	if err := templates.ExecuteTemplate(w, "markdown.html", markdown); err != nil {
		log.Printf("Error: %v", err)
		return
	}
}
