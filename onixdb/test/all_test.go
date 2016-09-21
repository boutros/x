package test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/boutros/x/onixdb"
)

func TestAll(t *testing.T) {
	f := tempfile()
	defer os.Remove(f)
	db, err := onixdb.Open(f)
	if err != nil {
		log.Fatal(err)
	}

	defer checked(t, db.Close)
}

func checked(t *testing.T, f func() error) {
	if err := f(); err != nil {
		t.Error(err)
	}
}

// tempfile returns a temporary file path.
func tempfile() string {
	f, _ := ioutil.TempFile("", "onixdb-")
	f.Close()
	os.Remove(f.Name())
	return f.Name()
}
