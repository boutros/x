package rdf

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func collectTrErr(d *NTDecoder) (trs []Triple, errs []error) {
	for tr, err := d.Decode(); err != io.EOF; tr, err = d.Decode() {
		if err != nil {
			errs = append(errs, err)
		} else {
			trs = append(trs, tr)
		}
	}
	return trs, errs
}

func TestDecodeNT(t *testing.T) {
	tests := []struct {
		input   string
		trWant  []Triple
		errWant []error
	}{
		{
			"",
			[]Triple{},
			[]error{},
		},
		{
			"\n#comment <a> <b> <c> .\n",
			[]Triple{},
			[]error{},
		},
		{
			"\n\n\n<>.",
			[]Triple{},
			[]error{errors.New(`4: empty IRI: "<>"`)},
		},
		{
			"<s><p><o>.\n<s><p><o2>.",
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o")},
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o2")}},
			[]error{},
		},
		{
			"\n\n<s>\t<p> <o>.#comment\n#commment\n<s><p><o2>.",
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o")},
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o2")}},
			[]error{},
		},
		{
			"<s><p><o>.<z>\n",
			[]Triple{Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o")}},
			[]error{},
		},
		{
			"<s><p><o>.<z>\n<s><p><o2>.<y>",
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o")},
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o2")}},
			[]error{},
		},
		{
			"_:b1 <p> <o> .",
			[]Triple{},
			[]error{},
		},
		{
			"<s> <p> _:b2 .",
			[]Triple{},
			[]error{},
		},
		{
			"<s> <p> _:b2 .\n<S><P><O>.\n\t \n",
			[]Triple{Triple{subj: mustNewIRI("S"), pred: mustNewIRI("P"), obj: mustNewIRI("O")}},
			[]error{},
		},
		{
			`<s> <p> "".`,
			[]Triple{},
			[]error{errors.New(`1: empty literal: "\"\""`)},
		},
		{
			`<s> <p> "abc".`,
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewLiteral("abc")}},
			[]error{},
		},
		{
			`<s> <p> "hi"@en.`,
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewLangLiteral("hi", "en")}},
			[]error{},
		},
		{
			`<s> <p> "1"^^"a".`,
			[]Triple{},
			[]error{errors.New("1: expected IRI as literal datatype, got literal: \"a\"")},
		},
		/*{
			`<s> <p> "1"^^<http://www.w3.org/2001/XMLSchema#long>.`,
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewTypedLiteral("1", XSDLong)}},
			[]error{},
		},*/
	}

	for _, test := range tests {
		dec := NewNTDecoder(bytes.NewBufferString(test.input))
		trs, errs := collectTrErr(dec)

		if len(errs) != len(test.errWant) {
			t.Errorf("decoding:\n%q\ngot:\n%v\nwant:\n%v", test.input, errs, test.errWant)
		} else {
			for i, err := range test.errWant {
				if errs[i].Error() != err.Error() {
					t.Errorf("decoding:\n%q\ngot:\n%v\nwant:\n%v", test.input, errs, test.errWant)
				}
			}
		}

		if len(trs) != len(test.trWant) {
			t.Errorf("decoding:\n%q\ngot:\n%v\nwant:\n%v", test.input, trs, test.trWant)
		} else {
			for i, tr := range test.trWant {
				if !trs[i].Eq(tr) {
					t.Errorf("decoding:\n%q\ngot:\n%v\nwant:\n%v", test.input, trs, test.trWant)
				}
			}
		}
	}
}
