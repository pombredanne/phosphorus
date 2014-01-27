package vector

import (
	"math"
	"testing"
)

var vectorA = Vector{1.0, 0.0}
var vectorB = Vector{1.0, 1.0}

func _assertInt(t *testing.T, expected int, actual int, field string) {
	if expected != actual {
		t.Errorf("Expected %s=%d, not %d", field, expected, actual)
	}
}

func _assertFloat(t *testing.T, expected float64, actual float64, field string) {
	if math.Abs(expected-actual) > 0.0001 {
		t.Errorf("Expected %s=%g, not %g", field, expected, actual)
	}
}

func TestCosine(t *testing.T) {
	cos, _ := Cosine(vectorA, vectorB)
	if cos != 1.0/math.Sqrt(2) {
		t.Errorf("Expected 1.0/sqrt(2), not %f", cos)
	}
}

func TestSparseDot(t *testing.T) {
	a := &SparseVector{5, []SparseVectorComponent{
		SparseVectorComponent{1, 1.0},
		SparseVectorComponent{3, 1.0},
		SparseVectorComponent{4, 1.0}}}

	b := &SparseVector{5, []SparseVectorComponent{
		SparseVectorComponent{1, 5.0},
		SparseVectorComponent{2, 2.0},
		SparseVectorComponent{4, 4.0}}}

	dotProduct, _ := Dot(a, b)
	if dotProduct != 9.0 {
		t.Errorf("Expected 9.0, not %f", dotProduct)
	}
}

func TestMixedDot(t *testing.T) {
	a := Vector{0.0, 1.0, 0.0, 1.0, 1.0}
	b := &SparseVector{5, []SparseVectorComponent{
		SparseVectorComponent{1, 5.0},
		SparseVectorComponent{2, 2.0},
		SparseVectorComponent{4, 4.0}}}

	dotProduct, _ := Dot(a, b)
	if dotProduct != 9.0 {
		t.Errorf("Expected 9.0, not %f", dotProduct)
	}
}

func TestField(t *testing.T) {
	f := new(Field)
	f.Add("apple")
	_assertInt(t, 1, f.TermCount, "TermCount")
	f.Add("apple")
	f.Add("orange")
	_assertInt(t, 3, f.RecordCount, "RecordCount")
	_assertInt(t, 1, f.Terms["apple"], "Terms[\"apple\"]")
	_assertInt(t, 2, f.Counts[1], "Counts[1]")
}

func TestFieldCountResizing(t *testing.T) {
	f := new(Field)
	for i := 0; i < 10000; i++ {
		f.Add(i)
	}
}

func TestWeighting(t *testing.T) {
	testValues := []interface{}{1, "a", "a", "a", "c", "a", "a", "a", "a", "a", "c", 1, 1, "c", "a", "c"}
	f := new(Field)
	for _, term := range testValues {
		f.Add(term)
	}

	_, weight_a, _ := f.TermWeight("a")
	_assertFloat(t, 0.5753641449035618, weight_a, "TermWeight(\"a\")")
	_, weight_1, _ := f.TermWeight(1)
	_assertFloat(t, 1.6739764335716716, weight_1, "TermWeight(1)")

}

type dummyRecord struct {
	f []interface{}
}

func (r *dummyRecord) Fields() []interface{} {
	return r.f
}

func (r *dummyRecord) Id() uint {
	return uint(0)
}

func TestVectorize(t *testing.T) {
	var c Classifier

	testRecords := []Record{
		&dummyRecord{[]interface{}{"a", 1}},
		&dummyRecord{[]interface{}{"b", 1}},
		&dummyRecord{[]interface{}{"c", 2}}}

	for _, r := range testRecords {
		c.Classify(r)
	}
	v, _ := c.Vector(testRecords[0])
	_assertFloat(t, 1.0986122886681096, v.Component(0), "v[0]")
	_assertFloat(t, 0, v.Component(1), "v[1]")
	_assertFloat(t, 0, v.Component(2), "v[2]")
	_assertFloat(t, 0.4054651081081644, v.Component(3), "v[3]")
	_assertFloat(t, 0, v.Component(4), "v[4]")
}

func TestHash(t *testing.T) {
}

func TestSignature(t *testing.T) {
	tv := Vector{0.2, 0.3, 0.4, 0.5, 0.6}

	hv := []Interface{
		&Vector{1.0, -1.0, 0.0, 0.0, 0.0}, // 0.2 + -0.3 = -0.1
		&Vector{0.0, 1.0, 0.0, -1.0, 0.0}, // 0.3 + -0.4 = -0.1
		&Vector{-1.0, 0.0, 1.0, 0.0, 0.0}, // -0.2 + 0.4 = 0.2
		&Vector{0.0, -1.0, 0.0, 0.0, 1.0}} // -0.3 + 0.6 = 0.6

	sig, _ := Signature(tv, hv)
	if sig != 0xc {
		t.Errorf("Incorrect hash signature")
	}
}

func TestSignatureSet(t *testing.T) {
	tv := Vector{0.2, 0.3, 0.4, 0.5, 0.6}

	hf := [][]Interface{
		[]Interface{
			&Vector{1.0, -1.0, 0.0, 0.0, 0.0}, // 0.2 + -0.3 = -0.1
			&Vector{0.0, 1.0, 0.0, -1.0, 0.0}, // 0.3 + -0.4 = -0.1
			&Vector{-1.0, 0.0, 1.0, 0.0, 0.0}, // -0.2 + 0.4 = 0.2
			&Vector{0.0, -1.0, 0.0, 0.0, 1.0}}, // -0.3 + 0.6 = 0.6
		[]Interface{
			&Vector{1.0, -1.0, 0.0, 0.0, 0.0}, // 0.2 + -0.3 = -0.1
			&Vector{-1.0, 0.0, 1.0, 0.0, 0.0}, // -0.2 + 0.4 = 0.2
			&Vector{0.0, 1.0, 0.0, -1.0, 0.0}, // 0.3 + -0.4 = -0.1
			&Vector{0.0, -1.0, 0.0, 0.0, 1.0}}} // -0.3 + 0.6 = 0.6

	sigset, _ := SignatureSet(tv, hf)
	if sigset[0] != 0xc || sigset[1] != 0xa {
		t.Errorf("%s", "Incorrect signature set")
	}
}
