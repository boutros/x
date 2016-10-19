package test

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/boutros/x/onixdb"
	"github.com/knakk/kbp/onix"
)

const records = `
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
      <SubjectHeadingText>Subject A</SubjectHeadingText>
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
        <TitleWithoutPrefix textcase="01">Book B</TitleWithoutPrefix>
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
      <KeyNames>Hansen</KeyNames>
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
      <SubjectHeadingText>Subject C</SubjectHeadingText>
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
`

func TestAll(t *testing.T) {
	f := tempfile()
	defer os.Remove(f)
	db, err := onixdb.Open(f)
	if err != nil {
		log.Fatal(err)
	}

	defer checked(t, db.Close)

	type Products struct {
		Product []*onix.Product
	}
	var products Products
	if err := xml.Unmarshal([]byte(records), &products); err != nil {
		t.Fatal(err)
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
			t.Error("stored record not equal. Got:\n%v\nWant:\n%v", p, products.Product[i])
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
