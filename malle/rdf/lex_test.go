package rdf

import (
	"bytes"
	"strings"
	"testing"
)

func collect(l *lexer) []token {
	tokens := []token{}
	for {
		tk := l.next()
		tokens = append(tokens, token{tk.Typ, tk.value})
		if tk.Typ == tokenEOF || tk.Typ == tokenError {
			break
		}

	}
	return tokens
}

func equalTokens(a, b []token) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k].Typ != b[k].Typ {
			return false
		}
		if !bytes.Equal(a[k].value, b[k].value) {
			return false
		}
	}
	return true
}

func TestLexer(t *testing.T) {
	tests := []struct {
		in   string
		want []token
	}{
		{"", []token{}},
		{" \t ", []token{}},
		{"<>", []token{{tokenError, []byte(`1: empty IRI: "<>"`)}}},
		{"<a>", []token{{tokenIRI, []byte("a")}}},
		{"<a", []token{{tokenError, []byte(`1: unclosed IRI: "<a"`)}}},
		{"a>", []token{{tokenError, []byte(`1: unexpected token: "a>"`)}}},
		{"abc.", []token{{tokenError, []byte(`1: unexpected token: "abc"`)}}},
		{" <http://xyz/æøå.123> \t ", []token{{tokenIRI, []byte("http://xyz/æøå.123")}}},
		{"<a><b> <c> .", []token{
			{tokenIRI, []byte("a")},
			{tokenIRI, []byte("b")},
			{tokenIRI, []byte("c")},
			{tokenDot, []byte("")}}},
		{"# a comment <a>", []token{}},
		{"<a> # a comment <b>", []token{{tokenIRI, []byte("a")}}},
		{`"abc"`, []token{{tokenLiteral, []byte("abc")}}},
		{`"line #1\nline #2"`, []token{{tokenLiteral, []byte("line #1\nline #2")}}},
		{"'abc'", []token{{tokenError, []byte(`1: unexpected token: "'abc'"`)}}},
		{`<s>"o`, []token{{tokenIRI, []byte("s")}, {tokenError, []byte(`1: unclosed Literal: "\"o"`)}}},
		{"_:b1", []token{{tokenBNode, []byte("b1")}}},
		{"_:abc44 <p>", []token{{tokenBNode, []byte("abc44")}, {tokenIRI, []byte("p")}}},
		{`<http://example/æøå> <http://example/禅> "\"\\\r\n Здра́вствуйте	☺" .`, []token{
			{tokenIRI, []byte("http://example/æøå")},
			{tokenIRI, []byte("http://example/禅")},
			{tokenLiteral, []byte("\"\\\r\n Здра́вствуйте\t☺")},
			{tokenDot, []byte("")}}},
		{`"\u006F \U0000006F"`, []token{{tokenLiteral, []byte("o o")}}},
		{`"hi"@en`, []token{{tokenLiteral, []byte("hi")}, {tokenLang, []byte("en")}}},
		{`"hei"@nb-no .`, []token{
			{tokenLiteral, []byte("hei")}, {tokenLang, []byte("nb-no")}, {tokenDot, []byte("")}}},
		{`@ en`, []token{{tokenError, []byte(`1: empty language tag: ""`)}}},
		{`^<a>`, []token{{tokenError, []byte("1: unexpected token: \"^\"")}}},
		{`"1"^^<a>`, []token{
			{tokenLiteral, []byte("1")},
			{tokenDTMarker, []byte("")},
			{tokenIRI, []byte("a")}}},
		{`""`, []token{{tokenError, []byte(`1: empty literal: "\"\""`)}}},
		{`"xy\z"`, []token{{tokenError, []byte(`1: illegal escape sequence: "\\z"`)}}},
		{`"\t\r\n\f\b\\\u00b7\u00B7\U000000b7\U000000B7"`, []token{{tokenLiteral, []byte("\t\r\n\f\b\\····")}}},
		{`"\u00F"`, []token{{tokenError, []byte(`1: illegal escape sequence: "\\u00F"`)}}},
		{`"\u123"`, []token{{tokenError, []byte(`1: illegal escape sequence: "\\u123"`)}}},
		{`"\u123ø."`, []token{{tokenError, []byte(`1: illegal escape sequence: "\\u123ø"`)}}},
		{"\"line 1\nline 2\"", []token{{tokenError, []byte(`1: unclosed Literal: "\"line 1\n"`)}}},
	}

	for _, tt := range tests {
		lex := newLexer(strings.NewReader(tt.in))
		res := []token{}
		for _, tok := range collect(lex) {
			res = append(res, tok)
		}
		if res[len(res)-1].Typ == tokenEOF {
			res = res[:len(res)-1]
		}
		if len(res) >= 1 && res[len(res)-1].Typ == tokenEOL {
			res = res[:len(res)-1]
		}

		if !equalTokens(tt.want, res) {
			t.Errorf("lexing %q, got:\n\t%q\nwant\n\t%q", tt.in, res, tt.want)
		}
	}
}
