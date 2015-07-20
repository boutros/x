package rdf

import (
	"errors"
	"fmt"
	"io"
)

// NTDecoder is a decodes RDF triples i N-Triples format.
type NTDecoder struct {
	lex *lexer
}

// NewNTDecoder returns a new NTDecoder on the given stream.
//
// Notes and deviations from W3 specification:
// * IRIs are not validated, except to make sure they are not empty.
// * Literal values are not validated against their datatype.
// * Any tokens after triple termination (.) until end of line are ignored.
// * Triples with blank nodes are ignored.
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
		panic("TODO parse Literal object")
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
