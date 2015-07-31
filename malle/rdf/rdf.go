package rdf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// Exported datatypes
var (
	RDFLangString   = IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#langString") // string 0x01
	RDFHTML         = IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML")       // ?
	RDFXMLLiteral   = IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral") // ?
	XSDString       = IRI("http://www.w3.org/2001/XMLSchema#string")               // string 	0x02
	XSDBoolean      = IRI("http://www.w3.org/2001/XMLSchema#boolean")              // boolean
	XSDDecimal      = IRI("http://www.w3.org/2001/XMLSchema#decimal")              // big.Float
	XSDInteger      = IRI("http://www.w3.org/2001/XMLSchema#integer")              // big.Int
	XSDLong         = IRI("http://www.w3.org/2001/XMLSchema#long")                 // int64 	0x03
	XSDUnsignedLong = IRI("http://www.w3.org/2001/XMLSchema#unsignedLong")         // uint64 	0x04
	// ...
	// TODO all RDF-compatible xsd datatypes
)

// Exported errors
var (
	ErrUndecodable        = errors.New("rdf: cannot decode bytes into Term")
	ErrInvalidIRI         = errors.New("rdf: invalid IRI: cannot be empty")
	ErrInvalidLiteral     = errors.New("rdf: invalid Literal: cannot be empty")
	ErrInvalidLangLiteral = errors.New("rdf: invalid rdf:LangLiteral: language cannot be empty")
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
	// TODO remove when structs are comparable with ==
	Eq(Term) bool
}

// DecodeTerm decodes a byte-serialzed term into a Term.
func DecodeTerm(b []byte) (Term, error) {
	if b == nil || len(b) < 2 {
		return nil, ErrUndecodable
	}
	switch b[0] {
	case 0x00: // IRI
		return IRI(string(b[1:])), nil
	case 0x01: // rdf:langString
		if len(b) <= 2 || len(b) < int(b[1])+1 {
			return nil, ErrUndecodable
		}
		if int(b[1]) == 0 {
			// an empty language tag - consider it an xsd:String
			// TODO or return ErrUndecodable?
			return Literal{val: string(b[2:]), datatype: XSDString}, nil
		}
		ll := 2 + b[1]
		return Literal{val: string(b[ll:]), lang: string(b[2:ll]), datatype: RDFLangString}, nil
	case 0x02: // xsd:String
		return Literal{val: string(b[1:]), datatype: XSDString}, nil
	case 0xFF: // Other typed literals
		ll := int(b[1])
		return Literal{val: string(b[2+ll:]), datatype: IRI(string(b[2 : 2+ll]))}, nil
	default:
		panic("TODO DecodeTerm")
	}
}

// IRI represents a IRI resource.
type IRI string

// NewIRI return a new IRI. It cannot be empty.
func NewIRI(iri string) (IRI, error) {
	if len(iri) == 0 {
		return IRI(""), ErrInvalidIRI
	}
	return IRI(iri), nil
}

// Bytes returns the IRIs encoded byte representation.
func (i IRI) Bytes() []byte {
	b := make([]byte, len(i)+1)
	// b[0] = 0x00 - allready zero value of []byte
	copy(b[1:], []byte(i))
	return b
}

// String returns a N-Triples serialization of an IRI.
func (i IRI) String() string {
	return "<" + string(i) + ">"
}

// Value returns the IRI as a string.
func (i IRI) Value() interface{} {
	return i
}

// Eq tests if IRI is equal to another Term.
func (i IRI) Eq(other Term) bool {
	return other != nil && i.String() == other.String()
}

// Literal represents a RDF Literal.
// TODO consider split into LangLiteral and TypedLiteral
type Literal struct {
	val      string
	lang     string
	datatype IRI
}

// NewLiteral returns a new Literal, with a datatype infered from the type of the value,
// or an error if the Literal is invalid or of unknown Go type.
func NewLiteral(val interface{}) (Literal, error) {
	switch t := val.(type) {
	case string:
		if len(t) == 0 {
			return Literal{}, ErrInvalidLiteral
		}
		return Literal{val: t, datatype: XSDString}, nil
	case int:
		return Literal{val: strconv.Itoa(t), datatype: XSDLong}, nil
	case uint:
		return Literal{val: strconv.FormatUint(uint64(t), 10), datatype: XSDUnsignedLong}, nil
	default:
		panic("NewLiteral: TODO")
	}
}

// NewLangLiteral returns a new Literal of type RDFLangString with given language tag.
// If lang is empty, the Literal will be typed as xsd:String
func NewLangLiteral(val string, lang string) (Literal, error) {
	if len(val) == 0 {
		return Literal{}, ErrInvalidLiteral
	}
	if len(lang) == 0 {
		return Literal{}, ErrInvalidLangLiteral
	}
	return Literal{val: val, lang: lang, datatype: RDFLangString}, nil
}

// NewTypedLiteral returns a new Literal with the given datatype.
func NewTypedLiteral(val string, typ IRI) (Literal, error) {
	if len(val) == 0 {
		return Literal{}, ErrInvalidLiteral
	}
	return Literal{val: val, datatype: typ}, nil
}

// DataType returns the DataType IRI of the Literal.
func (l Literal) DataType() IRI {
	return l.datatype
}

// Lang returns the language tag of a literal, or an empty string
// if it is not of type RDFLangString.
func (l Literal) Lang() string {
	return l.lang
}

// String returns a N-Triples serialization of a Literal.
func (l Literal) String() string {
	switch l.datatype {
	case RDFLangString:
		return fmt.Sprintf("\"%s\"@%s", l.val, l.lang)
	case XSDString:
		return fmt.Sprintf("\"%s\"", l.val)
	case XSDLong, XSDUnsignedLong:
		return fmt.Sprintf("\"%s\"^^%s", l.val, l.datatype)
	default:
		return fmt.Sprintf("\"%s\"^^%s", l.val, l.datatype)
	}
}

// Value returns the Literal as a Go-typed value
func (l Literal) Value() interface{} {
	switch l.datatype {
	case RDFLangString, XSDString:
		return l.val
	case XSDLong:
		i, _ := strconv.ParseInt(l.val, 10, 0)
		return i
	case XSDUnsignedLong:
		i, _ := strconv.ParseUint(l.val, 10, 0)
		return i
	default:
		return l.val
	}
}

// Eq tests if the Literal is equal to another Term.
func (l Literal) Eq(other Term) bool {
	return other != nil && l.String() == other.String()
}

// Bytes return the Literal's encoded byte representation.
func (l Literal) Bytes() []byte {
	switch l.datatype {
	case XSDString:
		b := make([]byte, len(l.val)+1)
		copy(b[1:], []byte(l.val))
		b[0] = 0x02
		return b
	case RDFLangString:
		ll := len(l.lang)
		b := make([]byte, len(l.val)+ll+2)
		b[0] = 0x01
		b[1] = uint8(ll)
		copy(b[2:], []byte(l.lang))
		copy(b[2+ll:], []byte(l.val))
		return b
	case XSDLong:
		v, _ := strconv.Atoi(l.val)
		b := make([]byte, 1+binary.MaxVarintLen64)
		b[0] = 0x03
		l := binary.PutVarint(b[1:], int64(v))
		return b[0 : l+1]
	case XSDUnsignedLong:
		v, _ := strconv.Atoi(l.val)
		b := make([]byte, 1+binary.MaxVarintLen64)
		b[0] = 0x04
		l := binary.PutUvarint(b[1:], uint64(v))
		return b[0 : l+1]
	default:
		b := make([]byte, len(l.val)+len(l.datatype)+2)
		b[0] = 0xFF
		b[1] = uint8(len(l.datatype))
		copy(b[2:], []byte(l.datatype))
		copy(b[b[1]+2:], []byte(l.val))
		return b
	}
}

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
type Graph map[IRI]map[IRI]Terms

// NewGraph returns a new graph.
func NewGraph() Graph {
	return make(map[IRI]map[IRI]Terms)
}

// Size returns the number of triples in the graph.
func (g Graph) Size() int {
	n := 0
	for _, props := range g {
		for _, vals := range props {
			n += len(vals)
		}
	}
	return n
}

// Eq tests for equality between graphs, meaning that they contain
// the same triples, and no graph has triples not in the other graph.
func (g Graph) Eq(other Graph) bool {
	if len(g) != len(other) {
		return false
	}
	for subj, props := range g {
		if _, ok := other[subj]; !ok {
			return false
		}
		for pred, terms := range props {
			if _, ok := other[subj][pred]; !ok {
				return false
			}
			if !eqTerms(terms, other[subj][pred]) {
				return false
			}
		}
	}
	return true
}

// eqTerms checks if two Terms contains the same triples.
func eqTerms(a, b Terms) bool {
	sort.Sort(a)
	sort.Sort(b)
	for i, t := range a {
		if !t.Eq(b[i]) {
			return false
		}
	}
	return true
}

// Terms is a slice of []Term. (Nessecary to make it sortable)
type Terms []Term

// Len satisfies the Sort interface for Terms.
func (t Terms) Len() int { return len(t) }

// Swap satisfies the Sort interface for Terms.
func (t Terms) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// Less satisfies the Sort interface for Terms.
func (t Terms) Less(i, j int) bool { return t[i].String() < t[j].String() }

// Add adds a triple to the graph
func (g Graph) Add(t Triple) Graph {
	if _, ok := g[t.subj]; ok {
		// subject exists
		if terms, ok := g[t.subj][t.pred]; ok {
			// predicate exists
			for _, term := range terms {
				if term.Eq(t.obj) {
					// triple allready in graph
					return g
				}
			}
			// add object
			g[t.subj][t.pred] = append(g[t.subj][t.pred], t.obj)
		} else {
			// new predicate for subject
			g[t.subj][t.pred] = make(Terms, 0, 1)
			// add object
			g[t.subj][t.pred] = append(g[t.subj][t.pred], t.obj)
		}
	} else {
		// new subject
		g[t.subj] = make(map[IRI]Terms)
		// add predicate
		g[t.subj][t.pred] = make(Terms, 0, 1)
		// add object
		g[t.subj][t.pred] = append(g[t.subj][t.pred], t.obj)
	}
	return g
}

//func (g Graph) Dump(w io.Writer) error {}
//func (g Graph) String() string {}

// Load reads N-Triples from a stream until EOF and returns the parsed
// triples as a Graph. Any triples with blank nodes or syntactic errors
// will be ignored.
func Load(r io.Reader) Graph {
	d := NewNTDecoder(r)
	g := NewGraph()
	for tr, err := d.Decode(); err != io.EOF; tr, err = d.Decode() {
		if err == nil {
			g.Add(tr)
		}
	}
	return g
}
