package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/russross/blackfriday/v2"
)

var listen = flag.String("listen", "localhost:1404", "listen host:port") // M(14), D(04)

func main() {
	flag.Parse()

	httpdir := http.Dir(".")
	handler := renderer{httpdir, http.FileServer(httpdir)}

	fmt.Printf("Serving on http://%s\n", *listen)
	go Open("http://" + *listen)

	log.Fatal(http.ListenAndServe(*listen, handler))
}

var outputTemplate = template.Must(template.New("base").Parse(MDTemplate))

type renderer struct {
	d http.Dir
	h http.Handler
}

func isDir(req *http.Request) bool {
	return strings.HasSuffix(req.URL.Path, "/")
}

func hasSuffix(text string, list []string) bool {
	for _, s := range list {
		if strings.HasSuffix(text, s) {
			return true
		}
	}
	return false
}

var codeExtensions = []string{".a", ".asm", ".asp", ".awk", ".bat", ".c", ".class", ".cmd", ".cpp", ".csv", ".json", ".yaml", ".yml", ".cxx", ".h", ".html", ".ini", ".java", ".js", ".jsp", ".log", ".map", ".mod", ".sh", ".bash", ".txt", ".xml", ".py", ".go", ".rs", ".coffee", ".conf", ".config", "cpp", "cr", "css", "d", "dart", "exmaple", "fish", "gradle", "h", "jade", "json5", "jsx", "key", "less", "m4", "markdown", "md", "patch", "pem", "plist", "properties", "pub", "pug", "rb", "rc", "sass", "scpt", "scss", "sql", "template", "todo", "toml", "ts", "tsx", "vim", "vue", "xhtml", "xml"}

func (r renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, ".md") {
		// net/http is already running a path.Clean on the req.URL.Path,
		// so this is not a directory traversal, at least by my testing
		input, err := ioutil.ReadFile("." + req.URL.Path)
		if err != nil {
			http.Error(rw, "Internal Server Error", 500)
			log.Fatalf("Couldn't read path %s: %v", req.URL.Path, err)
		}
		output := blackfriday.Run(input)

		rw.Header().Set("Content-Type", "text/html")

		outputTemplate.Execute(rw, struct {
			Path string
			Body template.HTML
		}{
			Path: req.URL.Path,
			Body: template.HTML(string(output)),
		})
	} else if hasSuffix(req.URL.Path, codeExtensions) {
		content, err := ioutil.ReadFile("." + req.URL.Path)
		if err != nil {
			//http.Error(rw, "Internal Server Error", 500)
			//log.Fatalf("Couldn't read path %s: %v", req.URL.Path, err)
			http.Error(rw, "File not found", http.StatusNotFound)
			log.Printf("Couldn't read path %s: %v", req.URL.Path, err)
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.Write(content)
	} else {
		if isDir(req) {
			rw.Write([]byte(MDTemplateIndex))
		}
		r.h.ServeHTTP(rw, req)
		if isDir(req) {
			rw.Write([]byte(MDTemplateIndexTail))
		}
	}
}
