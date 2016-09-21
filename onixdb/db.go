package onixdb

import (
	"encoding/binary"
	"errors"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/knakk/kbp/onix"
)

var (
	ErrNotFound = errors.New("not found")
	ErrDBFull   = errors.New("database full: id limit reached")
)

const MaxProducts = 4294967295

type DB struct {
	kv      *bolt.DB
	encPool sync.Pool
	decPool sync.Pool
}

func Open(path string) (*DB, error) {
	kv, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	codec, err := newPrimedCodec(&onix.Product{})
	if err != nil {
		return nil, err
	}
	db := &DB{
		kv:      kv,
		encPool: sync.Pool{New: func() interface{} { return codec.NewMarshaler() }},
		decPool: sync.Pool{New: func() interface{} { return codec.NewUnmarshaler() }},
	}
	return db.setup()
}

func (db *DB) Close() error {
	return db.kv.Close()
}

func (db *DB) setup() (*DB, error) {
	// set up required buckets
	err := db.kv.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{[]byte("products"), []byte("indices")} {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return db, err
}

// Get will retrieve the Product with the give ID, if it exists.
func (db *DB) Get(id uint32) (*onix.Product, error) {
	var b []byte
	if err := db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("products"))
		b = bkt.Get(u32tob(id))
		if b == nil {
			return ErrNotFound
		}
		return nil
	}); err != nil {
		return nil, err
	}
	dec := db.decPool.Get().(*primedDecoder)
	defer db.decPool.Put(dec)
	return dec.Unmarshal(b)
}

// Store will persist an onix.Product in the database, returning the ID it
// was assigned.
func (db *DB) Store(p *onix.Product) (id uint32, err error) {
	err = db.kv.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("products"))
		n, _ := bkt.NextSequence()
		if n > MaxProducts {
			return ErrDBFull
		}

		id = uint32(n)
		idb := u32tob(uint32(n))
		enc := db.encPool.Get().(*primedEncoder)
		defer db.encPool.Put(enc)
		b, err := enc.Marshal(p)
		if err != nil {
			return err
		}
		if err := bkt.Put(idb, b); err != nil {
			return err
		}
		return nil
	})
	return id, err
}

// u32tob converts a uint32 into a 4-byte slice.
func u32tob(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// btou32 converts a 4-byte slice into an uint32.
func btou32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

/*

type IndexEntry struct {
	Index string
	Entry string
}

type IndexFn func(onix.Product) []IndexEntry

*/
