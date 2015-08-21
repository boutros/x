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
		<tr><td><b>Number of different IRI namespaces</b></td><td>{{.NumNamespaces}}</td></tr>
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
	<title>{{or (.Props | chooseTitle) .Subj}}</title>
	<style type="text/css">
		body { font-family: sans serif; margin: 40px auto; max-width: 1140px; line-height: 1.6; font-size: 18px; color: #222; padding: 0 10px }
		h1, h2, h3 { line-height: 1.2; }
		h3 { border-top: 4px solid #222; padding-top: 0.5em; }
		td { padding-right: 2em; vertical-align: top; }
		a { text-decoration: none }
		.right { text-align: right; }
		.grey { color: #aaa; }
		.container { margin-bottom: 2em; padding: 0 2em; }
		.clearfix { clear: both; }
		.border { border-top: 1px solid #ccc; }
		.props { margin-bottom: 0.5em;}
		.predicate { width: 22%; float: left; box-sizing: border-box; padding-left: 0.5em; }
		.values { width: 78%; float: right; position: relative; }
		ul { list-style: none; padding: 0; margin: 0; }
		li { padding: 0.15em; }
		.literal { display: inline-block; margin-right: 1em; min-width: 200px; }
		.resource { position: relative; }
	</style>
</head>
<body>
	<div class="container">
		<h2>{{.Props | chooseTitle}}</h2>
		<h3>{{.Subj | html}}</h3>
		<div>
			{{range $pred, $terms := .Props}}
				<div class="props border clearfix">
					<div class="predicate" title="{{$pred | html}}"><b>{{$pred | shortPred}}</b>{{if gt (len $terms) 1 }} <span class="grey">({{len $terms}})</span>{{end}}</div>
					<ul class="values">
					{{ range $obj := $terms}}
						<li class={{if not (isLink $obj)}}"literal"{{else}}"resource"{{end}}>{{$obj | linkify}}</li>
					{{end}}
					</li>
				</div>
			{{end}}
		</div>
		<div class="clearfix"></div>
		<h3 class="right">{{.Subj | html}}</h3>
	</div>
	<div class="clearfix"></div>
</body>
</html>`

var titlePreferences = map[rdf.Term]rdf.IRI{
	mustNewIRI("http://lexvo.org/ontology#Language"):                mustNewIRI("http://www.w3.org/2008/05/skos#prefLabel"),
	mustNewIRI("http://www.w3.org/2004/02/skos/core#Concept"):       mustNewIRI("http://www.w3.org/2004/02/skos/core#prefLabel"),
	mustNewIRI("http://www.w3.org/2004/02/skos/core#ConceptScheme"): mustNewIRI("http://purl.org/dc/terms/title"),
	mustNewIRI("http://xmlns.com/foaf/0.1/Person"):                  mustNewIRI("http://xmlns.com/foaf/0.1/name"),
}

var rdfType = mustNewIRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")

func mustNewIRI(iri string) rdf.IRI {
	i, err := rdf.NewIRI(iri)
	if err != nil {
		panic(err)
	}
	return i
}

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
		"isLink": func(term rdf.Term) bool {
			_, ok := term.(rdf.IRI)
			return ok
		},
		"linkify": func(term rdf.Term) template.HTML {
			switch t := term.(type) {
			case rdf.IRI:
				link := fmt.Sprintf("</a><a href=\"/describe?IRI=%v\">%v</a>",
					template.HTMLEscapeString(t.Value().(string)), template.HTMLEscapeString(t.String()))
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
		"chooseTitle": func(triples map[rdf.IRI]rdf.Terms) string {
			// TODO also have prefered language tag?
			if types, ok := triples[rdfType]; ok {
				if titlePred, ok := titlePreferences[types[0]]; ok {
					if titles, ok := triples[titlePred]; ok {
						return titles[0].Value().(string)
					}
				}
			}
			return ""
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
		graph, err := db.Query(malle.NewQuery().Resource(iri))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if graph == nil || graph.IsEmpty() {
			http.Error(w, "No triples found", http.StatusNotFound)
			return
		}
		tplResource.Execute(w, struct {
			Subj  rdf.IRI
			Props map[rdf.IRI]rdf.Terms
		}{iri, graph[iri]})
	})
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
