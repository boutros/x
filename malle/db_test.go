package malle

import (
	"math/rand"
	"os"
	"testing"

	"github.com/boutros/x/malle/rdf"
)

const (
	seed    = 0x123
	numIter = 10 // number of random iterations for storing/retrieving tests
)

var (
	testDB *Store
	rnd    = rand.New(rand.NewSource(seed))
)

func mustNewIRI(iri string) rdf.IRI {
	i, err := rdf.NewIRI(iri)
	if err != nil {
		panic(err)
	}
	return i
}

func mustNewLiteral(val interface{}) rdf.Literal {
	l, err := rdf.NewLiteral(val)
	if err != nil {
		panic(err)
	}
	return l
}

func mustNewLangLiteral(val, lang string) rdf.Literal {
	l, err := rdf.NewLangLiteral(val, lang)
	if err != nil {
		panic(err)
	}
	return l
}

func mustNewTypedLiteral(val string, tp rdf.IRI) rdf.Literal {
	l, err := rdf.NewTypedLiteral(val, tp)
	if err != nil {
		panic(err)
	}
	return l
}

func TestMain(m *testing.M) {
	// setup
	var err error
	testDB, err = Init("_temp.db")
	if err != nil {
		panic(err)
	}

	retCode := m.Run()

	// teardown
	testDB.Close()
	err = os.Remove("_temp.db")
	if err != nil {
		panic(err)
	}
	os.Exit(retCode)
}

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
	iri, err := rdf.NewIRI(h + genRandSCIIString(20))
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

func TestEncodeDecode(t *testing.T) {
	nilDB := Store{}
	tests := []rdf.Term{
		mustNewIRI("a"),
		mustNewIRI("http://example.org/1/xyz.æøå"),
		mustNewLangLiteral("a", "en"),
		mustNewLangLiteral("æøå", "nb-no"),
	}
	for _, term := range tests {
		if !rdf.TermsEq(nilDB.decode(nilDB.encode(term)), term) {
			t.Errorf("Store.encode/decode roundtrip failed for %+v", term)
		}
	}
}

func TestStoreTerm(t *testing.T) {
	for i := 0; i < numIter; i++ {
		term := genRandTerm()
		id, err := testDB.AddTerm(term)
		if err != nil {
			t.Fatalf("Store.AddTerm(%v)) == %v, want no error", term, err)
		}

		stored, err := testDB.HasTerm(term)
		if err != nil {
			t.Fatalf("Store.HasTerm(%v) == %v, want no error", term, err)
		}

		if !stored {
			t.Fatalf("Store.AddTerm(%v) didn't store term in database", term)
		}

		want, err := testDB.GetTerm(id)
		if err != nil {
			t.Fatalf("Store.GetTerm(%v)) == %v, want no error", id, err)
		}

		if !rdf.TermsEq(term, want) {
			t.Fatal("Store.AddTerm returned wrong id")
		}

		id2, err := testDB.AddTerm(term)
		if err != nil {
			t.Fatalf("Store.AddTerm(%v)) == %v, want no error", term, err)
		}

		if id2 != id {
			t.Errorf("Store.AddTerm(%v)) stored exisiting term as a new term", term)
		}
	}
}
