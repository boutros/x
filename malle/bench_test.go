package malle

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/boutros/x/malle/rdf"
)

func genRandTerm() rdf.Term {
	i := rnd.Intn(10)
	if i < 5 {
		return genRandIRI()
	}
	return genRandLiteral()
}

func genRandLiteral() rdf.Literal {
	i := rnd.Intn(10)
	var err error
	var l rdf.Literal
	switch {
	case i < 5:
		l, err = rdf.NewLiteral(genRandString(2000))
	// TODO xsd datatypes
	default:
		l, err = rdf.NewLangLiteral(genRandString(2000), genRandSCIIString(8))
	}
	if err != nil {
		panic(err)
	}
	return l
}

func genRandIRI() rdf.IRI {
	hosts := []string{
		"http://example.org/title/",
		"http://example.org/person#",
		"http://ok.com/resource/",
		"http://xyz.no/data/",
		"http://db.com/id#",
		"http://example.com/place/",
		"http://example.com/animal#",
	}
	h := hosts[rnd.Intn(len(hosts))]
	iri, err := rdf.NewIRI(h + genRandSCIIString(10))
	if err != nil {
		panic(err)
	}
	return iri
}

func genRandPred() rdf.IRI {
	preds := []string{
		"http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		"http://purl.org/dc/terms/created",
		"http://purl.org/dc/terms/modified",
		"http://www.w3.org/2000/01/rdf-schema#label",
		"http://www.w3.org/1999/02/22-rdf-syntax-ns#first",
		"http://www.w3.org/1999/02/22-rdf-syntax-ns#rest",
		"http://www.w3.org/2004/02/skos/core#prefLabel",
		"http://dbpedia.org/ontology/literaryGenre",
		"http://purl.org/dc/terms/contributor",
		"http://commontag.org/ns#label",
		"http://commontag.org/ns#tagged",
		"http://purl.org/ontology/bibo/edition",
		"http://data.deichman.no/dugnadsbaseID",
		"http://xmlns.com/foaf/0.1/depiction",
		"http://data.deichman.no/narLabel",
		"http://www.w3.org/2002/07/owl#sameAs",
		"http://purl.org/dc/terms/language",
		"http://data.deichman.no/bibliofilID",
		"http://purl.org/dc/terms/format",
		"http://data.deichman.no/location_signature",
	}
	iri := preds[rnd.Intn(len(preds))]
	return mustNewIRI(iri)
}

func genRandString(length int) string {
	l := rnd.Intn(length) + 1
	r := make([]rune, l)
	for i := range r {
		r[i] = rune(rnd.Int31n(2000 + 60))
	}
	return string(r)
}

func genRandSCIIString(length int) string {
	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ._")
	l := rnd.Intn(length) + 1
	r := make([]rune, l)
	for i := range r {
		r[i] = letters[rand.Intn(len(letters))]
	}
	return string(r)
}

func genRandTriple() rdf.Triple {
	return rdf.NewTriple(genRandIRI(), genRandPred(), genRandTerm())
}

func genRandTriples(n int) []rdf.Triple {
	trs := make([]rdf.Triple, 0, n)
	for i := 0; i < n; i++ {
		tr := genRandTriple()
		trs = append(trs, tr)
		if r := rnd.Intn(10); r > 6 {
			c := rnd.Intn(3 + 1)
			pred := genRandPred()
			for j := 0; j < c; j++ {
				trs = append(trs, rdf.NewTriple(tr.Subject(), pred, genRandTerm()))
				i++
			}
		}
	}
	return trs
}

func BenchmarkAddTriple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := testDB.AddTriple(genRandTriple())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImport10Triples(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := testDB.ImportTriples(genRandTriples(10))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImport100Triples(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := testDB.ImportTriples(genRandTriples(100))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImport1000Triples(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := testDB.ImportTriples(genRandTriples(1000))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImport100NTriples(b *testing.B) {
	trs := genRandTriples(100)
	var bf bytes.Buffer
	for _, tr := range trs {
		bf.WriteString(tr.String())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.Import(bytes.NewReader(bf.Bytes()), 100, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImportGraph100(b *testing.B) {
	trs := genRandTriples(100)
	g := rdf.NewGraph()
	for _, tr := range trs {
		g.Add(tr)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := testDB.ImportGraph(g)
		if err != nil {
			b.Fatal(err)
		}
	}
}
