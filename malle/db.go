package malle

import (
	"encoding/binary"
	"errors"
	"log"

	"github.com/boltdb/bolt"
	"github.com/boutros/x/malle/rdf"
)

const (
	// MaxTerms is the maximum number of RDF terms that can be stored.
	MaxTerms = 4294967295
)

// buckets in the key-value store
var (
	bTerms    = []byte("terms")
	bIdxTerms = []byte("iterms")
	bDT       = []byte("dt")
	bIdxDT    = []byte("idt")
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
	*bolt.DB
}

// Public API -----------------------------------------------------------------

// Init opens a new or existing database file, sets up buckets and indices and
// makes it ready for reading and writing.
func Init(file string) (*Store, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	s := &Store{db}
	return s.setup()
}

// Close closes the datastore, relasing the lock on the database file.
func (db *Store) Close() error {
	return db.DB.Close()
}

// AddTerm stores a rdf.Term in the database and returns the id it has been given.
// If the term allready is stored, it will simply return the id.
func (db *Store) AddTerm(term rdf.Term) (uint32, error) {
	var id uint32
	err := db.Update(func(tx *bolt.Tx) error {
		if i, err := db.getID(tx, term); err == nil {
			// Term is allready in database, return it's id
			id = i
			return nil
		} else if err != ErrNotFound {
			// Some other IO error occured
			return err
		}
		bkt := tx.Bucket(bTerms)
		n, err := bkt.NextSequence()
		if err != nil {
			log.Println(err)
			return ErrDBFailure
		}
		// TODO err if id > max uint32 = 4294967295
		id = uint32(n)
		idb := u32tob(uint32(n))
		err = bkt.Put(idb, db.encode(term))
		if err != nil {
			return err
		}
		bkt = tx.Bucket(bIdxTerms)
		err = bkt.Put(db.encode(term), idb)
		return err
	})
	return id, err
}

// HasTerm checks if the given term is stored.
func (db *Store) HasTerm(term rdf.Term) (bool, error) {
	found := false
	err := db.View(func(tx *bolt.Tx) error {
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
	err := db.View(func(tx *bolt.Tx) error {
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
	err = db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bIdxTerms)
		b := bkt.Get(db.encode(term))
		if b == nil {
			return ErrNotFound
		}
		id = btou32(b)
		return nil
	})
	return id, err
}

/*

func (db *Store) RemoveTerm(rdf.Term) error      {}

func (db *Store) AddTriple(rdf.Triple) error         {}
func (db *Store) RemoveTriple(rdf.Triple) error      {}
func (db *Store) HasTriple(rdf.Triple) (bool, error) {}

type Query struct {
	err  error
	subj rdf.IRI
}

//res, err := db.Query().WhereSubj(rdf.Term...).Limit(10)
*/

// Unexported methods ---------------------------------------------------------

// setup makes sure the database has all the needed buckets and predefined values
func (db *Store) setup() (*Store, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bTerms, bIdxTerms, bDT, bIdxDT} {
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

// func (db *Store) decode([]byte]) rdf.Term {}

/*
func (db *Store) hasTerm(tx *bolt.Tx, rdf.Term) (bool, error) {}
func (db *Store) getID(tx )
*/

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
