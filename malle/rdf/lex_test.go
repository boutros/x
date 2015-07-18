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
		if tk.Typ == tokenEOL || tk.Typ == tokenError {
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

func TestUnescape(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{`\t`, "\t"},
		{`\b`, "\b"},
		{`\n`, "\n"},
		{`\r`, "\r"},
		{`\f`, "\f"},
		{`\\`, "\\"},
		{`\u00b7`, "·"},
		{`\U000000b7`, "·"},
		{`\t\u00b7`, "\t·"},
		{`\b\U000000b7`, "\b·"},
		{`\u00b7\n`, "·\n"},
		{`\U000000b7\r`, "·\r"},
		{`\u00b7\f\U000000b7`, "·\f·"},
		{`\U000000b7\\\u00b7`, "·\\·"},
	}

	for _, tt := range tests {
		got := unescape([]byte(tt.in))
		if string(got) != tt.want {
			t.Errorf("unescape(%q) => %q; want %q", tt.in, got, tt.want)
		}
	}
}
func TestLexer(t *testing.T) {
	tests := []struct {
		in   string
		want []token
	}{
		{"", []token{}},
		{" \t ", []token{}},
		{"<>", []token{{tokenIRI, []byte("")}}},
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
	}

	for _, tt := range tests {
		lex := newLexer(strings.NewReader(tt.in))
		res := []token{}
		for _, tok := range collect(lex) {
			res = append(res, tok)
		}
		if res[len(res)-1].Typ == tokenEOL {
			res = res[:len(res)-1]
		}

		if !equalTokens(tt.want, res) {
			t.Errorf("lexing %q, got:\n\t%q\nwant\n\t%q", tt.in, res, tt.want)
		}
	}
}
