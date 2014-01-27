package classifier

import (
	"fmt"
	"math"
	"willstclair.com/phosphorus/vector"
)

type Field struct {
	Initialized bool
	TermCount   int
	RecordCount int
	Terms       map[interface{}]int
	Counts      []int
}

func (f *Field) Initialize () {
	if !f.Initialized {
		f.Terms = make(map[interface{}]int)
		f.Counts = make([]int, 1, 1000)
		f.Initialized = true
	}
}

func (f *Field) Quantize (term interface{}) (int, error) {
	if !f.Initialized { return 0, fmt.Errorf("Field not Initialized") }
	term_id := f.Terms[term]
	if term_id == 0 {
		return 0, fmt.Errorf("Term not found: %q", term)
	}
	return term_id, nil
}

func (f *Field) Add (term interface{}) {
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

func (f *Field) Weight (termId int) (float64, error) {
	if !f.Initialized { return 0.0, fmt.Errorf("Field not Initialized") }
	return math.Log(float64(f.RecordCount) / float64(f.Counts[termId])), nil
}

func (f *Field) TermWeight (term interface{}) (int, float64, error) {
	termId, err := f.Quantize(term)
	if err != nil { return termId, 0.0, err }

	weight, err := f.Weight(termId)
	return termId, weight, err
}

type Record interface {
	Fields() []interface{}
}

type Classifier struct {
	Fields []Field
}

func (c *Classifier) Classify (r Record) {
	for i, term := range r.Fields() {
		if len(c.Fields) <= i {
			c.Fields = append(c.Fields, *new(Field))
		}
		c.Fields[i].Add(term)
	}
}

func (c *Classifier) Dimension () int {
	dimensionCount := 0
	for _, field := range c.Fields {
		dimensionCount += field.TermCount
	}
	return dimensionCount
}

func (c *Classifier) Vector (r Record) (*vector.SparseVector, error) {
	recordFields := r.Fields()
	v := vector.NewSparseVector(c.Dimension(), len(c.Fields))

	offset := 0
	for i, field := range c.Fields {
		termId, weight, err := field.TermWeight(recordFields[i])
		if err != nil { return v, err }
		componentIndex := offset + termId - 1
		v.Components[i] = vector.SparseVectorComponent{componentIndex, weight}
		offset += field.TermCount
	}
	return v, nil
}

func (c *Classifier) Listen (records chan Record) {
	// MUTEX
	for r := range records { c.Classify(r) }
}
