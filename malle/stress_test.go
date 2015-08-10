package malle

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/tgruben/roaring"
)

func TestStressAddAndRemoveTriple(t *testing.T) {
	t.Skip() // disabled
	startStats := testDB.Stats()
	g := genRandGraph(1000)
	for _, tr := range g.Triples() {

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

	for _, tr := range g.Triples() {
		exists, err := testDB.HasTriple(tr)
		if err != nil || !exists {
			t.Fatalf("Store.HasTriple(%v) == %v, %v; want true, nil", tr, exists, err)
		}

		err = testDB.RemoveTriple(tr)
		if err != nil {
			t.Fatalf("Store.RemoveTriple(%v) == %v; want no error", tr, err)
		}

		exists, err = testDB.HasTriple(tr)
		if err != nil || exists {
			t.Fatalf("Store.HasTriple(%v) == %v, %v; want false, nil", tr, exists, err)
		}

	}

	for _, tr := range g.Triples() {
		var err error
		err = testDB.kv.View(func(tx *bolt.Tx) error {
			_, err = testDB.getID(tx, tr.Subject())
			return err
		})
		if err != ErrNotFound {
			t.Errorf("Store.RemoveTriple(%v) didn't remove all terms (subject)", tr)
		}

		err = testDB.kv.View(func(tx *bolt.Tx) error {
			_, err = testDB.getID(tx, tr.Predicate())
			return err
		})
		if err != ErrNotFound {
			t.Errorf("Store.RemoveTriple(%v) didn't remove all terms (predicate)", tr)

		}

		err = testDB.kv.View(func(tx *bolt.Tx) error {
			_, err = testDB.getID(tx, tr.Object())
			return err
		})
		if err != ErrNotFound {
			t.Errorf("Store.RemoveTriple(%v) didn't remove all terms (object)", tr)
		}
	}

	endStats := testDB.Stats()
	if startStats.NumTerms != endStats.NumTerms && startStats.NumTriples != endStats.NumTriples {
		t.Errorf("RemoveGraph(G) didn't restore db to state before ImportGraph(G): before: %v != after: %v",
			startStats, endStats)
	}

}
