package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"unicode/utf8"

	"github.com/boutros/x/malle"
	"github.com/boutros/x/malle/rdf"
)

const htmlIndex = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Malle triple store frontend</title>
	<style type="text/css">
		body { font-family: sans serif; margin: 40px auto; max-width: 1140px; line-height: 1.6; font-size: 18px; color: #222; padding: 0 10px }
		h1, h2, h3 { line-height: 1.2;}
	</style>
</head>
<body>
	<h2>Malle Triple Store Frontend</h2>	
	<table>
		<tr><td><b>Database file</b></td><td>{{.File}}</td></tr>
		<tr><td><b>Database file size</b></td><td>{{.SizeInBytes}}</td></tr>
		<tr><td><b>Number of unique terms</b></td><td>{{.NumTerms}}</td></tr>
		<tr><td><b>Number of triples</b></td><td>{{.NumTriples}}</td></tr>
	</table>
	<form action="/describe">
		<p>Enter the IRI of a RDF resource to start browsing:</p>
		<input type="search" name="IRI"/> <button>Explore</button>
	</form>
</body>
</html>`

const htmlResource = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>{{. | chooseTitle}}</title>
	<style type="text/css">
		body { font-family: sans serif; margin: 40px auto; max-width: 1140px; line-height: 1.6; font-size: 18px; color: #222; padding: 0 10px }
		h1, h2, h3 { line-height: 1.2;}
		td { padding-right: 2em; vertical-align: top; }
		.grey { color: #aaa; }
	</style>
</head>
<body>
	<div>
		<h3>{{with index .  0}}{{.Subject | html}}{{end}}</h3>
		<table>
			{{range .}}
			<tr>
				<td title="{{.Predicate | html}}">{{.Predicate | shortPred}}</td>
				<td>{{.Object | linkify}}</td>
			</tr>
			{{end}}
		</table>
	</div>
</body>
</html>`

func shorten(s string) string {
	i := len(s)
	for r, w := utf8.DecodeLastRuneInString(s[:i]); i > 0; r, w = utf8.DecodeLastRuneInString(s[:i]) {
		i -= w
		if r == '#' || r == '/' {
			return s[i+1:]
		}
	}
	return s
}
func main() {
	funcMap := template.FuncMap{
		"shortPred": func(t rdf.Term) string {
			s := t.Value().(string)
			return shorten(s)
		},
		"linkify": func(term rdf.Term) template.HTML {
			switch t := term.(type) {
			case rdf.IRI:
				link := fmt.Sprintf("<a href=\"/describe?IRI=%v\">%v</a>", template.HTMLEscapeString(t.Value().(string)), template.HTMLEscapeString(t.String()))
				return template.HTML(link)
			case rdf.Literal:
				switch t.DataType().Value().(string) {
				case "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString":
					literal := fmt.Sprintf("%v <span class=\"grey\">@%v</span>",
						t.Value(),
						template.HTMLEscapeString(t.Lang()))
					return template.HTML(literal)
				case "http://www.w3.org/2001/XMLSchema#string":
					return template.HTML(template.HTMLEscapeString(t.Value().(string)))
				default:
					literal := fmt.Sprintf("%v <span class=\"grey\" title=\"%s\">(%v)</span>",
						t.Value(),
						template.HTMLEscapeString(t.DataType().Value().(string)),
						shorten(t.DataType().Value().(string)))
					return template.HTML(literal)
				}
			}
			panic("unreachable")
		},
		"chooseTitle": func(triples []rdf.Triple) string {
			return "TODO choose title from triples"
		},
	}
	var (
		// templates:
		tplIndex    = template.Must(template.New("index").Parse(htmlIndex))
		tplResource = template.Must(template.New("index").Funcs(funcMap).Parse(htmlResource))
		// command line flags:
		dbFile     = flag.String("db", "", "database file")
		port       = flag.Int("p", 8080, "port to serve from")
		importFile = flag.String("import", "", "import triples from file (n-triples)")
	)
	flag.Parse()
	if *dbFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *importFile == "" {
		_, err := os.Stat(*dbFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Initializing triple store from file: %s", *dbFile)
	db, err := malle.Init(*dbFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Triple store OK")
	defer db.Close()

	if *importFile != "" {
		log.Printf("Importing triples from file: %v", *importFile)
		go func() {
			f, err := os.Open(*importFile)
			if err != nil {
				log.Printf("Cannot open file: %v", err.Error())
				return
			}
			start := time.Now()
			n, err := db.Import(f, 1000, true)
			if err != nil {
				log.Printf("Import from %v failed: %v", *importFile, err.Error())
				return
			}
			now := time.Now()
			log.Printf("Done importing %d triples from file %v in %v", n, *importFile, now.Sub(start))
		}()
	}

	log.Printf("DB: %+v", db.Stats())
	log.Printf("Serving from port %d", *port)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		tplIndex.Execute(w, db.Stats())
	})
	http.HandleFunc("/describe", func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()["IRI"][0] // TODO check if IRI param present
		iri, err := rdf.NewIRI(q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := db.Query(malle.NewQuery().Resource(iri))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if res == nil || len(res) == 0 {
			http.Error(w, "No triples found", http.StatusNotFound)
			return
		}
		tplResource.Execute(w, res)
	})
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
