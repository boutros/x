package onixdb

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/knakk/kbp/onix"
)

func BenchmarkMarshalXML(b *testing.B) {
	xmlbytes, err := ioutil.ReadFile(filepath.Join("testdata", "sample.xml"))
	if err != nil {
		b.Fatal(err)
	}
	var p onix.Product
	if err := xml.Unmarshal(xmlbytes, &p); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	var size int
	for n := 0; n < b.N; n++ {
		out, err := xml.Marshal(p)
		if err != nil {
			b.Fatal(err)
		}
		size = len(out)
	}
	b.Logf("size: %d", size)
}

func BenchmarkMarshalXMLGzipped(b *testing.B) {
	xmlbytes, err := ioutil.ReadFile(filepath.Join("testdata", "sample.xml"))
	if err != nil {
		b.Fatal(err)
	}
	var p onix.Product
	if err := xml.Unmarshal(xmlbytes, &p); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	var size int
	var out bytes.Buffer
	for n := 0; n < b.N; n++ {
		out.Reset()
		w := gzip.NewWriter(&out)
		enc := xml.NewEncoder(w)
		if err := enc.Encode(p); err != nil {
			b.Fatal(err)
		}
		w.Close()
		size = len(out.Bytes())
	}
	b.Logf("size: %d", size)
}

func BenchmarkMarshalGob(b *testing.B) {
	xmlbytes, err := ioutil.ReadFile(filepath.Join("testdata", "sample.xml"))
	if err != nil {
		b.Fatal(err)
	}
	var p onix.Product
	if err := xml.Unmarshal(xmlbytes, &p); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	var size int
	var out bytes.Buffer
	for n := 0; n < b.N; n++ {
		out.Reset()
		if err := gob.NewEncoder(&out).Encode(p); err != nil {
			b.Fatal(err)
		}
		size = len(out.Bytes())
	}
	b.Logf("size: %d", size)
}

func BenchmarkMarshalGob2(b *testing.B) {
	xmlbytes, err := ioutil.ReadFile(filepath.Join("testdata", "sample.xml"))
	if err != nil {
		b.Fatal(err)
	}
	var p onix.Product
	if err := xml.Unmarshal(xmlbytes, &p); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	var size int
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	for n := 0; n < b.N; n++ {
		out.Reset()
		if err := enc.Encode(p); err != nil {
			b.Fatal(err)
		}
		size = len(out.Bytes())
	}
	b.Logf("size: %d", size)
}

func TestPrimedCoded(t *testing.T) {
	xmlbytes, err := ioutil.ReadFile(filepath.Join("testdata", "sample.xml"))
	if err != nil {
		t.Fatal(err)
	}
	var p onix.Product
	if err := xml.Unmarshal(xmlbytes, &p); err != nil {
		t.Fatal(err)
	}

	codec, err := newPrimedCodec(&onix.Product{})
	if err != nil {
		t.Fatal(err)
	}
	dec := codec.NewMarshaler()
	out, err := dec.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(out))
	enc := codec.NewUnmarshaler()
	got, err := enc.Unmarshal(out)
	if err != nil {
		t.Fatal(err)
	}

	xml, err := xml.MarshalIndent(got, "  ", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(xml))
}
