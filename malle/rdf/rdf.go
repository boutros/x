package rdf

import (
	"bytes"
	"fmt"
)

var (
	rdfLangString = NewIRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#langString")
)

type Term interface {
	Bytes() []byte
	String() string // in N-triples format.
}

// TODO move this logic to db?
func ReadTerm(b []byte) (Term, error) {
	if b[0] == '0' {
		// IRI
		return IRI{val: b[1:]}, nil
	} else {
		// Literal
	}
	return nil, fmt.Errorf("cannot parse Term from %q", b)
}

func TermsEq(a, b Term) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

type IRI struct {
	val []byte
}

func NewIRI(s string) IRI {
	return IRI{val: []byte(s)}
}

func (u IRI) Bytes() []byte {
	return u.val
}

func (u IRI) String() string {
	return fmt.Sprintf("<%s>", string(u.val))
}

type Literal struct {
	langLength uint8
	val        []byte
	dataType   IRI
}

func NewLangLiteral(val string, lang string) Literal {
	b := make([]byte, len(val)+len(lang))
	copy([]byte(lang), b)
	copy([]byte(val), b[len(lang):])
	return Literal{
		langLength: uint8(len(lang)),
		val:        b,
		dataType:   rdfLangString,
	}
}

//func NewLiteral(val interface{}) Literal {}

func NewTypedLiteral(val []byte, typ IRI) Literal {
	return Literal{
		val:      val,
		dataType: typ,
	}
}

func (l Literal) DataType() IRI {
	return l.dataType
}

func (l Literal) Lang() string {
	return string(l.val[:l.langLength])
}

func (l Literal) String() string {
	if l.langLength > 0 {
		return fmt.Sprintf("\"%s\"@%s",
			string(l.val[l.langLength:]),
			string(l.val[:l.langLength]))
	}
	return fmt.Sprintf("\"%s\"^^%s", string(l.val), l.dataType.String())
}

//func (l Literal) AsGoType() interface{}

type Triple struct {
	Subj IRI
	Pred IRI
	Obj  Term
}
