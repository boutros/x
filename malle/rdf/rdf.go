package rdf

import (
	"errors"
	"fmt"
	"strconv"
)

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

var (
	ErrUndecodable    = errors.New("rdf: cannot decode bytes into Term")
	ErrInvalidIRI     = errors.New("rdf: invalid IRI: cannot be empty")
	ErrInvalidLiteral = errors.New("rdf: invalid Literal: cannot be empty")
)

// Term represents a RDF Term
type Term interface {
	// Encode returns a byte representation of a term.
	Encode() []byte

	// NT returns a string representation of a Term in N-Triples format.
	NT() string
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
	return a != nil && b != nil && a.NT() == b.NT()
}

// IRI represents a IRI resource.
type IRI struct {
	val string
}

func NewIRI(iri string) (IRI, error) {
	if len(iri) == 0 {
		return IRI{}, ErrInvalidIRI
	}
	return IRI{val: iri}, nil
}

func (i IRI) Encode() []byte {
	b := make([]byte, len(i.val)+1)
	b[0] = 0x00
	copy(b[1:], []byte(i.val))
	return b
}

func (i IRI) NT() string {
	return fmt.Sprintf("<%s>", i.val)
}

type Literal struct {
	val      string
	lang     string
	dataType IRI
}

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

func NewTypedLiteral(val string, typ IRI) (Literal, error) {
	if len(val) == 0 {
		return Literal{}, ErrInvalidLiteral
	}
	return Literal{
		val:      val,
		dataType: typ,
	}, nil
}

func (l Literal) DataType() IRI {
	return l.dataType
}

func (l Literal) Lang() string {
	return l.lang
}

func (l Literal) NT() string {
	if l.lang != "" {
		return fmt.Sprintf("\"%s\"@%s", l.val, l.lang)
	} else if TermsEq(l.DataType(), XSDString) {
		return fmt.Sprintf("\"%s\"", l.val)
	}
	return fmt.Sprintf("\"%s\"^^%s", l.val, l.dataType.NT())
}

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

func (t Triple) Subject() IRI {
	return t.subj
}

func (t Triple) Predicate() IRI {
	return t.pred
}

func (t Triple) Object() Term {
	return t.obj
}

// NT returns an N-triples serialization of the Triple.
func (t Triple) NT() string {
	return fmt.Sprintf("%s %s %s .", t.subj.NT(), t.pred.NT(), t.obj.NT())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
