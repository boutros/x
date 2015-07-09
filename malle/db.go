package malle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"sync/atomic"

	"github.com/boltdb/bolt"
	"github.com/boutros/x/malle/rdf"
	"github.com/tgruben/roaring"
)

const (
	// MaxTerms is the maximum number of RDF terms that can be stored.
	MaxTerms = 4294967295
)

// buckets in the key-value store
var (
	// Terms:
	bTerms    = []byte("terms")  // uint32 -> term
	bIdxTerms = []byte("iterms") // term -> uint32
	bDT       = []byte("dt")     // uint32 -> iri
	bIdxDT    = []byte("idt")    // iri -> uint32
	// Triple indices:
	bSPO = []byte("spo") // Subect + Predicate -> bitmap of Object
	bOSP = []byte("osp") // Object + Subject   -> bitmap of Predicate
	bPOS = []byte("pos") // PredicateO + bject -> bitmap of Subject
)

// datatypes are the built in datatypes (IDs 0 through 41).
// (the RDF-compatible XSD types plus rdf:langString, rdf:HTML and rdf:XMLLiteral)
var datatypes = []string{
	"IRI", // Dummy value, in DB it indicates and IRI, not literal
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#langString",
	"http://www.w3.org/2001/XMLSchema#string",
	"http://www.w3.org/2001/XMLSchema#boolean",
	"http://www.w3.org/2001/XMLSchema#decimal",
	"http://www.w3.org/2001/XMLSchema#integer",
	"http://www.w3.org/2001/XMLSchema#double",
	"http://www.w3.org/2001/XMLSchema#float",
	"http://www.w3.org/2001/XMLSchema#date",
	"http://www.w3.org/2001/XMLSchema#time",
	"http://www.w3.org/2001/XMLSchema#dateTime",
	"http://www.w3.org/2001/XMLSchema#dateTimeStamp",
	"http://www.w3.org/2001/XMLSchema#gYear",
	"http://www.w3.org/2001/XMLSchema#gMonth",
	"http://www.w3.org/2001/XMLSchema#gDay",
	"http://www.w3.org/2001/XMLSchema#gYearMonth",
	"http://www.w3.org/2001/XMLSchema#gMonthDay",
	"http://www.w3.org/2001/XMLSchema#duration",
	"http://www.w3.org/2001/XMLSchema#yearMonthDuration",
	"http://www.w3.org/2001/XMLSchema#dayTimeDuration",
	"http://www.w3.org/2001/XMLSchema#byte",
	"http://www.w3.org/2001/XMLSchema#short",
	"http://www.w3.org/2001/XMLSchema#int",
	"http://www.w3.org/2001/XMLSchema#long",
	"http://www.w3.org/2001/XMLSchema#unsignedByte",
	"http://www.w3.org/2001/XMLSchema#unsignedShort",
	"http://www.w3.org/2001/XMLSchema#unsignedInt",
	"http://www.w3.org/2001/XMLSchema#unsignedLong",
	"http://www.w3.org/2001/XMLSchema#positiveInteger",
	"http://www.w3.org/2001/XMLSchema#nonNegativeInteger",
	"http://www.w3.org/2001/XMLSchema#negativeInteger",
	"http://www.w3.org/2001/XMLSchema#nonPositiveInteger",
	"http://www.w3.org/2001/XMLSchema#hexBinary",
	"http://www.w3.org/2001/XMLSchema#base64Binary",
	"http://www.w3.org/2001/XMLSchema#anyURI",
	"http://www.w3.org/2001/XMLSchema#language",
	"http://www.w3.org/2001/XMLSchema#normalizedString",
	"http://www.w3.org/2001/XMLSchema#token",
	"http://www.w3.org/2001/XMLSchema#NMTOKEN",
	"http://www.w3.org/2001/XMLSchema#Name",
	"http://www.w3.org/2001/XMLSchema#CName",
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML",
	"http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral",
}

// Exported errors
var (
	ErrDBFailure = errors.New("database error")
	ErrNotFound  = errors.New("not found")
)

// Store is a RDF triple store backed by a key-value store (boltdb).
type Store struct {
	kv *bolt.DB

	numTr int64 // number of triples stored
}

// Stats holds some statistics of the triple store.
type Stats struct {
	NumTerms   int
	NumTriples int
	// SizeInBytes int
}

// Public API -----------------------------------------------------------------

// Init opens a new or existing database file, sets up buckets and indices and
// makes it ready for reading and writing.
func Init(file string) (*Store, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	s := &Store{kv: db}
	return s.setup()
}

// Close closes the datastore, relasing the lock on the database file.
func (db *Store) Close() error {
	return db.kv.Close()
}

// Stats return statistics about the triple store.
func (db *Store) Stats() Stats {
	st := Stats{}
	db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bTerms)
		st.NumTerms = bkt.Stats().KeyN
		bkt = tx.Bucket(bSPO)
		st.NumTriples = int(atomic.LoadInt64(&db.numTr))
		return nil
	})
	return st
}

// AddTerm stores a rdf.Term in the database and returns the id it has been given.
// If the term allready is stored, it will simply return the id.
func (db *Store) AddTerm(term rdf.Term) (id uint32, err error) {
	err = db.kv.Update(func(tx *bolt.Tx) error {
		id, err = db.addTerm(tx, term)
		return err
	})
	return id, err
}

// HasTerm checks if the given term is stored.
func (db *Store) HasTerm(term rdf.Term) (bool, error) {
	found := false
	err := db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bIdxTerms)
		id := bkt.Get(db.encode(term))
		if id != nil {
			found = true
		}
		return nil
	})
	return found, err
}

// GetTerm returns the term for a given ID.
func (db *Store) GetTerm(id uint32) (rdf.Term, error) {
	var term rdf.Term
	err := db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bTerms)
		b := bkt.Get(u32tob(id))
		if b == nil {
			return ErrNotFound
		}
		term = db.decode(b)
		return nil
	})
	return term, err
}

// GetID returns the ID of a given term.
func (db *Store) GetID(term rdf.Term) (id uint32, err error) {
	err = db.kv.View(func(tx *bolt.Tx) error {
		id, err = db.getID(tx, term)
		return err
	})
	return id, err
}

// RemoveTerm removes a rdf.Term from the triple store.
func (db *Store) RemoveTerm(term rdf.Term) error {
	err := db.kv.Update(func(tx *bolt.Tx) error {
		bt := db.encode(term)
		bkt := tx.Bucket(bIdxTerms)
		id := bkt.Get(bt)
		if id == nil {
			return ErrNotFound
		}
		err := bkt.Delete(bt)
		if err != nil {
			return err
		}
		bkt = tx.Bucket(bTerms)
		err = bkt.Delete(id)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// AddTriple stores the given Triple.
func (db *Store) AddTriple(tr rdf.Triple) error {
	err := db.kv.Update(func(tx *bolt.Tx) error {
		sID, err := db.addTerm(tx, tr.Subject())
		if err != nil {
			return err
		}

		pID, err := db.addTerm(tx, tr.Predicate())
		if err != nil {
			return err
		}

		oID, err := db.addTerm(tx, tr.Object())
		if err != nil {
			return err
		}

		return db.storeTriple(tx, sID, pID, oID)
	})
	return err
}

// RemoveTriple removes the given Triple from the indices. It also removes
// any Term unique to that Triple from the store.
// It return ErrNotFound if the Triple does not exist
func (db *Store) RemoveTriple(tr rdf.Triple) error {
	err := db.kv.Update(func(tx *bolt.Tx) error {
		sID, err := db.getID(tx, tr.Subject())
		if err != nil {
			return err
		}

		pID, err := db.getID(tx, tr.Predicate())
		if err != nil {
			return err
		}

		oID, err := db.getID(tx, tr.Object())
		if err != nil {
			return err
		}

		return db.removeTriple(tx, sID, pID, oID)
	})
	return err
}

// HasTriple checks if the given Triple is stored.
func (db *Store) HasTriple(tr rdf.Triple) (exists bool, err error) {
	err = db.kv.View(func(tx *bolt.Tx) error {
		sID, err := db.getID(tx, tr.Subject())
		if err == ErrNotFound {
			return nil
		} else if err != nil {
			return err
		}
		pID, err := db.getID(tx, tr.Predicate())
		if err == ErrNotFound {
			return nil
		} else if err != nil {
			return err
		}
		oID, err := db.getID(tx, tr.Object())
		if err == ErrNotFound {
			return nil
		} else if err != nil {
			return err
		}

		bkt := tx.Bucket(bSPO)

		sp := make([]byte, 8)
		copy(sp, u32tob(sID))
		copy(sp[4:], u32tob(pID))

		bitmap := roaring.NewRoaringBitmap()
		bo := bkt.Get(sp)
		if bo == nil {
			return nil
		}

		_, err = bitmap.ReadFrom(bytes.NewReader(bo))
		if err != nil {
			return err
		}

		exists = bitmap.Contains(oID)
		return nil
	})
	return exists, err
}

// Unexported methods ---------------------------------------------------------

// setup makes sure the database has all the needed buckets and predefined values
func (db *Store) setup() (*Store, error) {
	err := db.kv.Update(func(tx *bolt.Tx) error {
		// Make sure all the required buckets are created
		for _, b := range [][]byte{bTerms, bIdxTerms, bDT, bIdxDT, bSPO, bOSP, bPOS} {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}
		// Make sure the predefined datatypes are stored
		bkt := tx.Bucket(bDT)
		cur := bkt.Cursor()
		i := 0
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			if i != int(btou32(k)) || datatypes[i] != string(v) {
				// store datatype
				// err :=
				// TODO!
			}
			i++
		}

		// Count number of triples
		bkt = tx.Bucket(bSPO)
		cur = bkt.Cursor()
		bitmap := roaring.NewRoaringBitmap()
		var n uint64
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			if v != nil {
				_, err := bitmap.ReadFrom(bytes.NewReader(v))
				if err != nil {
					return err
				}
				n += bitmap.GetCardinality()
			} // else ?
		}
		db.numTr = int64(n)

		return nil
	})
	return db, err
}

func (db *Store) encode(term rdf.Term) []byte {
	return term.Encode()
}

func (db *Store) decode(b []byte) rdf.Term {
	t, err := rdf.DecodeTerm(b)
	if err != nil {
		// We control the encoding, so it shouldn't be possible to store
		// an undecodable term. TODO remove this when confident.
		panic(err)
	}
	return t
}

// getID works like the exported GetID, but using the given transaction.
func (db *Store) getID(tx *bolt.Tx, term rdf.Term) (id uint32, err error) {
	bkt := tx.Bucket(bIdxTerms)
	b := bkt.Get(db.encode(term))
	if b == nil {
		err = ErrNotFound
	} else {
		id = btou32(b)
	}
	return id, err
}

// addTerm works like the exported AddTerm, but using the given transaction.
func (db *Store) addTerm(tx *bolt.Tx, term rdf.Term) (id uint32, err error) {
	if id, err = db.getID(tx, term); err == nil {
		// Term is allready in database
		return id, nil
	} else if err != ErrNotFound {
		// Some other IO error occured
		return uint32(0), err
	}
	bkt := tx.Bucket(bTerms)
	n, err := bkt.NextSequence()
	if err != nil {
		log.Println(err)
		return uint32(0), ErrDBFailure
	}
	// TODO err if id > max uint32 = 4294967295
	id = uint32(n)
	idb := u32tob(uint32(n))
	bt := db.encode(term)
	err = bkt.Put(idb, bt)
	if err != nil {
		return uint32(0), err
	}
	bkt = tx.Bucket(bIdxTerms)
	err = bkt.Put(bt, idb)
	return id, err
}

// storeTriple stores a triple in the indices.
func (db *Store) storeTriple(tx *bolt.Tx, s, p, o uint32) error {
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

	newTriple := false
	key := make([]byte, 8)

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

		newTriple = bitmap.CheckedAdd(i.v)
		if newTriple {
			var b bytes.Buffer
			_, err := bitmap.WriteTo(&b)
			if err != nil {
				return err
			}
			err = bkt.Put(key, b.Bytes())
			if err != nil {
				return err
			}
		}
	}

	if newTriple {
		atomic.AddInt64(&db.numTr, 1)
	}

	return nil
}

// removeTriple removes a triple from the indices. If the triple
// contains any terms unique to that triple, they will also be removed.
func (db *Store) removeTriple(tx *bolt.Tx, s, p, o uint32) error {
	// TODO think about what to do if present in one index but
	// not in another: maybe panic? Cause It's a bug that should be fixed.

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

	hasTriple := false
	key := make([]byte, 8)

	for _, i := range indices {
		bkt := tx.Bucket(i.bk)
		copy(key, u32tob(i.k1))
		copy(key[4:], u32tob(i.k2))

		bitmap := roaring.NewRoaringBitmap()

		bo := bkt.Get(key)

		if bo == nil {
			return ErrNotFound
		}
		_, err := bitmap.ReadFrom(bytes.NewReader(bo))
		if err != nil {
			return err
		}

		hasTriple = bitmap.CheckedRemove(i.v)
		if hasTriple {
			var b bytes.Buffer
			_, err := bitmap.WriteTo(&b)
			if err != nil {
				return err
			}
			err = bkt.Put(key, b.Bytes())
			if err != nil {
				return err
			}
		} // else?
	}

	if hasTriple {
		atomic.AddInt64(&db.numTr, -1)
	}

	return nil
}

// Helper functions -----------------------------------------------------------

// u32tob converts a uint32 into an 4-byte slice.
func u32tob(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// btou32 converts an 4-byte slice into an uint32.
func btou32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}
