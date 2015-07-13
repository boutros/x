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
	line  int
	col   int
}

type tokenType int

const eof = -1

const (
	tokenEOL tokenType = iota
	tokenError
	tokenIRI
	tokenDot
	tokenLiteral
)

type lexer struct {
	r       *bufio.Reader
	line    []byte
	pos     int
	start   int
	escaped bool
}

func newLexer(r io.Reader) *lexer {
	return &lexer{r: bufio.NewReader(r)}
}

func (l *lexer) readRune() rune {
	if l.pos == len(l.line) {
		line, err := l.r.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			return eof
		}
		l.line = line
		l.start = 0
		l.pos = 0
	}

	r, w := utf8.DecodeRune(l.line[l.pos:])
	l.pos += w
	return r
}

func (l *lexer) emit(typ tokenType) token {
	s := l.start
	l.start = l.pos

	if l.escaped {
		l.escaped = false
		return token{
			Typ:   typ,
			value: unescape(l.line[s:l.pos]),
		}
	}

	return token{
		Typ:   typ,
		value: unescape(l.line[s:l.pos]),
	}
}

func (l *lexer) emitAndIgnore(typ tokenType, ignore int) token {
	s := l.start
	l.start = l.pos

	if l.escaped {
		l.escaped = false
		return token{
			Typ:   typ,
			value: unescape(l.line[s : l.pos-ignore]),
		}
	}

	return token{
		Typ:   typ,
		value: l.line[s : l.pos-ignore],
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
		}
	}
	return true
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
				return l.emit(tokenError)
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
			l.pos = len(l.line)
			l.ignore()
			return l.emit(tokenEOL)
		case '"':
			if found := l.consume('"'); !found {
				return l.emit(tokenError)
			}
			l.start++ // ignore starting "
			return l.emitAndIgnore(tokenLiteral, 1)
		}

	}
}

func unescape(text []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(text)))
	i := 0
	for r, w := utf8.DecodeRune(text[i:]); w != 0; r, w = utf8.DecodeRune(text[i:]) {
		switch r {
		case '\\':
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
		default:
			buf.WriteRune(r)
			i += w
		}
	}

	return buf.Bytes()
}
