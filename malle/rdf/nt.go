package rdf

import (
	"errors"
	"fmt"
	"io"
)

// NTDecoder is a decodes RDF triples i N-Triples format.
//
// Notes and (possible) deviations from W3 specification:
// * IRIs are not validated, except to make sure they are not empty.
// * Literal that have any of the XSD datatypes are not validated against their datatype;
//   ex this will not fail: "abc"^^<http://www.w3.org/2001/XMLSchema#int>
// * Any tokens after triple termination (.) until end of line are ignored.
// * Triples with blank nodes are ignored by default, but can be converted to IRIs.
type NTDecoder struct {
	lex        *lexer
	BNodeAsIRI bool   // If true, convert blank nodes to IRIs
	BNodeNS    string // Namespace for converted blank nodes
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
		if !d.BNodeAsIRI {
			d.ignoreLine()
			goto newLine
		}
		tr.subj = IRI(d.BNodeNS + tok.value)
	} else {
		tr.subj = IRI(tok.value)
	}

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

	tr.pred = IRI(tok.value)

	// object
	tok, err = d.parseObject()
	if err != nil {
		d.ignoreLine()
		return Triple{}, err
	}
	if tok.Typ == tokenBNode {
		if !d.BNodeAsIRI {
			d.ignoreLine()
			goto newLine
		}
		tr.obj = IRI(d.BNodeNS + tok.value)
		goto dot
	}
	if tok.Typ == tokenIRI {
		tr.obj = IRI(tok.value)
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
			tr.obj = Literal{val: tok.value, datatype: XSDString}
			d.ignoreLine()
			return tr, nil
		case tokenLang:
			// rdf:langString
			tr.obj = Literal{val: tok.value, lang: peek.value, datatype: RDFLangString}
		case tokenDTMarker:
			// typed literal
			peek = d.lex.next()
			if peek.Typ != tokenIRI {
				d.ignoreLine()
				return Triple{}, fmt.Errorf("%d: expected IRI as literal datatype, got %v: %q", d.lex.line, tok.Typ, string(peek.value))
			}
			switch string(peek.value) {
			case "http://www.w3.org/2001/XMLSchema#string":
				tr.obj = Literal{val: tok.value, datatype: XSDString}
			case "http://www.w3.org/2001/XMLSchema#long":
				tr.obj = Literal{val: tok.value, datatype: XSDLong}
			default:
				tr.obj = Literal{val: tok.value, datatype: IRI(peek.value)}
			}
		default:
			panic("TODO parse Literal object")
		}
	}

dot:
	// dot+newline/eof
	err = d.parseEnd()
	if err != nil {
		d.ignoreLine()
		return Triple{}, err
	}

	return tr, nil
}

// DecodeAll consumes stream until the end and decodes all triples into a Graph.
func (d *NTDecoder) DecodeAll() Graph {
	g := NewGraph()
	for tr, err := d.Decode(); err != io.EOF; tr, err = d.Decode() {
		if err == nil {
			g.Add(tr)
		}
	}
	return g
}
