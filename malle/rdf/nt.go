package rdf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// NTDecoder is a decodes RDF triples i N-Triples format.
//
// Notes and (possible) deviations from W3 specification:
// * IRIs are not validated, except to make sure they are not empty.
// * Literal that have any of the XSD datatypes are validated against their datatype;
//   ex this will fail: "abc"^^<http://www.w3.org/2001/XMLSchema#int>
// * Any tokens after triple termination (.) until end of line are ignored.
// * Triples with blank nodes are ignored.
type NTDecoder struct {
	lex *lexer
}

// NewNTDecoder returns a new NTDecoder on the given stream.
func NewNTDecoder(r io.Reader) *NTDecoder {
	return &NTDecoder{lex: newLexer(r)}
}

func (d *NTDecoder) ignoreLine() {
	for tok := d.lex.next(); tok.Typ != tokenEOL && tok.Typ != tokenEOF; tok = d.lex.next() {
	}
}

func (d *NTDecoder) parseSubject() (token, error) {
	var tok token
	for tok = d.lex.next(); tok.Typ == tokenEOL; tok = d.lex.next() {
	}
	switch tok.Typ {
	case tokenError:
		return token{}, errors.New(string(tok.value))
	case tokenEOF:
		return token{}, io.EOF
	case tokenBNode:
		return tok, nil
	case tokenIRI:
		break
	default:
		return token{}, fmt.Errorf("expected IRI as subject, got %s", tok.Typ.String())
	}

	return tok, nil
}

func (d *NTDecoder) parsePredicate() (token, error) {
	tok := d.lex.next()
	switch tok.Typ {
	case tokenError:
		return token{}, errors.New(string(tok.value))
	case tokenEOF:
		return token{}, io.EOF
	case tokenBNode:
		return tok, nil
	case tokenIRI:
		break
	default:
		return token{}, fmt.Errorf("expected IRI as predicate, got %s", tok.Typ.String())
	}
	return tok, nil
}

func (d *NTDecoder) parseObject() (token, error) {
	tok := d.lex.next()
	switch tok.Typ {
	case tokenError:
		return token{}, errors.New(string(tok.value))
	case tokenEOF:
		return token{}, io.EOF
	case tokenIRI, tokenLiteral, tokenBNode:
		break
	default:
		return token{}, fmt.Errorf("expected IRI or Literal as object, got %s", tok.Typ.String())
	}

	return tok, nil
}

func (d *NTDecoder) parseEnd() error {
	// Each statement must end in a dot (.)
	tok := d.lex.next()
	switch tok.Typ {
	case tokenError:
		return errors.New(string(tok.value))
	case tokenEOF:
		return io.EOF
	case tokenDot:
		break
	default:
		return fmt.Errorf("expected dot, got %s", tok.Typ.String())
	}

	// Any tokens after dot until end of line are ignored
	d.ignoreLine()

	return nil
}

// Decode returns the next valid triple in the the stream, or an error.
func (d *NTDecoder) Decode() (Triple, error) {
	var tr Triple
newLine:
	// subject
	tok, err := d.parseSubject()
	if err != nil {
		d.ignoreLine()
		return Triple{}, err
	}
	if tok.Typ == tokenBNode {
		d.ignoreLine()
		goto newLine
	}

	b := make([]byte, len(tok.value)+1)
	copy(b[1:], tok.value)
	tr.subj = IRI{val: b}

	// predicate
	tok, err = d.parsePredicate()
	if err != nil {
		d.ignoreLine()
		return Triple{}, err
	}
	if tok.Typ == tokenBNode {
		d.ignoreLine()
		goto newLine
	}

	b = make([]byte, len(tok.value)+1)
	copy(b[1:], tok.value)
	tr.pred = IRI{val: b}

	// object
	tok, err = d.parseObject()
	if err != nil {
		d.ignoreLine()
		return Triple{}, err
	}
	if tok.Typ == tokenBNode {
		d.ignoreLine()
		goto newLine
	}
	if tok.Typ == tokenIRI {
		b = make([]byte, len(tok.value)+1)
		copy(b[1:], tok.value)
		tr.obj = IRI{val: b}
	} else {
		// literal
		peek := d.lex.next()
		switch peek.Typ {
		case tokenEOL:
			return Triple{}, errors.New("expected dot, got EOL")
		case tokenError:
			d.ignoreLine()
			return Triple{}, errors.New(string(peek.value))
		case tokenDot:
			// plain literal xsd:String
			b = make([]byte, len(tok.value)+1)
			b[0] = 0x02
			copy(b[1:], tok.value)
			tr.obj = Literal{val: b}
			d.ignoreLine()
			return tr, nil
		case tokenLang:
			// rdf:langString
			ll := len(peek.value)
			b := make([]byte, len(tok.value)+ll+2)
			b[0] = 0x01
			b[1] = uint8(ll)
			copy(b[2:], []byte(peek.value))
			copy(b[2+ll:], []byte(tok.value))
			tr.obj = Literal{val: b}
		case tokenDTMarker:
			// typed literal
			peek = d.lex.next()
			if peek.Typ != tokenIRI {
				d.ignoreLine()
				return Triple{}, fmt.Errorf("%d: expected IRI as literal datatype, got %v: %q", d.lex.line, tok.Typ, string(peek.value))
			}
			switch string(peek.value) {
			case "http://www.w3.org/2001/XMLSchema#string":
				b = make([]byte, len(tok.value)+1)
				b[0] = 0x02
				copy(b[1:], tok.value)
				tr.obj = Literal{val: b}
			case "http://www.w3.org/2001/XMLSchema#long":
				i, err := strconv.ParseInt(string(tok.value), 10, 64)
				if err != nil {
					d.ignoreLine()
					return Triple{}, fmt.Errorf("%d: literal does not match its datatype (xsd:long): %q", d.lex.line, string(tok.value))
				}
				b = make([]byte, 1+binary.MaxVarintLen64)
				b[0] = 0x03
				l := binary.PutVarint(b[1:], i)
				tr.obj = Literal{val: b[:l+1], typed: i}
			default:
				b = make([]byte, len(tok.value)+len(peek.value)+2)
				b[0] = 0xFF
				b[1] = uint8(len(peek.value))
				copy(b[2:], peek.value)
				copy(b[b[1]+2:], tok.value)
				tr.obj = Literal{val: b}
			}
		default:
			panic("TODO parse Literal object")
		}
	}

	// dot+newline/eof
	err = d.parseEnd()
	if err != nil {
		d.ignoreLine()
		return Triple{}, err
	}

	return tr, nil
}

// DecodeAll TODO
func (d *NTDecoder) DecodeAll() []Triple {
	return nil
}
