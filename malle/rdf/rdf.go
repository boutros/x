package rdf

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
)

// Exported datatypes
var (
	RDFLangString   = IRI{"http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"}
	RDFHTML         = IRI{"http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"}
	RDFXMLLiteral   = IRI{"http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral"}
	XSDString       = IRI{"http://www.w3.org/2001/XMLSchema#string"}       // string
	XSDBoolean      = IRI{"http://www.w3.org/2001/XMLSchema#boolean"}      // boolean
	XSDDecimal      = IRI{"http://www.w3.org/2001/XMLSchema#decimal"}      // big.Float
	XSDInteger      = IRI{"http://www.w3.org/2001/XMLSchema#integer"}      // big.Int
	XSDLong         = IRI{"http://www.w3.org/2001/XMLSchema#long"}         // int64
	XSDUnsignedLong = IRI{"http://www.w3.org/2001/XMLSchema#unsignedLong"} // uint64
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
	// Encode returns a byte representation of a term.
	Encode() []byte

	// String returns a string representation of a Term in N-Triples format.
	String() string
}

// DecodeTerm decodes a byte-serialzed term into a Term.
func DecodeTerm(b []byte) (Term, error) {
	if b == nil || len(b) < 2 {
		return nil, ErrUndecodable
	}
	switch b[0] {
	// IRI
	case 0x00:
		return IRI{string(b[1:])}, nil
	// rdf:langString
	case 0x01:
		if len(b) <= 2 || len(b) < int(b[1])+1 {
			return nil, ErrUndecodable
		}
		if int(b[1]) == 0 {
			// an empty language tag - consider it an xsd:String
			return Literal{
				val:      string(b[2:]),
				dataType: XSDString,
			}, nil
		}
		return Literal{
			val:      string(b[(2 + int(b[1])):]),
			lang:     string(b[2:(2 + int(b[1]))]),
			dataType: RDFLangString,
		}, nil
	// xsd:String
	case 0x02:
		return Literal{
			val:      string(b[1:]),
			dataType: XSDString,
		}, nil
	// Other typed literals
	case 0xFF:
		ll := int(b[1]) + 2
		return Literal{
			val:      string(b[ll:]),
			dataType: IRI{string(b[2:ll])},
		}, nil
	default:
		panic("TODO")
	}
}

// TermsEq tests for term equality.
func TermsEq(a, b Term) bool {
	return a != nil && b != nil && a.String() == b.String()
}

// IRI represents a IRI resource.
type IRI struct {
	val string
}

// NewIRI return a new IRI.
func NewIRI(iri string) (IRI, error) {
	if len(iri) == 0 {
		return IRI{}, ErrInvalidIRI
	}
	return IRI{val: iri}, nil
}

// Encode encodes an IRI.
func (i IRI) Encode() []byte {
	b := make([]byte, len(i.val)+1)
	b[0] = 0x00
	copy(b[1:], []byte(i.val))
	return b
}

// String returns a N-Triples serialization of an IRI.
func (i IRI) String() string {
	return fmt.Sprintf("<%s>", i.val)
}

// Literal represents a RDF Literal.
type Literal struct {
	val      string
	lang     string
	dataType IRI
}

// NewLiteral returns a new Literal, with a datatype infered from the type of the value,
// or an error if the Literal is invalid or of unknown Go type.
func NewLiteral(val interface{}) (Literal, error) {
	switch t := val.(type) {
	case string:
		if len(t) == 0 {
			return Literal{}, ErrInvalidLiteral
		}
		return Literal{val: t, dataType: XSDString}, nil
	case int:
		return Literal{val: strconv.Itoa(t), dataType: XSDLong}, nil
	case uint:
		return Literal{val: strconv.FormatUint(uint64(t), 10), dataType: XSDUnsignedLong}, nil
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
		return Literal{
			val:      val,
			dataType: XSDString,
		}, nil
	}
	return Literal{
		val:      val,
		lang:     lang,
		dataType: RDFLangString,
	}, nil
}

// NewTypedLiteral returns a new Literal with the given datatype.
func NewTypedLiteral(val string, typ IRI) (Literal, error) {
	if len(val) == 0 {
		return Literal{}, ErrInvalidLiteral
	}
	return Literal{
		val:      val,
		dataType: typ,
	}, nil
}

// DataType returns the DataType IRI of the Literal.
func (l Literal) DataType() IRI {
	return l.dataType
}

// Lang returns the language tag of a literal, or an empty string
// if it is not of type RDFLangString.
func (l Literal) Lang() string {
	return l.lang
}

// String returns a N-Triples serialization of a Literal.
func (l Literal) String() string {
	if l.lang != "" {
		return fmt.Sprintf("\"%s\"@%s", l.val, l.lang)
	} else if TermsEq(l.DataType(), XSDString) {
		return fmt.Sprintf("\"%s\"", l.val)
	}
	return fmt.Sprintf("\"%s\"^^%s", l.val, l.dataType.String())
}

// Encode encodes a Literal.
func (l Literal) Encode() []byte {
	switch l.DataType() {
	case RDFLangString:
		b := make([]byte, len(l.val)+len(l.lang)+2)
		b[0] = 0x01
		ll := len(l.lang)
		b[1] = uint8(ll)
		copy(b[2:], []byte(l.lang))
		copy(b[2+ll:], []byte(l.val))
		return b
	case XSDString:
		b := make([]byte, len(l.val)+1)
		b[0] = 0x02
		copy(b[1:], []byte(l.val))
		return b
	default:
		b := make([]byte, len(l.val)+len(l.DataType().val)+2)
		b[0] = 0xFF
		b[1] = uint8(len(l.DataType().val))
		copy(b[2:], []byte(l.DataType().val))
		copy(b[len(l.DataType().val)+2:], []byte(l.val))
		return b
	}
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
	return t.String() == other.String()
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
