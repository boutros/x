package malle

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/boutros/x/malle/rdf"
	"github.com/tgruben/roaring"
)

const (
	seed    = 0x123
	numIter = 3 // number of random iterations for storing/retrieving tests
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
			t.Fatalf("Store.AddTerm(%v)) == %v; want no error", term, err)
		}

		stored, err := testDB.HasTerm(term)
		if err != nil {
			t.Fatalf("Store.HasTerm(%v) == %v; want no error", term, err)
		}

		if !stored {
			t.Fatalf("Store.AddTerm(%v) didn't store term in database", term)
		}

		want, err := testDB.GetTerm(id)
		if err != nil {
			t.Fatalf("Store.GetTerm(%v)) == %v; want no error", id, err)
		}

		if !rdf.TermsEq(term, want) {
			t.Fatal("Store.AddTerm returned wrong id")
		}

		id2, err := testDB.AddTerm(term)
		if err != nil {
			t.Fatalf("Store.AddTerm(%v)) == %v; want no error", term, err)
		}

		if id2 != id {
			t.Errorf("Store.AddTerm(%v)) stored exisiting term as a new term", term)
		}
	}
}

func TestRemoveTerm(t *testing.T) {
	for i := 0; i < numIter; i++ {
		term := genRandTerm()
		_, err := testDB.AddTerm(term)
		if err != nil {
			t.Fatalf("Store.AddTerm(%v)) == %v; want no error", term, err)
		}

		err = testDB.RemoveTerm(term)
		if err != nil {
			t.Fatalf("Store.RemoveTerm(%v)) == %v; want no error", term, err)
		}

		err = testDB.RemoveTerm(term)
		if err != ErrNotFound {
			t.Fatalf("Store.RemoveTerm(%v)) == %v; want no ErrNotFound", term, err)
		}

		stored, err := testDB.HasTerm(term)
		if err != nil {
			t.Fatalf("Store.HasTerm(%v) == %v; want no error", term, err)
		}

		if stored {
			t.Fatalf("Store.RemoveTerm(%v) didn't remove term from database", term)
		}
	}
}

func TestAddTriple(t *testing.T) {
	tr := rdf.NewTriple(mustNewIRI("a"), mustNewIRI("p"), mustNewLiteral("o"))

	exists, err := testDB.HasTriple(tr)
	if err != nil || exists {
		t.Fatalf("Store.HasTriple(%v) == %v, %v; want false, nil", tr, exists, err)
	}

	err = testDB.AddTriple(tr)
	if err != nil {
		t.Fatalf("Store.AddTriple(%v) == %v; want no error", tr, err)
	}

	exists, err = testDB.HasTriple(tr)
	if err != nil || !exists {
		t.Fatalf("Store.HasTriple(%v) == %v, %v; want true, nil", tr, exists, err)
	}

	s, err := testDB.GetID(tr.Subject())
	if err != nil {
		t.Logf("Store.AddTriple(%v) didn't store all terms", tr)
		t.Fatalf("Store.GetID(%v) == %v; want no error", s, err)
	}

	p, err := testDB.GetID(tr.Predicate())
	if err != nil {
		t.Logf("Store.AddTriple(%v) didn't store all terms", tr)
		t.Fatalf("Store.GetID(%v) == %v; want no error", p, err)
	}

	o, err := testDB.GetID(tr.Object())
	if err != nil {
		t.Logf("Store.AddTriple(%v) didn't store all terms", tr)
		t.Fatalf("Store.GetID(%v) == %v; want no error", o, err)
	}

	indices := []struct {
		k1 uint32
		k2 uint32
		v  uint32
		bk []byte
	}{
		{s, p, o, bSPO},
		{o, s, p, bOSP},
		{p, o, s, bPOS},
	}

	key := make([]byte, 8)
	err = testDB.kv.View(func(tx *bolt.Tx) error {
		for _, i := range indices {
			bkt := tx.Bucket(i.bk)
			copy(key, u32tob(i.k1))
			copy(key[4:], u32tob(i.k2))
			bitmap := roaring.NewRoaringBitmap()

			bo := bkt.Get(key)
			if bo != nil {
				_, err := bitmap.ReadFrom(bytes.NewReader(bo))
				if err != nil {
					return err
				}
			}

			hasTriple := bitmap.Contains(i.v)
			if !hasTriple {
				return fmt.Errorf("Store.AddTriple(%v) didn't store triple in index %v", tr, string(i.bk))
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

}

func TestRemoveTriple(t *testing.T) {
	tr1 := rdf.NewTriple(mustNewIRI("s1"), mustNewIRI("p1"), mustNewLiteral("o1"))
	tr2 := rdf.NewTriple(mustNewIRI("s1"), mustNewIRI("p1"), mustNewLiteral("o2"))
	tr3 := rdf.NewTriple(mustNewIRI("s1"), mustNewIRI("p2"), mustNewLiteral("o1"))
	tr4 := rdf.NewTriple(mustNewIRI("s100"), mustNewIRI("p100"), mustNewLiteral("o100"))

	for _, tr := range []rdf.Triple{tr1, tr2, tr3} {
		err := testDB.AddTriple(tr)
		if err != nil {
			t.Fatalf("Store.AddTriple(%v) == %v; want no error", tr, err)
		}
	}

	startStats := testDB.Stats()

	err := testDB.RemoveTriple(tr3)
	if err != nil {
		t.Fatalf("Store.RemoveTriple(%v) == %v; want no error", tr3, err)
	}

	stats := testDB.Stats()
	if stats.NumTriples != startStats.NumTriples-1 {
		t.Fatalf("Store.Stats().NumTriples == %d; want %d", stats.NumTriples, startStats.NumTriples-1)
	}
	if stats.NumTerms != startStats.NumTerms-1 {
		t.Fatalf("Store.Stats().NumTerms == %d; want %d", stats.NumTerms, startStats.NumTerms-1)
	}
	startStats = stats

	err = testDB.RemoveTriple(tr1)
	if err != nil {
		t.Fatalf("Store.RemoveTriple(%v) == %v; want no error", tr1, err)
	}

	stats = testDB.Stats()
	if stats.NumTriples != startStats.NumTriples-1 {
		t.Fatalf("Store.Stats().NumTriples == %d; want %d", stats.NumTriples, startStats.NumTriples-1)
	}
	if stats.NumTerms != startStats.NumTerms-1 {
		t.Fatalf("Store.Stats().NumTerms == %d; want %d", stats.NumTerms, startStats.NumTerms-1)
	}
	startStats = stats

	err = testDB.RemoveTriple(tr2)
	if err != nil {
		t.Fatalf("Store.RemoveTriple(%v) == %v; want no error", tr2, err)
	}

	stats = testDB.Stats()
	if stats.NumTriples != startStats.NumTriples-1 {
		t.Fatalf("Store.Stats().NumTriples == %d; want %d", stats.NumTriples, startStats.NumTriples-1)
	}
	if stats.NumTerms != startStats.NumTerms-3 {
		t.Fatalf("Store.Stats().NumTerms == %d; want %d", stats.NumTerms, startStats.NumTerms-3)
	}

	err = testDB.RemoveTriple(tr4)
	if err != ErrNotFound {
		t.Fatalf("Store.RemoveTriple(%v) == %v; want ErrNotFound", tr2, err)
	}

}

func TestStats(t *testing.T) {
	tr1 := rdf.NewTriple(mustNewIRI("A"), mustNewIRI("P"), mustNewLiteral("O"))
	tr2 := rdf.NewTriple(mustNewIRI("A"), mustNewIRI("P"), mustNewLiteral("O2"))
	tr3 := rdf.NewTriple(mustNewIRI("A"), mustNewIRI("P2"), mustNewLiteral("O"))

	startStats := testDB.Stats()

	err := testDB.AddTriple(tr1)
	if err != nil {
		t.Fatalf("Store.AddTriple(%v) == %v; want no error", tr1, err)
	}

	stats := testDB.Stats()
	if stats.NumTerms != startStats.NumTerms+3 {
		t.Fatalf("Store.Stats().NumTerms == %d; want %d", stats.NumTerms, startStats.NumTerms+3)
	}
	if stats.NumTriples != startStats.NumTriples+1 {
		t.Fatalf("Store.Stats().NumTriples == %d; want %d", stats.NumTriples, startStats.NumTriples+1)
	}
	startStats = stats

	err = testDB.AddTriple(tr2)
	if err != nil {
		t.Fatalf("Store.AddTriple(%v) == %v; want no error", tr2, err)
	}
	stats = testDB.Stats()
	if stats.NumTerms != startStats.NumTerms+1 {
		t.Fatalf("Store.Stats().NumTerms == %d; want %d", stats.NumTerms, startStats.NumTerms+1)
	}
	if stats.NumTriples != startStats.NumTriples+1 {
		t.Fatalf("Store.Stats().NumTriples == %d; want %d", stats.NumTriples, startStats.NumTriples+1)
	}
	startStats = stats

	err = testDB.AddTriple(tr3)
	if err != nil {
		t.Fatalf("Store.AddTriple(%v) == %v; want no error", tr3, err)
	}
	err = testDB.AddTriple(tr3)
	if err != nil {
		t.Fatalf("Store.AddTriple(%v) == %v; want no error", tr3, err)
	}
	stats = testDB.Stats()
	if stats.NumTerms != startStats.NumTerms+1 {
		t.Fatalf("Store.Stats().NumTerms == %d; want %d", stats.NumTerms, startStats.NumTerms+1)
	}
	if stats.NumTriples != startStats.NumTriples+1 {
		t.Fatalf("Store.Stats().NumTriples == %d; want %d", stats.NumTriples, startStats.NumTriples+1)
	}
}

func TestQuery(t *testing.T) {
	s := mustNewIRI("s10")
	g := rdf.Graph{
		rdf.NewTriple(s, mustNewIRI("p1"), mustNewLiteral("o1")),
		rdf.NewTriple(s, mustNewIRI("p1"), mustNewLiteral("o2")),
		rdf.NewTriple(s, mustNewIRI("p2"), mustNewLiteral("o1")),
		rdf.NewTriple(s, mustNewIRI("p3"), mustNewLiteral("o1")),
	}

	err := testDB.ImportGraph(g)
	if err != nil {
		t.Fatalf("Store.LoadGraph(%v) == %v; want no error", g, err)
	}

	res, err := testDB.Query(NewQuery().Resource(s))
	if err != nil {
		t.Fatalf("Store.Query(NewQuery().Resource(%v)) == %v; want no error", s, err)
	}

	if !res.Eq(g) {
		t.Fatalf("Store.Query(NewQuery().Resource(%v)) == \n%v\nwant:\n%v", s, res, g)
	}
}
