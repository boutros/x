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
)

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
	setupDB(db)
	return &Store{db}, nil
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
		if i, err := db.GetID(term); err == nil {
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
		b := make([]byte, len(term.Bytes())+1)
		b[0] = '0'
		copy(b[1:], term.Bytes())
		err = bkt.Put(b, idb)
		if err != nil {
			return err
		}
		bkt = tx.Bucket(bIdxTerms)
		err = bkt.Put(idb, b)
		return err
	})
	return id, err
}

// HasTerm checks if the given term is stored.
func (db *Store) HasTerm(term rdf.Term) (bool, error) {
	found := false
	err := db.View(func(tx *bolt.Tx) error {
		b := make([]byte, len(term.Bytes())+1)
		b[0] = '0'
		copy(b[1:], term.Bytes())
		bkt := tx.Bucket(bTerms)
		term := bkt.Get(b)
		if term != nil {
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
		bkt := tx.Bucket(bIdxTerms)
		b := bkt.Get(u32tob(id))
		if b == nil {
			return ErrNotFound
		}
		var err error
		term, err = rdf.ReadTerm(b)
		return err
	})
	return term, err
}

// GetID returns the ID of a given term.
func (db *Store) GetID(term rdf.Term) (id uint32, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bTerms)
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

// Unexported methods ---------------------------------------------------------s

func setupDB(db *bolt.DB) {
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bTerms, bIdxTerms} {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func (db *Store) encode(term rdf.Term) []byte {
	switch term.(type) {
	case rdf.IRI:
		b := make([]byte, len(term.Bytes())+1)
		b[0] = '0'
		copy(b[1:], term.Bytes())
		return b
	//case rdf.Literal:
	default:
		panic("todo")
	}
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
