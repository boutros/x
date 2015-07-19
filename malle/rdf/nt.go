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
func NewNTDecoder(r io.Reader) *NTDecoder {
	return &NTDecoder{lex: newLexer(r)}
}

func (d *NTDecoder) parseSubject() (token, error) {
	var tok token
	for tok = d.lex.next(); tok.Typ == tokenEOL; tok = d.lex.next() {
	}
	if tok.Typ == tokenError {
		return token{}, errors.New(string(tok.value))
	}
	if tok.Typ == tokenEOF {
		return token{}, io.EOF
	}
	if tok.Typ != tokenIRI {
		return token{}, fmt.Errorf("expected IRI as subject, got %s", tok.Typ.String())
	}
	return tok, nil
}

func (d *NTDecoder) parsePredicate() (token, error) {
	tok := d.lex.next()
	if tok.Typ == tokenError {
		return token{}, errors.New(string(tok.value))
	}
	if tok.Typ == tokenEOF {
		return token{}, io.EOF
	}
	if tok.Typ != tokenIRI {
		return token{}, fmt.Errorf("expected IRI as predicate, got %s", tok.Typ.String())
	}
	return tok, nil
}

func (d *NTDecoder) parseObject() (token, error) {
	tok := d.lex.next()
	if tok.Typ == tokenError {
		return token{}, errors.New(string(tok.value))
	}
	if tok.Typ == tokenEOF {
		return token{}, io.EOF
	}
	if tok.Typ != tokenIRI && tok.Typ != tokenLiteral {
		return token{}, fmt.Errorf("expected IRI or Literal as object, got %s", tok.Typ.String())
	}
	return tok, nil
}

func (d *NTDecoder) parseDot() error {
	tok := d.lex.next()
	if tok.Typ == tokenError {
		return errors.New(string(tok.value))
	}
	if tok.Typ == tokenEOF {
		return io.EOF
	}
	if tok.Typ != tokenDot {
		return fmt.Errorf("expected dot, got %s", tok.Typ.String())
	}
	return nil
}

// Decode returns the next valid triple in the the stream, or an error.
func (d *NTDecoder) Decode() (Triple, error) {
	var tr Triple
	// subject
	tok, err := d.parseSubject()
	if err != nil {
		return Triple{}, err
	}

	b := make([]byte, len(tok.value)+1)
	copy(b[1:], tok.value)
	tr.subj = IRI{val: b}

	// predicate
	tok, err = d.parsePredicate()
	if err != nil {
		return Triple{}, err
	}
	b = make([]byte, len(tok.value)+1)
	copy(b[1:], tok.value)
	tr.pred = IRI{val: b}

	// object
	tok, err = d.parseObject()
	if err != nil {
		return Triple{}, err
	}
	if tok.Typ == tokenIRI {
		b = make([]byte, len(tok.value)+1)
		copy(b[1:], tok.value)
		tr.obj = IRI{val: b}
	} else {
		panic("TODO parse Literal object")
	}

	// dot
	err = d.parseDot()
	if err != nil {
		return Triple{}, err
	}

	return tr, nil
}

// DecodeAll TODO
func (d *NTDecoder) DecodeAll() []Triple {
	return nil
}
