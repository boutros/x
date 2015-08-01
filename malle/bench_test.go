package malle

import (
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
	return rdf.NewTriple(genRandIRI(), genRandIRI(), genRandTerm())
}

func genRandTriples(n int) []rdf.Triple {
	trs := make([]rdf.Triple, 0, n)
	for i := 0; i < n; i++ {
		trs = append(trs, genRandTriple())
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
