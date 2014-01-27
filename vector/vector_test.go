package vector

import (
	"math"
	"testing"
)

var vectorA = Vector{1.0, 0.0}
var vectorB = Vector{1.0, 1.0}

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
