package schema

import (
	// "log"
	"testing"
)

var sig1 = []uint32{0, 255, 104, 172, 138, 51, 132, 248}
var sig2 = []uint32{0, 255, 104, 172, 138, 51, 232, 177}
var sig3 = []uint32{0, 255, 104, 197, 20, 149, 132, 62}

var rec1 = &Record{1, map[string]string{"first": "John", "last": "Doe"}}
var rec2 = &Record{2, map[string]string{"first": "Jane", "last": "Roe"}}

type _schema struct {
	fixture []uint32
}

func (s *_schema) Sign(map[string]string, RandomProvider) ([]uint32, error) {
	return s.fixture, nil
}

func (s *_schema) SignatureLen() int {
	return 8
}

func (s *_schema) ChunkBits() int {
	return 8
}

type _random struct{}

func (r *_random) Get(int64) float64 {
	return 0.0
}

func TestMemoryIndex(t *testing.T) {
	s := &_schema{sig1}
	ix := NewMemoryIndex(s)
	r := &_random{}

	err := ix.Write(rec1, r)
	if err != nil {
		t.Error(err)
	}

	s.fixture = sig2
	err = ix.Write(rec2, r)
	if err != nil {
		t.Error(err)
	}

	s.fixture = sig3

	results, err := ix.Query(map[string]string{}, r)
	if err != nil {
		t.Error(err)
	}

	if results[0].Record.Id != 1 {
		t.Fail()
	}

	if results[0].Matches != 4 {
		t.Fail()
	}

	if results[1].Record.Id != 2 {
		t.Fail()
	}

	if results[1].Matches != 3 {
		t.Fail()
	}
}
