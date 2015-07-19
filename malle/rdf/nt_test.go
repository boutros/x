package rdf

import (
	"bytes"
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
			"<s><p><o>.\n<s><p><o2>.",
			[]Triple{
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o")},
				Triple{subj: mustNewIRI("s"), pred: mustNewIRI("p"), obj: mustNewIRI("o2")}},
			[]error{},
		},
	}

	for _, test := range tests {
		dec := NewNTDecoder(bytes.NewBufferString(test.input))
		trs, errs := collectTrErr(dec)
		for i, tr := range test.trWant {
			if !trs[i].Eq(tr) {
				t.Fatalf("decoding:\n%q\ngot:\n%v\nwant:\n%v", test.input, trs, test.trWant)
			}
		}

		for i, err := range test.errWant {
			if errs[i] != err {
				t.Fatalf("decoding:\n%q\ngot:\n%v\nwant:\n%v", test.input, errs, test.errWant)
			}
		}
	}
}
