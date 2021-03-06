package test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/boutros/x/onixdb"
	"github.com/knakk/kbp/onix"
)

var records = []byte(`
<Products>
<Product>
  <RecordReference>id.0</RecordReference>
  <NotificationType>03</NotificationType>
  <RecordSourceType>04</RecordSourceType>
  <ProductIdentifier>
    <ProductIDType>03</ProductIDType>
    <IDValue>9780000000111</IDValue>
  </ProductIdentifier>
  <DescriptiveDetail>
    <TitleDetail>
      <TitleType>01</TitleType>
      <TitleElement>
        <TitleElementLevel>01</TitleElementLevel>
        <NoPrefix></NoPrefix>
        <TitleWithoutPrefix textcase="01">Book A</TitleWithoutPrefix>
      </TitleElement>
    </TitleDetail>
    <Contributor>
      <SequenceNumber>1</SequenceNumber>
      <ContributorRole>A01</ContributorRole>
      <NameIdentifier>
        <NameIDType>16</NameIDType>
        <IDValue>0000000001</IDValue>
      </NameIdentifier>
      <NamesBeforeKey>Ole</NamesBeforeKey>
      <KeyNames>Jensen</KeyNames>
    </Contributor>
    <NoEdition></NoEdition>
    <Subject>
      <SubjectSchemeIdentifier>20</SubjectSchemeIdentifier>
      <SubjectHeadingText>Subject Ape</SubjectHeadingText>
    </Subject>
  </DescriptiveDetail>
  <PublishingDetail>
    <Publisher>
      <PublishingRole>01</PublishingRole>
      <PublisherName>Knakks forlag</PublisherName>
    </Publisher>
    <CityOfPublication>Oslo</CityOfPublication>
    <CountryOfPublication>NO</CountryOfPublication>
  </PublishingDetail>
</Product>
<Product>
  <RecordReference>id.1</RecordReference>
  <NotificationType>03</NotificationType>
  <RecordSourceType>04</RecordSourceType>
  <ProductIdentifier>
    <ProductIDType>03</ProductIDType>
    <IDValue>9780000000222</IDValue>
  </ProductIdentifier>
  <DescriptiveDetail>
    <TitleDetail>
      <TitleType>01</TitleType>
      <TitleElement>
        <TitleElementLevel>01</TitleElementLevel>
        <NoPrefix></NoPrefix>
        <TitleWithoutPrefix textcase="01">Book Babel</TitleWithoutPrefix>
      </TitleElement>
    </TitleDetail>
    <Contributor>
      <SequenceNumber>1</SequenceNumber>
      <ContributorRole>A01</ContributorRole>
      <NameIdentifier>
        <NameIDType>16</NameIDType>
        <IDValue>0000000001</IDValue>
      </NameIdentifier>
      <NamesBeforeKey>Kari</NamesBeforeKey>
      <KeyNames>Jensen</KeyNames>
    </Contributor>
    <NoEdition></NoEdition>
    <Subject>
      <SubjectSchemeIdentifier>20</SubjectSchemeIdentifier>
      <SubjectHeadingText>Subject B</SubjectHeadingText>
    </Subject>
  </DescriptiveDetail>
  <PublishingDetail>
    <Publisher>
      <PublishingRole>01</PublishingRole>
      <PublisherName>Knakks forlag</PublisherName>
    </Publisher>
    <CityOfPublication>Oslo</CityOfPublication>
    <CountryOfPublication>NO</CountryOfPublication>
  </PublishingDetail>
</Product>
<Product>
  <RecordReference>id.2</RecordReference>
  <NotificationType>03</NotificationType>
  <RecordSourceType>04</RecordSourceType>
  <ProductIdentifier>
    <ProductIDType>03</ProductIDType>
    <IDValue>9780000000333</IDValue>
  </ProductIdentifier>
  <DescriptiveDetail>
    <TitleDetail>
      <TitleType>01</TitleType>
      <TitleElement>
        <TitleElementLevel>01</TitleElementLevel>
        <NoPrefix></NoPrefix>
        <TitleWithoutPrefix textcase="01">Book C</TitleWithoutPrefix>
      </TitleElement>
    </TitleDetail>
    <Contributor>
      <SequenceNumber>1</SequenceNumber>
      <ContributorRole>A01</ContributorRole>
      <NameIdentifier>
        <NameIDType>16</NameIDType>
        <IDValue>0000000001</IDValue>
      </NameIdentifier>
      <NamesBeforeKey>Jens</NamesBeforeKey>
      <KeyNames>Olsen</KeyNames>
    </Contributor>
    <NoEdition></NoEdition>
    <Subject>
      <SubjectSchemeIdentifier>20</SubjectSchemeIdentifier>
      <SubjectHeadingText>Subject Api</SubjectHeadingText>
    </Subject>
  </DescriptiveDetail>
  <PublishingDetail>
    <Publisher>
      <PublishingRole>01</PublishingRole>
      <PublisherName>Knakks forlag</PublisherName>
    </Publisher>
    <CityOfPublication>Oslo</CityOfPublication>
    <CountryOfPublication>NO</CountryOfPublication>
  </PublishingDetail>
</Product>
</Products>
`)

func indexFn(p *onix.Product) (res []onixdb.IndexEntry) {
	// ISBN
	for _, id := range p.ProductIdentifier {
		if id.ProductIDType.Value == "03" {
			res = append(res, onixdb.IndexEntry{
				Index: "isbn",
				Entry: id.IDValue.Value,
			})
		}
	}

	// Title
	for _, td := range p.DescriptiveDetail.TitleDetail {
		for _, te := range td.TitleElement {
			for _, s := range strings.Split(te.TitleWithoutPrefix.Value, " ") {
				res = append(res, onixdb.IndexEntry{
					Index: "title",
					Entry: s,
				})
			}
		}
	}

	// Contributor
	for _, c := range p.DescriptiveDetail.Contributor {
		res = append(res, onixdb.IndexEntry{
			Index: "author",
			Entry: fmt.Sprintf("%s, %s", c.KeyNames.Value, c.NamesBeforeKey.Value),
		})
		res = append(res, onixdb.IndexEntry{
			Index: "author",
			Entry: c.KeyNames.Value,
		})
		res = append(res, onixdb.IndexEntry{
			Index: "author",
			Entry: c.NamesBeforeKey.Value,
		})
	}

	// Subject
	for _, s := range p.DescriptiveDetail.Subject {
		for _, st := range s.SubjectHeadingText {
			res = append(res, onixdb.IndexEntry{
				Index: "subject",
				Entry: st.Value,
			})

		}
	}

	return res
}

func TestAll(t *testing.T) {
	f := tempfile()
	defer os.Remove(f)
	db, err := onixdb.Open(f, indexFn)
	if err != nil {
		log.Fatal(err)
	}

	defer checked(t, db.Close)

	type Products struct {
		Product []*onix.Product
	}
	var products Products
	if err := xml.Unmarshal(records, &products); err != nil {
		t.Fatal(err)
	}

	// Verify that there are no indexes in an empty db
	if len(db.Indexes()) != 0 {
		t.Fatalf("indexes not empty: %v", db.Indexes())
	}
	// Verify that records can be stored and given an ID.
	ids := make([]uint32, len(products.Product))
	for i, p := range products.Product {
		var err error
		ids[i], err = db.Store(p)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify that records can be retrieved using ID, and are equal to
	// the records we put in.
	for n, i := range ids {
		p, err := db.Get(i)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p, products.Product[n]) {
			t.Errorf("stored record not equal. Got:\n%v\nWant:\n%v", p, products.Product[i])
		}
	}

	// Verify that we have indexes used in indexFn
	want := []string{"author", "isbn", "subject", "title"}
	got := db.Indexes()
	sort.Sort(sort.StringSlice(got))
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("db.Indicies() => %v; want %v", got, want)
	}

	searchTests := []struct {
		idx      string
		q        string
		scans    []string
		products []uint32
	}{
		{
			idx:      "isbn",
			q:        "9780000000",
			scans:    []string{"9780000000111", "9780000000222", "9780000000333"},
			products: nil,
		},
		{
			idx:      "isbn",
			q:        "9780000000111",
			scans:    []string{"9780000000111"},
			products: []uint32{ids[0]},
		},
		{
			idx:      "title",
			q:        "babel",
			scans:    []string{"babel"},
			products: []uint32{ids[1]},
		},
		{
			idx:      "author",
			q:        "jens",
			scans:    []string{"jens", "jensen", "jensen, kari", "jensen, ole"},
			products: []uint32{ids[2]},
		},
		{
			idx:      "author",
			q:        "jensen",
			scans:    []string{"jensen", "jensen, kari", "jensen, ole"},
			products: []uint32{ids[0], ids[1]},
		},
		{
			idx:      "subject",
			q:        "Subject a",
			scans:    []string{"subject ape", "subject api"},
			products: nil,
		},
	}

	for _, test := range searchTests {
		scans, err := db.Scan(test.idx, test.q, 10)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(scans, test.scans) {
			t.Errorf("db.Scan(%s, %s, 10) => %v; want %v", test.idx, test.q, scans, test.scans)
		}

		ids, err := db.Query(test.idx, test.q, 10)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(ids, test.products) {
			t.Errorf("db.Query(%s, %s, 10) => %v; want %v", test.idx, test.q, ids, test.products)
		}
	}

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
