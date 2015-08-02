package malle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
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
	bPOS = []byte("pos") // Predicate + Object -> bitmap of Subject
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
	NumTerms    int
	NumTriples  int
	File        string
	SizeInBytes int
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
		st.File = db.kv.Path()
		s, err := os.Stat(st.File)
		if err == nil {
			st.SizeInBytes = int(s.Size())
		}
		return nil
	})
	return st
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

// ImportGraph imports the graph into the triple store.
func (db *Store) ImportGraph(g rdf.Graph) (err error) {
	err = db.kv.Update(func(tx *bolt.Tx) error {
		for subj, props := range g {

			sID, err := db.addTerm(tx, subj)
			if err != nil {
				return err
			}

			for pred, terms := range props {
				pID, err := db.addTerm(tx, pred)
				if err != nil {
					return err
				}

				for _, obj := range terms {
					// TODO batch bitmap operations for all obj in terms
					oID, err := db.addTerm(tx, obj)
					if err != nil {
						return err
					}

					err = db.storeTriple(tx, sID, pID, oID)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	return err
}

// ImportTriples stores all the triples in the store.
func (db *Store) ImportTriples(triples []rdf.Triple) (err error) {
	err = db.kv.Update(func(tx *bolt.Tx) error {
		for _, tr := range triples {
			// TODO optimize this loop:
			// * the triples are sorted, we can reuse subject/predicate from previous iteration
			// * if a subject has more than on value of a predicate, it should batch the bitmap updates.

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

			err = db.storeTriple(tx, sID, pID, oID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// Import imports triples from an N-Triples stream, in batches of given size.
// It will ignore triples with blank nodes and errors. If the logErr flag is set it will log
// such incidents. It returns the total number of triples imported (regardless if they where in the
// store before or not)
func (db *Store) Import(r io.Reader, batchSize int, logErr bool) (int, error) {
	dec := rdf.NewNTDecoder(r)
	triples := make([]rdf.Triple, 0, batchSize+1)
	c := 0 // totalt count
	i := 0 // current batch count
	for tr, err := dec.Decode(); err != io.EOF; tr, err = dec.Decode() {
		if err != nil {
			if logErr {
				log.Println(err.Error())
			}
			continue
		}
		triples = append(triples, tr)
		i++
		if i == batchSize {
			err = db.ImportTriples(triples)
			if err != nil {
				return c, err
			}
			c += i
			i = 0
			triples = triples[0:0]
		}
	}
	if len(triples) > 0 {
		err := db.ImportTriples(triples)
		if err != nil {
			return c, err
		}
		c += i
	}
	return c, nil
}

// Query represents a query into the triple store.
// A query always returns a rdf.Graph.
type Query struct {
	subj  rdf.IRI // starting node
	depth int
}

// NewQuery returns a new Query.
func NewQuery() *Query {
	return &Query{}
}

// Resource returns a query asking for all triples with the given
// IRI as subject. (Same as SPARQL DESCRIBE)
func (q *Query) Resource(s rdf.IRI) *Query {
	q.subj = s
	q.depth = -1
	return q
}

// CBD returns a Consise Bounded Description query. A CBD query will
// return a graph of all the statements where the given IRI is subject
// or object, and recursivly up to a given depth for any other IRI in the graph
// when depth is > 0.
func (q *Query) CBD(s rdf.IRI, depth int) *Query {
	q.subj = s
	if depth < 0 {
		depth = 0
	}
	q.depth = depth
	return q
}

// Query executes the query against the triple store, returning a graph
// of the matching triples.
func (db *Store) Query(q *Query) (g rdf.Graph, err error) {
	g = rdf.NewGraph()
	err = db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bIdxTerms)
		bs := bkt.Get(q.subj.Bytes())
		if bs == nil {
			return ErrNotFound
		}
		// seek in SPO index
		sid := btou32(bs)
		cur := tx.Bucket(bSPO).Cursor()
	outerSPO:
		for k, v := cur.Seek(u32tob(sid - 1)); k != nil; k, v = cur.Next() {
			switch bytes.Compare(k[:4], bs) {
			case 0:
				bkt = tx.Bucket(bTerms)
				b := bkt.Get(k[4:])
				if b == nil {
					panic("term should be there!")
				}
				pred := db.decode(b)
				bitmap := roaring.NewRoaringBitmap()
				_, err = bitmap.ReadFrom(bytes.NewReader(v))
				if err != nil {
					return err
				}
				it := bitmap.Iterator()
				for it.HasNext() {
					o := it.Next()
					b = bkt.Get(u32tob(o))
					if b == nil {
						panic("term should be there!")
					}
					g.Add(rdf.NewTriple(q.subj, pred.(rdf.IRI), db.decode(b)))
				}
			case 1:
				break outerSPO
			}
		}

		// If doing a CBD query:
		if q.depth == 0 {
			// seek in OSP index

			cur := tx.Bucket(bOSP).Cursor()
		outerOSP:
			for k, v := cur.Seek(u32tob(sid - 1)); k != nil; k, v = cur.Next() {
				switch bytes.Compare(k[:4], bs) {
				case 0:
					bkt = tx.Bucket(bTerms)
					b := bkt.Get(k[4:])
					if b == nil {
						panic("term should be there!")
					}
					subj := db.decode(b)
					bitmap := roaring.NewRoaringBitmap()
					_, err = bitmap.ReadFrom(bytes.NewReader(v))
					if err != nil {
						return err
					}
					it := bitmap.Iterator()
					for it.HasNext() {
						o := it.Next()
						b = bkt.Get(u32tob(o))
						if b == nil {
							panic("term should be there!")
						}
						g.Add(rdf.NewTriple(subj.(rdf.IRI), db.decode(b).(rdf.IRI), q.subj))
					}
				case 1:
					break outerOSP
				}
			}

		} else if q.depth > 0 {
			panic("TODO query CBD with depth > 0")
		}

		return nil
	})
	if err == ErrNotFound {
		return g, nil
	}

	return g, err
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

		var n uint64
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			if v != nil {
				bitmap := roaring.NewRoaringBitmap()
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
	return term.Bytes()
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

// getTerm returns the term for a given ID.
func (db *Store) getTerm(tx *bolt.Tx, id uint32) (rdf.Term, error) {
	var term rdf.Term
	bkt := tx.Bucket(bTerms)
	b := bkt.Get(u32tob(id))
	if b == nil {
		return term, ErrNotFound
	}
	term = db.decode(b)
	return term, nil
}

// hasTerm returns trie if the term exists.
func (db *Store) hasTerm(tx *bolt.Tx, term rdf.Term) bool {
	bkt := tx.Bucket(bIdxTerms)
	id := bkt.Get(db.encode(term))
	if id != nil {
		return true
	}
	return false
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

		newTriple := bitmap.CheckedAdd(i.v)
		if !newTriple {
			return nil
		}
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
	atomic.AddInt64(&db.numTr, 1)

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
		hasTriple := bitmap.CheckedRemove(i.v)
		if !hasTriple {
			return ErrNotFound
		}
		// Remove from index if bitmap is empty
		if bitmap.GetCardinality() == 0 {
			err = bkt.Delete(key)
			if err != nil {
				return err
			}
		} else {
			var b bytes.Buffer
			_, err = bitmap.WriteTo(&b)
			if err != nil {
				return err
			}
			err = bkt.Put(key, b.Bytes())
			if err != nil {
				return err
			}
		}
	}

	atomic.AddInt64(&db.numTr, -1)

	return db.removeOrphanedTerms(tx, s, p, o)
}

// removeOrphanedTerms removes any of the given Terms if they are no longer
// part of any triple.
func (db *Store) removeOrphanedTerms(tx *bolt.Tx, s, p, o uint32) error {
	var err error
	cur := tx.Bucket(bSPO).Cursor()
	for k, _ := cur.Seek(u32tob(s - 1)); k != nil; k, _ = cur.Next() {
		switch bytes.Compare(u32tob(s), k[:4]) {
		case 0:
			goto checkP
		case -1:
			goto removeS
		}
	}
removeS:
	err = db.removeTerm(tx, s)
	if err != nil {
		return err
	}
checkP:
	cur = tx.Bucket(bPOS).Cursor()
	for k, _ := cur.Seek(u32tob(p - 1)); k != nil; k, _ = cur.Seek(u32tob(p - 1)) {
		switch bytes.Compare(u32tob(p), k[:4]) {
		case 0:
			goto checkO
		case -1:
			goto removeP
		}
	}
removeP:
	err = db.removeTerm(tx, p)
	if err != nil {
		return err
	}
checkO:
	cur = tx.Bucket(bOSP).Cursor()
	for k, _ := cur.Seek(u32tob(o - 1)); k != nil; k, _ = cur.Seek(u32tob(o - 1)) {
		switch bytes.Compare(u32tob(o), k[:4]) {
		case 0:
			return nil
		case -1:
			goto removeO
		}
	}
removeO:
	return db.removeTerm(tx, o)
}

// removeTerm removes a Term using the given transaction.
func (db *Store) removeTerm(tx *bolt.Tx, termID uint32) error {
	bkt := tx.Bucket(bTerms)
	term := bkt.Get(u32tob(termID))
	if term == nil {
		// TODO log or panic
		return ErrNotFound
	}
	err := bkt.Delete(u32tob(termID))
	if err != nil {
		return err
	}
	bkt = tx.Bucket(bIdxTerms)
	err = bkt.Delete(term)
	if err != nil {
		return err
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
