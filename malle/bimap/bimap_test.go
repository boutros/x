package bimap

import (
	"fmt"
	"testing"
)

func TestBimap(t *testing.T) {
	m := New(100)
	if m.Size() != 0 {
		t.Errorf("New() should return a map with Size 0, got %d", m.Size())
	}

	err := m.Add("", uint32(1))
	if err != ErrEmptyString {
		t.Errorf("Bimap.Add() empty string should fail with ErrEmptyString")
	}

	err = m.Add("a", uint32(1))
	if err != nil {
		t.Errorf("Bimap.Add() failed with %v", err)
	}

	i, ok := m.FindByStr("a")
	if !ok || i != uint32(1) {
		t.Errorf("FindByStr(\"a\") => %d, %v, want 1, true", i, ok)
	}

	s, ok := m.FindByInt(uint32(1))
	if !ok || s != "a" {
		t.Errorf("FindByStr(1) => %s, %v, want \"a\", true", s, ok)
	}

	err = m.Add("b", uint32(99))
	if err != nil {
		t.Errorf("Bimap.Add() failed with %v", err)
	}

	if m.Size() != 2 {
		t.Errorf("Bimap.Size() => %d; want 2", m.Size())
	}

	err = m.RemoveByStr("c")
	if err != ErrNotFound {
		t.Errorf("Bimap.RemoveByStr(\"c\") => %v; want ErrNotFound", err)
	}

	err = m.RemoveByStr("a")
	if err != nil {
		t.Errorf("Bimap.RemoveByStr(\"a\") => %v; want no error", err)
	}

	_, ok = m.FindByStr("a")
	if ok {
		t.Errorf("Bimap.RemoveByStr(\"a\") didn't remove map entry")
	}

	_, ok = m.FindByInt(uint32(1))
	if ok {
		t.Errorf("Bimap.RemoveByStr(\"a\") didn't remove map entry")
	}

	err = m.RemoveByInt(uint32(99))
	if err != nil {
		t.Errorf("Bimap.RemoveByStr(99) => %v; want no error", err)
	}

	_, ok = m.FindByInt(uint32(99))
	if ok {
		t.Errorf("Bimap.RemoveByInt(99) didn't remove map entry")
	}

	_, ok = m.FindByStr("b")
	if ok {
		t.Errorf("Bimap.RemoveByInt(99) didn't remove map entry")
	}

	if m.Size() != 0 {
		t.Errorf("Bimap.Size() => %d; want 1", m.Size())
	}
}

func TestGrowingBimap(t *testing.T) {
	m := New(2)
	for i := 0; i < 1000; i++ {
		err := m.Add(fmt.Sprintf("n:%d", i), uint32(i))
		if err != nil {
			t.Fatal(err)
		}
	}
}
