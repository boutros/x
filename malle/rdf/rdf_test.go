package rdf

import "testing"

func mustNewIRI(iri string) IRI {
	i, err := NewIRI(iri)
	if err != nil {
		panic(err)
	}
	return i
}

func mustNewLiteral(val interface{}) Literal {
	l, err := NewLiteral(val)
	if err != nil {
		panic(err)
	}
	return l
}

func mustNewLangLiteral(val, lang string) Literal {
	l, err := NewLangLiteral(val, lang)
	if err != nil {
		panic(err)
	}
	return l
}

func mustNewTypedLiteral(val string, tp IRI) Literal {
	l, err := NewTypedLiteral(val, tp)
	if err != nil {
		panic(err)
	}
	return l
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	tests := []Term{
		mustNewIRI("a"),
		mustNewIRI("http://example.org/1/xyz.æøå"),
		mustNewLangLiteral("xyz", "en"),
		mustNewLangLiteral("æøå", "nb-no"),
		mustNewTypedLiteral("a", XSDString),
		mustNewLiteral("En litt lengre streng\nmed æøå\tog andre tegn!"),
		mustNewTypedLiteral("101", mustNewIRI("http://ex.org/binary")),
		//mustNewLiteral(1),
		//mustNewLiteral(-4341581235912348234),
		//mustNewLiteral(uint(33)),
	}

	for _, t1 := range tests {
		t2, err := DecodeTerm(t1.Encode())
		if err != nil {
			t.Errorf("DecodeTerm(%v) err => %v ; want <nil>", t1.Encode(), err)
			continue
		}
		if !TermsEq(t1, t2) {
			t.Errorf("Encode/Decode roundtrip failed: %v != %v", t1, t2)
		}
	}
}

func TestTermDecode(t *testing.T) {
	tests := []struct {
		in      []byte
		want    Term
		errWant error
	}{
		{nil, nil, ErrUndecodable},
		{[]byte{}, nil, ErrUndecodable},
		{[]byte{0x00}, nil, ErrUndecodable},
		{[]byte{0x00, 0x61}, mustNewIRI("a"), nil},
		{[]byte{0x00, 0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x31, 0x2f, 0xc3, 0xa6, 0xc3, 0xb8, 0xc3, 0xa5},
			mustNewIRI("http://xyz/1/æøå"), nil},
		{[]byte{0x01}, nil, ErrUndecodable},
		{[]byte{0x01, 0x00}, nil, ErrUndecodable},
		{[]byte{0x01, 0x00, 0x61}, mustNewLiteral("a"), nil},
		{[]byte{0x01, 0x01, 0x61, 0x61}, mustNewLangLiteral("a", "a"), nil},
		{[]byte{0x01, 0x02, 0x65, 0x6e, 0x68, 0x69}, mustNewLangLiteral("hi", "en"), nil},
		{[]byte{0x01, 0x05, 0x6e, 0x62, 0x2d, 0x6e, 0x6f, 0x68, 0x65, 0x69}, mustNewLangLiteral("hei", "nb-no"), nil},
	}
	for _, tt := range tests {
		term, err := DecodeTerm(tt.in)
		if (err != nil && err != tt.errWant) || (err == nil && !TermsEq(tt.want, term)) {
			t.Errorf("DecodeTerm(%v) == %v, %v; want %v, %v", tt.in, term, err, tt.want, tt.errWant)
		}
	}
}

func TestTermNT(t *testing.T) {
	tests := []struct {
		in   Term
		want string
	}{
		{mustNewIRI("a"), "<a>"},
		{mustNewIRI("http://example.org/1/xyz.æøå"), "<http://example.org/1/xyz.æøå>"},
		{mustNewLangLiteral("xyz", "en"), `"xyz"@en`},
		{mustNewLangLiteral("æøå", "nb-no"), `"æøå"@nb-no`},
		{mustNewTypedLiteral("a", XSDString), `"a"`},
		{mustNewTypedLiteral("101", mustNewIRI("http://ex.org/binary")), `"101"^^<http://ex.org/binary>`},
		{mustNewLiteral(1), `"1"^^<http://www.w3.org/2001/XMLSchema#long>`},
		{mustNewLiteral(-4341581235912348234), `"-4341581235912348234"^^<http://www.w3.org/2001/XMLSchema#long>`},
		{mustNewLiteral(uint(33)), `"33"^^<http://www.w3.org/2001/XMLSchema#unsignedLong>`},
	}

	for _, tt := range tests {
		if tt.in.NT() != tt.want {
			t.Errorf("%+v.NT() == %s; want %s", tt.in, tt.in.NT(), tt.want)
		}
	}
}