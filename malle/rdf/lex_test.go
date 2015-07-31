package rdf

import (
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
		if a[k].value != b[k].value {
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
		{"<>", []token{{tokenError, `1: empty IRI: "<>"`}}},
		{"<a>", []token{{tokenIRI, "a"}}},
		{"<a", []token{{tokenError, `1: unclosed IRI: "<a"`}}},
		{"a>", []token{{tokenError, `1: unexpected token: "a>"`}}},
		{"abc.", []token{{tokenError, `1: unexpected token: "abc"`}}},
		{" <http://xyz/æøå.123> \t ", []token{{tokenIRI, "http://xyz/æøå.123"}}},
		{"<a><b> <c> .", []token{
			{tokenIRI, "a"},
			{tokenIRI, "b"},
			{tokenIRI, "c"},
			{tokenDot, ""}}},
		{"# a comment <a>", []token{}},
		{"<a> # a comment <b>", []token{{tokenIRI, "a"}}},
		{`"abc"`, []token{{tokenLiteral, "abc"}}},
		{`"line #1\nline #2"`, []token{{tokenLiteral, "line #1\nline #2"}}},
		{"'abc'", []token{{tokenError, `1: unexpected token: "'abc'"`}}},
		{`<s>"o`, []token{{tokenIRI, "s"}, {tokenError, `1: unclosed Literal: "\"o"`}}},
		{"_:b1", []token{{tokenBNode, "b1"}}},
		{"_:abc44 <p>", []token{{tokenBNode, "abc44"}, {tokenIRI, "p"}}},
		{`<http://example/æøå> <http://example/禅> "\"\\\r\n Здра́вствуйте	☺" .`, []token{
			{tokenIRI, "http://example/æøå"},
			{tokenIRI, "http://example/禅"},
			{tokenLiteral, "\"\\\r\n Здра́вствуйте\t☺"},
			{tokenDot, ""}}},
		{`"\u006F \U0000006F"`, []token{{tokenLiteral, "o o"}}},
		{`"hi"@en`, []token{{tokenLiteral, "hi"}, {tokenLang, "en"}}},
		{`"hei"@nb-no .`, []token{
			{tokenLiteral, "hei"}, {tokenLang, "nb-no"}, {tokenDot, ""}}},
		{`@ en`, []token{{tokenError, `1: empty language tag: ""`}}},
		{`^<a>`, []token{{tokenError, "1: unexpected token: \"^\""}}},
		{`"1"^^<a>`, []token{
			{tokenLiteral, "1"},
			{tokenDTMarker, ""},
			{tokenIRI, "a"}}},
		{`""`, []token{{tokenError, `1: empty literal: "\"\""`}}},
		{`"xy\z"`, []token{{tokenError, `1: illegal escape sequence: "\\z"`}}},
		{`"\t\r\n\f\b\\\u00b7\u00B7\U000000b7\U000000B7"`, []token{{tokenLiteral, "\t\r\n\f\b\\····"}}},
		{`"\u00F"`, []token{{tokenError, `1: illegal escape sequence: "\\u00F"`}}},
		{`"\u123"`, []token{{tokenError, `1: illegal escape sequence: "\\u123"`}}},
		{`"\u123ø."`, []token{{tokenError, `1: illegal escape sequence: "\\u123ø"`}}},
		{"\"line 1\nline 2\"", []token{{tokenError, `1: unclosed Literal: "\"line 1\n"`}}},
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
