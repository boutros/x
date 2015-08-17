// Package bimap implements a bi-directional mapping between strings and uint16.
package bimap

import "errors"

const maxSize = 4294967295

// Exported errors
var (
	ErrEmptyString = errors.New("cannot store empty string")
	ErrNotFound    = errors.New("key not found")
	ErrToLarge     = errors.New("exceeded bimap max size (4294967295)")
)

// Map is a bi-directional mapping between strings and uint16.
type Map struct {
	strToInt map[string]uint16
	intToStr []string
}

// New creates a new, empty Map with an initial capacity. It will grow
// if needed (up too max uint16 size), but never shrink.
func New(cap int) *Map {
	if cap > maxSize {
		panic(ErrToLarge)
	}
	return &Map{
		strToInt: make(map[string]uint16, cap),
		intToStr: make([]string, cap),
	}
}

// Add adds a string/uint entry to the Map. It will fail if string is
// empty, or when maximum size of map is reached (TODO)
func (m *Map) Add(s string, i uint16) error {
	if s == "" {
		return ErrEmptyString
	}
	m.strToInt[s] = i
	if len(m.intToStr) <= int(i) {
		i2s := make([]string, min(len(m.intToStr)*2, maxSize))
		copy(i2s, m.intToStr)
		m.intToStr = i2s
	}
	m.intToStr[int(i)] = s
	return nil
}

// RemoveByStr removes string/uint pair with the given string. It returns
// ErrNotFound if the string is not present in map.
func (m *Map) RemoveByStr(s string) error {
	if i, ok := m.strToInt[s]; ok {
		delete(m.strToInt, s)
		m.intToStr[int(i)] = ""
		return nil
	}
	return ErrNotFound
}

// RemoveByInt removes string/uint pair with the given uint. It returns
// ErrNotFound if the uint is not present in map.
func (m *Map) RemoveByInt(i uint16) error {
	if int(i) > len(m.intToStr) {
		return ErrNotFound
	}
	delete(m.strToInt, m.intToStr[i])
	m.intToStr[int(i)] = ""
	return nil
}

// FindByStr returns the uint which is paired with the given string, if it exists.
func (m *Map) FindByStr(s string) (uint16, bool) {
	if i, ok := m.strToInt[s]; ok {
		return i, true
	}
	return 0, false
}

// FindByInt returns the string which is paired with the given uint, if it exists.
func (m *Map) FindByInt(i uint16) (string, bool) {
	if int(i) < len(m.intToStr) {
		if s := m.intToStr[int(i)]; s != "" {
			return s, true
		}
	}
	return "", false
}

// Size returns the size of the map (the number of string/uint pairs).
func (m *Map) Size() uint16 {
	return uint16(len(m.strToInt))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
