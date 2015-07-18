package rdf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

// Exported datatypes
var (
	RDFLangString   = IRI{[]byte("\x00http://www.w3.org/1999/02/22-rdf-syntax-ns#langString")} // string 0x01
	RDFHTML         = IRI{[]byte("\x00http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML")}       // ?
	RDFXMLLiteral   = IRI{[]byte("\x00http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral")} // ?
	XSDString       = IRI{[]byte("\x00http://www.w3.org/2001/XMLSchema#string")}               // string 	0x02
	XSDBoolean      = IRI{[]byte("\x00http://www.w3.org/2001/XMLSchema#boolean")}              // boolean
	XSDDecimal      = IRI{[]byte("\x00http://www.w3.org/2001/XMLSchema#decimal")}              // big.Float
	XSDInteger      = IRI{[]byte("\x00http://www.w3.org/2001/XMLSchema#integer")}              // big.Int
	XSDLong         = IRI{[]byte("\x00http://www.w3.org/2001/XMLSchema#long")}                 // int64 	0x03
	XSDUnsignedLong = IRI{[]byte("\x00http://www.w3.org/2001/XMLSchema#unsignedLong")}         // uint64 	0x04
	// ...
	// TODO all RDF-compatible xsd datatypes
)

// Exported errors
var (
	ErrUndecodable    = errors.New("rdf: cannot decode bytes into Term")
	ErrInvalidIRI     = errors.New("rdf: invalid IRI: cannot be empty")
	ErrInvalidLiteral = errors.New("rdf: invalid Literal: cannot be empty")
)

// Term represents a RDF Term
type Term interface {
	// Bytes return the encoded byte representation of a Term.
	Bytes() []byte

	// String returns a string representation of a Term in N-Triples format.
	String() string

	// Value returns the typed value of a Term.
	Value() interface{}

	// Eq tests if two terms are equal.
	Eq(Term) bool
}

// DecodeTerm decodes a byte-serialzed term into a Term.
func DecodeTerm(b []byte) (Term, error) {
	if b == nil || len(b) < 2 {
		return nil, ErrUndecodable
	}
	switch b[0] {
	// IRI
	case 0x00:
		iri := make([]byte, len(b))
		copy(iri, b)
		return IRI{iri}, nil
	// rdf:langString
	case 0x01:
		if len(b) <= 2 || len(b) < int(b[1])+1 {
			return nil, ErrUndecodable
		}
		if int(b[1]) == 0 {
			// an empty language tag - consider it an xsd:String
			// TODO or return ErrUndecodable?
			val := make([]byte, len(b)-1)
			val[0] = 0x02
			copy(val[1:], b[2:])
			return Literal{val: val}, nil
		}
		val := make([]byte, len(b))
		copy(val, b)
		return Literal{val: val}, nil
	// xsd:String
	case 0x02:
		val := make([]byte, len(b))
		copy(val, b)
		return Literal{val: val}, nil
	// Other typed literals
	case 0xFF:
		ll := int(b[1]) + 2
		iri := make([]byte, ll-1)
		copy(iri[1:], b[2:ll])
		val := make([]byte, len(b))
		copy(val, b)
		return Literal{val: val}, nil
	default:
		panic("TODO")
	}
}

// IRI represents a IRI resource.
type IRI struct {
	val []byte
}

// NewIRI return a new IRI.
func NewIRI(iri string) (IRI, error) {
	if len(iri) == 0 {
		return IRI{}, ErrInvalidIRI
	}
	b := make([]byte, len(iri)+1)
	copy(b[1:], []byte(iri))
	return IRI{val: b}, nil
}

// Bytes returns the IRIs encoded byte representation.
func (i IRI) Bytes() []byte {
	return i.val
}

// String returns a N-Triples serialization of an IRI.
func (i IRI) String() string {
	return fmt.Sprintf("<%s>", string(i.val[1:]))
}

// Value returns the IRI as a string.
func (i IRI) Value() interface{} {
	return string(i.val[1:])
}

// Eq tests if IRI is equal to another Term.
func (i IRI) Eq(other Term) bool {
	return other != nil && i.String() == other.String()
}

// Literal represents a RDF Literal.
type Literal struct {
	val   []byte
	typed interface{}
}

// NewLiteral returns a new Literal, with a datatype infered from the type of the value,
// or an error if the Literal is invalid or of unknown Go type.
func NewLiteral(val interface{}) (Literal, error) {
	switch t := val.(type) {
	case string:
		if len(t) == 0 {
			return Literal{}, ErrInvalidLiteral
		}
		val := make([]byte, len(t)+1)
		copy(val[1:], []byte(t))
		val[0] = 0x02
		return Literal{val: val}, nil
	case int:
		b := make([]byte, 1+binary.MaxVarintLen64)
		b[0] = 0x03
		l := binary.PutVarint(b[1:], int64(t))
		return Literal{val: b[0 : l+1]}, nil
	case uint:
		b := make([]byte, 1+binary.MaxVarintLen64)
		b[0] = 0x04
		l := binary.PutUvarint(b[1:], uint64(t))
		return Literal{val: b[0 : l+1]}, nil
	}
	panic("NewLiteral: TODO")
}

// NewLangLiteral returns a new Literal of type RDFLangString with given language tag.
// If lang is empty, the Literal will be typed as xsd:String
func NewLangLiteral(val string, lang string) (Literal, error) {
	if len(val) == 0 {
		return Literal{}, ErrInvalidLiteral
	}
	if len(lang) == 0 {
		b := make([]byte, len(val)+1)
		b[1] = 0x02
		copy(b, []byte(val))
		return Literal{val: b}, nil
	}
	b := make([]byte, len(val)+len(lang)+2)
	b[0] = 0x01
	ll := len(lang)
	b[1] = uint8(ll)
	copy(b[2:], []byte(lang))
	copy(b[2+ll:], []byte(val))
	return Literal{val: b}, nil
}

// NewTypedLiteral returns a new Literal with the given datatype.
func NewTypedLiteral(val string, typ IRI) (Literal, error) {
	if len(val) == 0 {
		return Literal{}, ErrInvalidLiteral
	}
	if typ.Eq(XSDString) {
		b := make([]byte, len(val)+1)
		b[0] = 0x02
		copy(b[1:], []byte(val))
		return Literal{val: b}, nil
	} // elsif other xsd types
	b := make([]byte, len(val)+len(typ.val)+1)
	b[0] = 0xFF
	b[1] = uint8(len(typ.val) - 1)
	copy(b[2:], typ.val[1:])
	copy(b[b[1]+2:], []byte(val))
	return Literal{val: b}, nil
}

// DataType returns the DataType IRI of the Literal.
func (l Literal) DataType() IRI {
	switch l.val[0] {
	case 0x01:
		return RDFLangString
	case 0x02:
		return XSDString
	case 0x03:
		return XSDLong
	case 0x04:
		return XSDUnsignedLong
	case 0xFF:
		b := make([]byte, l.val[1]+1)
		copy(b[1:], l.val[2:l.val[1]+2])
		return IRI{b}
	}
	panic("TODO Literal.DataType() not string, langstring or other")
}

// Lang returns the language tag of a literal, or an empty string
// if it is not of type RDFLangString.
func (l Literal) Lang() string {
	if l.val[0] == 0x01 {
		return string(l.val[2 : l.val[1]+2])
	}
	return ""
}

// String returns a N-Triples serialization of a Literal.
func (l Literal) String() string {
	switch l.val[0] {
	case 0x01:
		lang := l.Lang()
		return fmt.Sprintf("\"%s\"@%s", string(l.val[2+len(lang):]), lang)
	case 0x02:
		return fmt.Sprintf("\"%s\"", string(l.val[1:]))
	case 0x03:
		return fmt.Sprintf("\"%d\"^^%s", l.Value(), l.DataType().String())
	case 0x04:
		return fmt.Sprintf("\"%d\"^^%s", l.Value(), l.DataType().String())
	case 0xFF:
		dt := l.DataType().String()
		return fmt.Sprintf("\"%s\"^^%s", string(l.val[len(dt):]), dt)
	}

	panic("TODO Literal.String() not langstring, string, other")
}

// Value returns the Literal as a Go-typed value
func (l Literal) Value() interface{} {
	if l.typed == nil {
		switch l.val[0] {
		case 0x01:
			l.typed = string(l.val[2+len(l.Lang()):])
		case 0x02:
			l.typed = string(l.val[1:])
		case 0x03:
			i, _ := binary.Varint(l.val[1:])
			l.typed = i
		case 0x04:
			i, _ := binary.Uvarint(l.val[1:])
			l.typed = i
		case 0xFF:
			l.typed = string(l.val[len(l.DataType().String()):])
		default:
			panic("TODO Literal.Value()")
		}
	}
	return l.typed
}

// Eq tests if the Literal is equal to another Term.
func (l Literal) Eq(other Term) bool {
	return other != nil && bytes.Equal(l.Bytes(), other.Bytes())
}

// Bytes return the Literal's encoded byte representation.
func (l Literal) Bytes() []byte {
	return l.val
}

// AsGoType returns the literal in a corresponding Go type.
//func (l Literal) AsGoType() interface{}

// Triple represents a RDF statement with subject, predicate and object
type Triple struct {
	subj IRI
	pred IRI
	obj  Term
}

// NewTriple returns a triple with the given subject, predicate and object.
func NewTriple(s IRI, p IRI, o Term) Triple {
	return Triple{
		subj: s,
		pred: p,
		obj:  o,
	}
}

// Subject returns the subject of a Triple.
func (t Triple) Subject() IRI {
	return t.subj
}

// Predicate returns the predicate of a Triple.
func (t Triple) Predicate() IRI {
	return t.pred
}

// Object returns the object of a Triple.
func (t Triple) Object() Term {
	return t.obj
}

// String returns an N-triples serialization of the Triple.
func (t Triple) String() string {
	return fmt.Sprintf("%s %s %s .", t.subj.String(), t.pred.String(), t.obj.String())
}

// Eq tests if two triples are equal.
func (t Triple) Eq(other Triple) bool {
	return t.Subject().Eq(other.Subject()) &&
		t.Predicate().Eq(other.Predicate()) &&
		t.Object().Eq(other.Object())
}

// Graph is a collection of triples.
type Graph []Triple

// Len satisfies the Sort interface for Graph.
func (g Graph) Len() int { return len(g) }

// Swap satisfies the Sort interface for Graph.
func (g Graph) Swap(i, j int) { g[i], g[j] = g[j], g[i] }

// Less satisfies the Sort interface for Graph.
func (g Graph) Less(i, j int) bool { return g[i].String() < g[j].String() }

// Eq tests for equality between graphs, meaning that they contain
// the same triples, and no graph has triples not in the other graph.
func (g Graph) Eq(other Graph) bool {
	if len(g) != len(other) {
		return false
	}
	sort.Sort(g)
	sort.Sort(other)
	for i, tr := range g {
		if !tr.Eq(other[i]) {
			return false
		}
	}
	return true
}

//func (g Graph) Dump(w io.Writer) error {}
//func (g Graph) String() string {}

//func Load(r io.Reader) (Graph, error) {}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
