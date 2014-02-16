// Copyright 2014 William H. St. Clair

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func TestSignature(t *testing.T) {
	tv := Vector{0.2, 0.3, 0.4, 0.5, 0.6}

	hv := []Interface{
		&Vector{1.0, -1.0, 0.0, 0.0, 0.0}, // 0.2 + -0.3 = -0.1
		&Vector{0.0, 1.0, 0.0, -1.0, 0.0}, // 0.3 + -0.4 = -0.1
		&Vector{-1.0, 0.0, 1.0, 0.0, 0.0}, // -0.2 + 0.4 = 0.2
		&Vector{0.0, -1.0, 0.0, 0.0, 1.0}} // -0.3 + 0.6 = 0.6

	sig := Signature(tv, hv)
	if sig != 0xc {
		t.Errorf("Incorrect hash signature")
	}
}
