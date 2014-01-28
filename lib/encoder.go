package lib

import (
	"math"
)

func Compact(x float64) uint16 {
	return uint16(math.Floor((x + 8.0) * 4096.0))
}

func Uncompact(x uint16) float64 {
	return (float64(x) / 4096.0) - 8.0
}


type Encoder struct {
	Dimension int
	Terms     []map[interface{}]int
	Weights   [][]float64
}

func NewEncoder (c *Counter) *Encoder {
	var e Encoder
	e.Weights = make([][]float64, len(c.Fields))
	e.Terms = make([]map[interface{}]int, len(c.Fields))

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

func (e *Encoder) Encode (fields []interface{}) *SparseVector {
	var v SparseVector = SparseVector{
		e.Dimension,
		make([]SparseVectorComponent, len(fields))}

	offset := 0
	for i, field := range fields {
		termId := e.Terms[i][field]
		weight := e.Weights[i][termId]
		v.Components[i] = SparseVectorComponent{offset + termId, weight}
		offset += len(e.Weights[i])
	}

	return &v
}
