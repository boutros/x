package rdf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

type token struct {
	Typ   tokenType
	value []byte
}

type tokenType int

const eof = -1

const (
	tokenEOL tokenType = iota
	tokenError
	tokenIRI
	tokenDot
	tokenLiteral
	tokenBNode
	tokenLang
)

type lexer struct {
	r       *bufio.Reader
	input   []byte // current line being lexed
	line    int    // current line number
	pos     int    // position in line (in bytes, not runes)
	start   int    // start of current token
	escaped bool   // true when token needs to be unescaped before emitting
}

func newLexer(r io.Reader) *lexer {
	return &lexer{r: bufio.NewReader(r)}
}

func (l *lexer) readRune() rune {
	if l.pos == len(l.input) {
		input, err := l.r.ReadBytes('\n')
		if err != nil && len(input) == 0 {
			return eof
		}
		l.input = input
		l.start = 0
		l.pos = 0
		l.line++
	}

	r, w := utf8.DecodeRune(l.input[l.pos:])
	l.pos += w
	return r
}

func (l *lexer) emit(typ tokenType) token {
	return l.emitAndIgnore(typ, 0)
}

func (l *lexer) emitAndIgnore(typ tokenType, ignore int) token {
	s := l.start
	l.start = l.pos

	if l.escaped {
		l.escaped = false
		return token{
			Typ:   typ,
			value: unescape(l.input[s : l.pos-ignore]),
		}
	}

	return token{
		Typ:   typ,
		value: l.input[s : l.pos-ignore],
	}
}

func (l *lexer) error(msg string) token {
	s := l.start
	l.start = l.pos

	errMsg := fmt.Sprintf("%d: %s: %q", l.line, msg, l.input[s:l.pos])

	if l.escaped {
		l.escaped = false
	}

	return token{
		Typ:   tokenError,
		value: []byte(errMsg),
	}
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) consume(want rune) bool {
	for r := l.readRune(); r != want; r = l.readRune() {
		if r == eof {
			return false
		}
		if r == '\\' {
			l.escaped = true
			if l.readRune() == want {
				continue
			}
		}
	}
	return true
}

func (l *lexer) consumeUntilNextToken() {
	for r := l.readRune(); ; r = l.readRune() {
		switch r {
		case eof:
			return
		//case '\\':
		case '<', '.', '"', '#', ' ', '\t':
			l.pos--
			return
		}
	}
}

func (l *lexer) next() token {
	for {
		r := l.readRune()
		switch r {
		case ' ', '\t':
			l.ignore()
			continue
		case '<':
			if found := l.consume('>'); !found {
				return l.error("unclosed IRI")
			}
			l.start++ // ignore <
			return l.emitAndIgnore(tokenIRI, 1)
		case '.':
			l.ignore()
			return l.emit(tokenDot)
		case eof:
			return l.emit(tokenEOL)
		case '#':
			// comments are ignored and not emitted
			l.pos = len(l.input)
			l.ignore()
			return l.emit(tokenEOL)
		case '"':
			if found := l.consume('"'); !found {
				return l.error("unclosed Literal")
			}
			l.start++ // ignore starting "
			return l.emitAndIgnore(tokenLiteral, 1)
		case '_':
			r = l.readRune()
			if r != ':' {
				l.consumeUntilNextToken()
				return l.error("unexpected token")
			}
			l.ignore() // ignore _:
			l.consumeUntilNextToken()
			return l.emit(tokenBNode)
		case '@':
			l.ignore() // ignore @
			p := l.pos
			l.consumeUntilNextToken()
			if l.pos == p {
				return l.error("empty language tag")
			}
			return l.emit(tokenLang)
		default:
			l.consumeUntilNextToken()
			return l.error("unexpected token")
		}

	}
}

func unescape(text []byte) []byte {
	// TODO clean up this function, make it safe from index out of range errors (see text[i])
	buf := bytes.NewBuffer(make([]byte, 0, len(text)))
	i := 0
	for r, w := utf8.DecodeRune(text[i:]); w != 0; r, w = utf8.DecodeRune(text[i:]) {
		if r != '\\' {
			buf.WriteRune(r)
			i += w
			continue
		}
		i++
		var c byte
		switch text[i] {
		case 't':
			c = '\t'
		case 'b':
			c = '\b'
		case 'n':
			c = '\n'
		case 'r':
			c = '\r'
		case 'f':
			c = '\f'
		case '"':
			c = '"'
		case '\'':
			c = '\''
		case '\\':
			c = '\\'
		case 'u':
			rc, err := strconv.ParseInt(string(text[i+1:i+5]), 16, 32)
			if err != nil {
				panic(fmt.Errorf("internal parser error: %v", err))
			}
			buf.WriteRune(rune(rc))
			i += 5
			continue
		case 'U':
			rc, err := strconv.ParseInt(string(text[i+1:i+9]), 16, 32)
			if err != nil {
				panic(fmt.Errorf("internal parser error: %v", err))
			}
			buf.WriteRune(rune(rc))
			i += 9
			continue
		}
		buf.WriteByte(c)
		i++
	}

	return buf.Bytes()
}
