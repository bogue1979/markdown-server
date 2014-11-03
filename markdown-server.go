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
  -h --help     Show this screen.
     --version  Show version.
  -v --verbose  Show more information.
     --root=DIR    Document root. [Default: .]`

var (
	httpAddr string
	rootDir  string
)

func init() {
	opts, _ := docopt.Parse(usage, nil, true, version, false)

	log.SetFlags(0)

	if opts["ADDR"] == nil {
		opts["ADDR"] = "127.0.0.1:8080"
	}

	var err error

	httpAddr = opts["ADDR"].(string)

	rootDir, err = filepath.Abs(opts["--root"].(string))
	if err != nil {
		fmt.Errorf("Fatal: invalid document root %q: %v", opts["--root"], err)
	}
}

func main() {
	log.Printf("rootDir=%v httpAddr=%v", rootDir, httpAddr)

	http.Handle("/markdown/", http.StripPrefix("/markdown/", http.HandlerFunc(markdown)))
	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

var indexTemplate = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Markdown Server</title>
</head>
<body>
	<ul>
	{{range .}}
		<li><a href="/markdown/{{.}}">{{.}}</a></li>
	{{end}}
	</ul>
</body>
</html>
`))

func index(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %v", r.RequestURI)

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

	if err := indexTemplate.Execute(w, files); err != nil {
		log.Printf("Error: %v", err)
		return
	}
}

var markdownTemplate = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Markdown Server</title>
</head>
<body>
	<ul>
	{{range .}}
		<li><a href="/markdown/{{.}}">{{.}}</a></li>
	{{end}}
	</ul>
</body>
</html>
`))

func markdown(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET %v -> %v", r.RequestURI, r.URL.Path)

	b, err := ioutil.ReadFile(filepath.Join(rootDir, r.URL.Path))
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	if _, err := w.Write(blackfriday.MarkdownCommon(b)); err != nil {
		log.Printf("Error: %v", err)
		return
	}
}
