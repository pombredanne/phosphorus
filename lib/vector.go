package lib

import (
	"fmt"
	"math"
	"math/rand"
)

type Interface interface {
	Dimension() int
	Norm() float64
	Component(dimension int) float64
}

type Vector []float64

func (v Vector) Dimension() int {
	return len(v)
}

func (v Vector) Norm() float64 {
	sum := 0.0
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum)
}

func (v Vector) Component(dimension int) float64 {
	return v[dimension]
}

type HashVector []uint16

func (v HashVector) Dimension() int {
	return len(v)
}

func (v HashVector) Norm() float64 {
	sum := 0.0
	for _, x := range v {
		val := Uncompact(x)
		sum += val * val
	}
	return math.Sqrt(sum)
}

func (v HashVector) Component(dimension int) float64 {
	return Uncompact(v[dimension])
}

type SparseVector struct {
	dimension  int
	Components []SparseVectorComponent
}

type SparseVectorComponent struct {
	Index int
	Value float64
}

func NewSparseVector(dim int, fields int) *SparseVector {
	return &SparseVector{dim,
		make([]SparseVectorComponent, fields)}
}

func (v *SparseVector) Dimension() int {
	return v.dimension
}

func (v *SparseVector) Norm() float64 {
	sum := 0.0
	for _, term := range v.Components {
		sum += term.Value * term.Value
	}
	return math.Sqrt(sum)
}

func (v *SparseVector) Component(dimension int) float64 {
	for _, term := range v.Components {
		if term.Index == dimension {
			return term.Value
		}
	}
	return 0.0
}

func vectorDot(a Interface, b Interface) float64 {
	sum := 0.0
	for i := 0; i < a.Dimension(); i++ {
		sum += a.Component(i) * b.Component(i)
	}
	return sum
}

func sparseDot(a *SparseVector, b *SparseVector) float64 {
	aIndex := 0
	bIndex := 0
	sum := 0.0

	for {
		aTerm := a.Components[aIndex]
		bTerm := b.Components[bIndex]

		if aTerm.Index < bTerm.Index {
			aIndex++
		} else if bTerm.Index < aTerm.Index {
			bIndex++
		} else {
			sum += aTerm.Value * bTerm.Value
			aIndex++
			bIndex++
		}
		if (aIndex > len(a.Components)-1) ||
			(bIndex > len(b.Components)-1) {
			break
		}
	}

	return sum
}

func mixedDot(a Interface, b *SparseVector) float64 {
	sum := 0.0
	for _, bTerm := range b.Components {
		sum += bTerm.Value * a.Component(bTerm.Index)
	}
	return sum
}

func Dot(a Interface, b Interface) (float64, error) {
	if a.Dimension() != b.Dimension() {
		return 0.0, fmt.Errorf("Mismatched dimensions: %d and %d", a.Dimension(), b.Dimension())
	}

	var product float64

	switch a.(type) {
	default:
		switch b.(type) {
		default:
			product = vectorDot(a.(Interface), b.(Interface))
		case *SparseVector:
			product = mixedDot(a.(Interface), b.(*SparseVector))
		}
	case *SparseVector:
		switch b.(type) {
		default:
			product = mixedDot(b.(Interface), a.(*SparseVector))
		case *SparseVector:
			product = sparseDot(a.(*SparseVector), b.(*SparseVector))
		}

	}
	return product, nil
}

func Cosine(a Interface, b Interface) (float64, error) {
	dot, err := Dot(a, b)
	if err != nil {
		return 0.0, err
	}
	return dot / (a.Norm() * b.Norm()), nil
}


func Random(dimension int) Interface {
	v := make(HashVector, dimension)
	for i := 0; i < dimension; i++ {
		v[i] = Compact(rand.NormFloat64())
	}
	return v
}

func Hash(x Interface, r Interface) (bool, error) {
	dot, err := Dot(r, x)
	if err != nil {
		return false, err
	}
	return dot >= 0.0, nil
}

func Signature(x Interface, r []HashVector) (uint16, error) {
	sig := uint16(0)
	for i, ri := range r {
		hash, err := Hash(x, ri)
		if err != nil {
			return 0, err
		}
		if hash {
			sig |= 1 << uint16(i)
		}
	}
	return sig, nil
}

func SignatureSet(r Interface, hf [][]HashVector) ([]uint16, error) {
	var signatures = make([]uint16, len(hf))
	var err error

	for i, h := range hf {
		signatures[i], err = Signature(r, h)
		if err != nil {
			return signatures, err
		}
	}
	return signatures, nil
}
