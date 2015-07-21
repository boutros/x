package rdf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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
	tokenEOF
	tokenError
	tokenIRI
	tokenDot
	tokenLiteral
	tokenBNode
	tokenLang
	tokenDTMarker
)

func (t tokenType) String() string {
	switch t {
	case tokenEOL:
		return "EOL"
	case tokenEOF:
		return "EOF"
	case tokenError:
		return "error"
	case tokenDot:
		return "dot (.)"
	case tokenLiteral:
		return "literal"
	case tokenBNode:
		return "blank node"
	case tokenLang:
		return "language tag"
	case tokenDTMarker:
		return "datatype marker (^^)"
	default:
		panic("TODO tokenType.String()")
	}
}

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

	return l.unescape(typ, l.input[s:l.pos-ignore])
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
			p := l.pos + 1
			if found := l.consume('>'); !found {
				return l.error("unclosed IRI")
			}
			if l.pos == p {
				return l.error("empty IRI")
			}
			l.start++ // ignore <
			return l.emitAndIgnore(tokenIRI, 1)
		case '.':
			l.ignore()
			return l.emit(tokenDot)
		case '\n':
			return l.emit(tokenEOL)
		case eof:
			return l.emit(tokenEOF)
		case '#':
			// comments are ignored and not emitted
			l.pos = len(l.input)
			l.ignore()
			return l.emit(tokenEOL)
		case '"':
			p := l.pos + 1
			if found := l.consume('"'); !found {
				return l.error("unclosed Literal")
			}
			if l.pos == p {
				return l.error("empty literal")
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
		case '^':
			r = l.readRune()
			if r != '^' {
				l.pos--
				l.consumeUntilNextToken()
				return l.error("unexpected token")
			}
			l.ignore() // ignore ^^
			return l.emit(tokenDTMarker)
		default:
			l.consumeUntilNextToken()
			return l.error("unexpected token")
		}

	}
}

func (l *lexer) unescape(typ tokenType, val []byte) token {
	if !l.escaped {
		return token{Typ: typ, value: val}
	}
	l.escaped = false
	switch typ {
	case tokenIRI:
		panic("TODO")
	case tokenLiteral:
		return l.unescapeLiteral(typ, val)
	default:
		return token{Typ: typ, value: val}
	}
}

func (l *lexer) unescapeLiteral(typ tokenType, text []byte) token {
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
		case 'u', 'U':
			d := uint64(0)
			start := i
			digits := 4
			if text[i] == 'U' {
				digits = 8
			}
			for i < start+digits {
				i++
				if i == len(text) {
					return token{
						Typ:   tokenError,
						value: []byte(fmt.Sprintf("%d: illegal escape sequence: %q", l.line, text[start-1:i]))}
				}
				x := uint64(text[i])
				if x >= 'a' {
					x -= 'a' - 'A'
				}
				d1 := x - '0'
				if d1 > 9 {
					d1 = 10 + d1 - ('A' - '0')
				}
				if 0 > d1 || d1 > 15 {
					j := i
					for !utf8.FullRune(text[j:i]) {
						i++
					}
					return token{
						Typ:   tokenError,
						value: []byte(fmt.Sprintf("%d: illegal escape sequence: %q", l.line, text[start-1:i]))}
				}
				d = (16 * d) + d1
			}
			buf.WriteRune(rune(d))
			i++
			continue
		default:
			return token{
				Typ:   tokenError,
				value: []byte(fmt.Sprintf("%d: illegal escape sequence: %q", l.line, text[i-1:]))}
		}
		buf.WriteByte(c)
		i++
	}

	return token{Typ: typ, value: buf.Bytes()}
}
