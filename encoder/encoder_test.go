package encoder

import (
	"math"
	// "log"
	"reflect"
	"testing"
)

var records = [][]interface{}{
	[]interface{}{"a", 1},
	[]interface{}{"b", 1},
	[]interface{}{"c", 2}}

var f *Field
var c Counter

func init() {
	f = NewField()
	f.Add("apple")
	f.Add("apple")
	f.Add("orange")

	for _, record := range records {
		c.Count(record)
	}
}

func TestFieldRecordCount(t *testing.T) {
	if f.RecordCount != 3 {
		t.Fail()
	}
}

func TestFieldTermId(t *testing.T) {
	if f.Terms["apple"] != 0 {
		t.Fail()
	}
}

func TestFieldCounts(t *testing.T) {
	if f.Counts[0] != 2 {
		t.Fail()
	}
}

func TestFieldCountResizing(t *testing.T) {
	for i := 0; i < 10000; i++ {
		f.Add(i)
	}
}

func TestCounterTerms(t *testing.T) {
	if c.Fields[0].Terms["c"] != 2 {
		t.Fail()
	}

	if c.Fields[1].Terms[2] != 1 {
		t.Fail()
	}
}

func TestCounterCounts(t *testing.T) {
	if !reflect.DeepEqual(c.Fields[0].Counts, []int{1,1,1}) {
		t.Fail()
	}
	if !reflect.DeepEqual(c.Fields[1].Counts, []int{2,1}) {
		t.Fail()
	}
}

func TestEncoder(t *testing.T) {
	e := NewEncoder(&c)
	v := *e.Encode(records[0])

	if math.Abs(1.0986122886681096 - v.Component(0)) > 0.00001 {
		t.Fail()
	}

	if math.Abs(0.4054651081081644 - v.Component(3)) > 0.00001 {
		t.Fail()
	}
}
