package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/template"
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
	<title>title</title>
	<style type="text/css">
		body { font-family: sans serif; margin: 40px auto; max-width: 1140px; line-height: 1.6; font-size: 18px; color: #222; padding: 0 10px }
		h1, h2, h3 { line-height: 1.2;}
	</style>
</head>
<body>
	<div>
		<h3>{{with index .  0}}{{.Subject.NT | html}}{{end}}</h3>
		<table>
			{{range .}}
			<tr>
				<td>{{.Predicate.NT | html}}</td>
				<td>{{.Object.NT | html}}</td>
			</tr>
			{{end}}
		</table>
	</div>
</body>
</html>`

func main() {
	var (
		// templates:
		tplIndex    = template.Must(template.New("index").Parse(htmlIndex))
		tplResource = template.Must(template.New("index").Parse(htmlResource))
		// command line flags:
		dbFile = flag.String("db", "", "database file")
		port   = flag.Int("p", 8080, "port to serve from")
		//import = flag.String("i", "", "import triples from file (n-triples)")
	)
	flag.Parse()
	if *dbFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	_, err := os.Stat(*dbFile)
	if err != nil {
		log.Fatal(err)
	}

	db, err := malle.Init(*dbFile)
	log.Printf("Initializing triple store from file: %s", *dbFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Triple store OK")
	defer db.Close()

	log.Printf("DB: %+v", db.Stats())
	log.Printf("Serving from port %d", *port)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		tplIndex.Execute(w, db.Stats())
	})
	http.HandleFunc("/describe", func(w http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()["IRI"][0] // TODO check if IRI param present
		iri, err := rdf.NewIRI(p)
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