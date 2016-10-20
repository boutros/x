package onixdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/RoaringBitmap/roaring"
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
	indexFn IndexFn
}

func Open(path string, fn IndexFn) (*DB, error) {
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
		indexFn: fn,
	}
	return db.setup()
}

func (db *DB) Close() error {
	return db.kv.Close()
}

func (db *DB) setup() (*DB, error) {
	// set up required buckets
	err := db.kv.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{[]byte("products"), []byte("indexes")} {
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

		return db.index(tx, p, id)
	})
	return id, err
}

func (db *DB) index(tx *bolt.Tx, p *onix.Product, id uint32) error {
	entries := db.indexFn(p)
	for _, e := range entries {
		bkt, err := tx.Bucket([]byte("indexes")).CreateBucketIfNotExists([]byte(e.Index))
		if err != nil {
			return err
		}

		term := []byte(strings.ToLower(e.Entry))
		hits := roaring.New()

		bo := bkt.Get(term)
		if bo != nil {
			if _, err := hits.ReadFrom(bytes.NewReader(bo)); err != nil {
				return err
			}
		}

		hits.Add(id)

		hitsb, err := hits.MarshalBinary()
		if err != nil {
			return err
		}

		if err := bkt.Put(term, hitsb); err != nil {
			return err
		}
	}
	return nil
}

type IndexEntry struct {
	Index string
	Entry string
}

type IndexFn func(*onix.Product) []IndexEntry

// Indexes returns the list of indicies in use.
func (db *DB) Indexes() (res []string) {
	db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("indexes"))
		c := bkt.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				res = append(res, string(k))
			}
		}
		return nil
	})
	return res
}

func (db *DB) Scan(index, start string, limit int) (res []string, err error) {
	err = db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("indexes")).Bucket([]byte(index))
		if bkt == nil {
			return fmt.Errorf("index not found: %s", index)
		}
		cur := bkt.Cursor()
		n := 0
		term := []byte(strings.ToLower(start))
		for k, _ := cur.Seek(term); k != nil; k, _ = cur.Next() {
			if n > limit {
				break
			}
			if !bytes.HasPrefix(k, term) {
				break
			}
			res = append(res, string(k))
			n++
		}
		return nil
	})
	return res, err
}

func (db *DB) Query(index, query string, limit int) (res []uint32, err error) {
	err = db.kv.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("indexes")).Bucket([]byte(index))
		if bkt == nil {
			return fmt.Errorf("index not found: %s", index)
		}
		bo := bkt.Get([]byte(strings.ToLower(query)))

		if bo == nil {
			return nil
		}

		hits := roaring.New()
		if _, err := hits.ReadFrom(bytes.NewReader(bo)); err != nil {
			return err
		}
		res = hits.ToArray()

		return nil
	})
	return res, err
}

// func (db *DB) DeleteIndex(index string) error

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
