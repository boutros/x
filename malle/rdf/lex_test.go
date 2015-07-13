package rdf

import (
	"strings"
	"testing"
)

// testToken is a token without line and column positions,
// to make it easier to test
type testToken struct {
	Typ   tokenType
	value string
}

func collect(l *lexer) []testToken {
	tokens := []testToken{}
	for {
		tk := l.next()
		tokens = append(tokens, testToken{tk.Typ, string(tk.value)})
		if tk.Typ == tokenEOL || tk.Typ == tokenError {
			break
		}

	}
	return tokens
}

func equalTokens(a, b []testToken) bool {
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
		want []testToken
	}{
		{"", []testToken{}},
		{" \t ", []testToken{}},
		{"<a>", []testToken{{tokenIRI, "a"}}},
		{"<a", []testToken{{tokenError, "<a"}}},
		{" <http://xyz/æøå.123> \t ", []testToken{{tokenIRI, "http://xyz/æøå.123"}}},
		{"<a><b> <c> .", []testToken{{tokenIRI, "a"}, {tokenIRI, "b"}, {tokenIRI, "c"}, {tokenDot, ""}}},
		{"# a comment <a>", []testToken{}},
		{"<a> # a comment <b>", []testToken{{tokenIRI, "a"}}},
		{`"abc"`, []testToken{{tokenLiteral, "abc"}}},
		{`"line #1\nline #2"`, []testToken{{tokenLiteral, "line #1\nline #2"}}},
	}

	for _, tt := range tests {
		lex := newLexer(strings.NewReader(tt.in))
		res := []testToken{}
		for _, t := range collect(lex) {
			res = append(res, t)
		}
		if res[len(res)-1].Typ == tokenEOL {
			res = res[:len(res)-1]
		}

		if !equalTokens(tt.want, res) {
			t.Errorf("lexing %q, got:\n\t%q\nwant\n\t%q", tt.in, res, tt.want)
		}
	}
}
