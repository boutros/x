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

const seed = 0x123

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

func TestEncodeDecode(t *testing.T) {
	nilDB := Store{}
	tests := []rdf.Term{
		mustNewIRI("a"),
		mustNewIRI("http://example.org/1/xyz.æøå"),
		mustNewLangLiteral("a", "en"),
		mustNewLangLiteral("æøå", "nb-no"),
	}
	for _, term := range tests {
		if !nilDB.decode(nilDB.encode(term)).Eq(term) {
			t.Errorf("Store.encode/decode roundtrip failed for %+v", term)
		}
	}
}

func TestAddTerm(t *testing.T) {
	term := genRandTerm()
	var err error
	var id uint32
	err = testDB.kv.Update(func(tx *bolt.Tx) error {
		id, err = testDB.addTerm(tx, term)
		return err
	})
	if err != nil {
		t.Fatalf("Store.addTerm(tx, %v)) == %v; want no error", term, err)
	}

	var stored bool
	err = testDB.kv.View(func(tx *bolt.Tx) error {
		stored = testDB.hasTerm(tx, term)
		return nil
	})

	if !stored {
		t.Fatalf("Store.addTerm(%v) didn't store term in database", term)
	}

	var want rdf.Term
	err = testDB.kv.View(func(tx *bolt.Tx) error {
		want, err = testDB.getTerm(tx, id)
		return err
	})
	if err != nil {
		t.Fatalf("Store.getTerm(%v)) == %v; want no error", id, err)
	}

	if !term.Eq(want) {
		t.Fatal("Store.addTerm returned wrong id")
	}

	var id2 uint32
	err = testDB.kv.Update(func(tx *bolt.Tx) error {
		id2, err = testDB.addTerm(tx, term)
		return err
	})
	if err != nil {
		t.Fatalf("Store.addTerm(%v)) == %v; want no error", term, err)
	}

	if id2 != id {
		t.Errorf("Store.addTerm(%v)) stored exisiting term as a new term", term)
	}
}

func TestRemoveTerm(t *testing.T) {
	term := genRandTerm()
	var err error
	var id uint32
	err = testDB.kv.Update(func(tx *bolt.Tx) error {
		id, err = testDB.addTerm(tx, term)
		return err
	})
	if err != nil {
		t.Fatalf("Store.addTerm(tx, %v)) == %v; want no error", term, err)
	}

	err = testDB.kv.Update(func(tx *bolt.Tx) error {
		err = testDB.removeTerm(tx, id)
		return err
	})

	if err != nil {
		t.Fatalf("Store.removeTerm(%v)) == %v; want no error", id, err)
	}

	err = testDB.kv.Update(func(tx *bolt.Tx) error {
		err = testDB.removeTerm(tx, id)
		return err
	})
	if err != ErrNotFound {
		t.Fatalf("Store.removeTerm(%v)) == %v; want no ErrNotFound", id, err)
	}

	var found bool
	err = testDB.kv.View(func(tx *bolt.Tx) error {
		found = testDB.hasTerm(tx, term)
		return nil
	})
	if found {
		t.Fatalf("Store.RemoveTerm(%v) didn't remove term from database", term)
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

	var s, p, o uint32
	err = testDB.kv.View(func(tx *bolt.Tx) error {
		s, err = testDB.getID(tx, tr.Subject())
		return err
	})
	if err != nil {
		t.Logf("Store.AddTriple(%v) didn't store all terms", tr)
		t.Fatalf("Store.GetID(%v) == %v; want no error", s, err)
	}

	err = testDB.kv.View(func(tx *bolt.Tx) error {
		p, err = testDB.getID(tx, tr.Predicate())
		return err
	})
	if err != nil {
		t.Logf("Store.AddTriple(%v) didn't store all terms", tr)
		t.Fatalf("Store.GetID(%v) == %v; want no error", p, err)
	}

	err = testDB.kv.View(func(tx *bolt.Tx) error {
		o, err = testDB.getID(tx, tr.Object())
		return err
	})
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

func TestResourceQuery(t *testing.T) {
	s := mustNewIRI("s10")
	g := rdf.NewGraph().
		Add(rdf.NewTriple(s, mustNewIRI("p1"), mustNewLiteral("o1"))).
		Add(rdf.NewTriple(s, mustNewIRI("p1"), mustNewLiteral("o2"))).
		Add(rdf.NewTriple(s, mustNewIRI("p2"), mustNewLiteral("o1"))).
		Add(rdf.NewTriple(s, mustNewIRI("p3"), mustNewLiteral("o1")))

	err := testDB.ImportTriples(g.Triples())
	if err != nil {
		t.Fatalf("Store.ImportTriples(%v) == %v; want no error", g, err)
	}

	res, err := testDB.Query(NewQuery().Resource(s))
	if err != nil {
		t.Fatalf("Store.Query(NewQuery().Resource(%v)) == %v; want no error", s, err)
	}

	if !res.Eq(g) {
		t.Fatalf("Store.Query(NewQuery().Resource(%v)) == \n%v\nwant:\n%v", s, res, g)
	}
}

func TestCBDQuery(t *testing.T) {
	graph := `<z1> <p1> <z2> .
<z1> <p2> "a" .
<z1> <p3> "b" .
<z1> <p4> "c" .
<z2> <p2> "f" .
<z3> <p1> <z1> .
<z3> <p2> "x" .
`

	_, err := testDB.Import(bytes.NewBufferString(graph), 10, false)
	if err != nil {
		t.Fatalf("Store.Import(%s) == %v; want no error", graph, err)
	}
	s := mustNewIRI("z1")
	q := NewQuery().CBD(s, 0)

	want := rdf.Load(bytes.NewBufferString(`<z1> <p1> <z2> .
<z1> <p2> "a" .
<z1> <p3> "b" .
<z1> <p4> "c" .
<z3> <p1> <z1> .`))

	res, err := testDB.Query(q)
	if err != nil || !want.Eq(res) {
		t.Fatalf("Store.Query(NewQuery().CBD(%v, 0)) == %v, %v; want %v, <nil>", s, res, err, want)
	}

}

func TestImport(t *testing.T) {
	graph := `<s_1> <p_1> <o_1> .
<s_1> <p_1> "abc . # invalid triple
<s_2> z f . # another invalid triple
_:b1 <p_2> <o_1> . # triples with blank node are ignored
<s_1> <p_2> "oz"@fr .
# a blank line
<s_11> <p_1> <o_1> .`

	n, err := testDB.Import(bytes.NewBufferString(graph), 10, false)
	if err != nil || n != 3 {
		t.Fatalf("Store.Import(%s) == %d, %v; want 3, <nil>", graph, n, err)
	}
}
