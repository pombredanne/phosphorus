package encoder

import (
	"encoding/gob"
	"math"
	"os"
	"willstclair.com/phosphorus/vector"
)

type Field struct {
	RecordCount int
	Terms       map[string]int
	Counts      []int
}

func NewField() *Field {
	var f Field
	f.Terms = make(map[string]int)
	f.Counts = make([]int, 0, 1024)
	return &f
}

func (f *Field) Add(term string) {
	termId, seen := f.Terms[term]
	if !seen {
		f.Counts = append(f.Counts, 1)
		termId = len(f.Counts) - 1
		f.Terms[term] = termId
	} else {
		f.Counts[termId]++
	}
	f.RecordCount++
}

type Counter struct {
	Fields []*Field
}

func (c *Counter) Count(fields []string) {
	for i, term := range fields {
		if len(c.Fields) <= i {
			c.Fields = append(c.Fields, NewField())
		}
		c.Fields[i].Add(term)
	}
}

type Encoder struct {
	Path      string
	Dimension int
	Terms     []map[string]int
	Weights   [][]float64
}

func NewEncoder(c *Counter) *Encoder {
	var e Encoder
	e.Weights = make([][]float64, len(c.Fields))
	e.Terms = make([]map[string]int, len(c.Fields))

	for i, f := range c.Fields {
		e.Dimension += len(f.Counts)
		e.Terms[i] = f.Terms
		e.Weights[i] = make([]float64, len(f.Counts))
		for j, n := range f.Counts {
			e.Weights[i][j] = math.Log(float64(f.RecordCount) / float64(n))
		}
	}

	return &e
}

func (e *Encoder) Encode(fields []string) vector.Interface {
	v := vector.NewSparseVector(e.Dimension, len(fields))

	offset := 0
	for i, field := range fields {
		termId := e.Terms[i][field]
		weight := e.Weights[i][termId]
		v.Components[i] = vector.SparseVectorComponent{offset + termId, weight}
		offset += len(e.Weights[i])
	}

	return vector.Interface(v)
}

func (e *Encoder) Save() error {
	file, err := os.Create(e.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(e)
}

func (e *Encoder) Load() error {
	file, err := os.Open(e.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewDecoder(file).Decode(e)
}
