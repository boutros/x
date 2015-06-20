package malle

import (
	"math/rand"
	"os"
	"testing"

	"github.com/boutros/x/malle/rdf"
)

const seed = 0x123

var (
	testDB *Store
	rnd    = rand.New(rand.NewSource(seed))
)

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

// func genRandTerm() rdf.Term { }
// func genRandLiteral() rdf.Literal {}
// func genRandResource() rdf.IRI {}
// func genRandProperty() rdf.IRI { }

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
	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.")
	h := hosts[rnd.Intn(len(hosts))]
	l := rnd.Intn(20)
	b := make([]rune, l)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return rdf.NewIRI(h + string(b))
}

func TestSaveIRI(t *testing.T) {
	iri := genRandIRI()
	id, err := testDB.AddTerm(iri)
	if err != nil {
		t.Fatalf("Store.AddTerm(%v)) == %v, want no error", iri, err)
	}

	stored, err := testDB.HasTerm(iri)
	if err != nil {
		t.Fatalf("Store.HasTerm(%v) == %v, want no error", iri, err)
	}

	if !stored {
		t.Fatalf("Store.AddTerm(%v) didn't store term in database", iri)
	}

	want, err := testDB.GetTerm(id)
	if err != nil {
		t.Fatalf("Store.GetTerm(%v)) == %v, want no error", id, err)
	}

	if !rdf.TermsEq(iri, want) {
		t.Fatal("Store.AddTerm returned wrong id")
	}

	id2, err := testDB.AddTerm(iri)
	if err != nil {
		t.Fatalf("Store.AddTerm(%v)) == %v, want no error", iri, err)
	}

	if id2 != id {
		t.Fatalf("Store.AddTerm(%v)) stored exisiting term as a new term", iri)
	}

}
