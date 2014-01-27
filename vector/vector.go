package vector

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
		val := UncompactFloat(x)
		sum += val * val
	}
	return math.Sqrt(sum)
}

func (v HashVector) Component(dimension int) float64 {
	return UncompactFloat(v[dimension])
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

func CompactFloat(x float64) uint16 {
	return uint16(math.Floor((x + 8.0) * 4096.0))
}

func UncompactFloat(x uint16) float64 {
	return (float64(x) / 4096.0) - 8.0
}

func Random(dimension int) Interface {
	v := make(HashVector, dimension)
	for i := 0; i < dimension; i++ {
		v[i] = CompactFloat(rand.NormFloat64())
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

func Signature(x Interface, r ...Interface) (uint16, error) {
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

type Field struct {
	Initialized bool
	TermCount   int
	RecordCount int
	Terms       map[interface{}]int
	Counts      []int
}

func (f *Field) Initialize() {
	if !f.Initialized {
		f.Terms = make(map[interface{}]int)
		f.Counts = make([]int, 1, 1000)
		f.Initialized = true
	}
}

func (f *Field) Quantize(term interface{}) (int, error) {
	if !f.Initialized {
		return 0, fmt.Errorf("Field not Initialized")
	}
	term_id := f.Terms[term]
	if term_id == 0 {
		return 0, fmt.Errorf("Term not found: %q", term)
	}
	return term_id, nil
}

func (f *Field) Add(term interface{}) {
	f.Initialize()
	termId, err := f.Quantize(term)
	if err != nil {
		f.TermCount++
		f.Terms[term] = f.TermCount
		termId = f.TermCount
		f.Counts = append(f.Counts, 0)
	}
	// if termId > (len(f.Counts) - 1) {
	// 	newCounts := make([]int, len(f.Counts) * 2)
	// 	copy(newCounts, f.Counts)
	// 	f.Counts = newCounts
	// }
	f.Counts[termId]++
	f.RecordCount++
}

func (f *Field) Weight(termId int) (float64, error) {
	if !f.Initialized {
		return 0.0, fmt.Errorf("Field not Initialized")
	}
	return math.Log(float64(f.RecordCount) / float64(f.Counts[termId])), nil
}

func (f *Field) TermWeight(term interface{}) (int, float64, error) {
	termId, err := f.Quantize(term)
	if err != nil {
		return termId, 0.0, err
	}

	weight, err := f.Weight(termId)
	return termId, weight, err
}

type Record interface {
	Id() uint
	Fields()   []interface{}
}

type Classifier struct {
	Fields []Field
}

func (c *Classifier) Classify(r Record) {
	for i, term := range r.Fields() {
		if len(c.Fields) <= i {
			c.Fields = append(c.Fields, *new(Field))
		}
		c.Fields[i].Add(term)
	}
}

func (c *Classifier) Dimension() int {
	dimensionCount := 0
	for _, field := range c.Fields {
		dimensionCount += field.TermCount
	}
	return dimensionCount
}

func (c *Classifier) Vector(r Record) (*SparseVector, error) {
	recordFields := r.Fields()
	v := NewSparseVector(c.Dimension(), len(c.Fields))

	offset := 0
	for i, field := range c.Fields {
		termId, weight, err := field.TermWeight(recordFields[i])
		if err != nil {
			return v, err
		}
		componentIndex := offset + termId - 1
		v.Components[i] = SparseVectorComponent{componentIndex, weight}
		offset += field.TermCount
	}
	return v, nil
}

func (c *Classifier) Listen(records chan Record) {
	// MUTEX
	for r := range records {
		c.Classify(r)
	}
}
