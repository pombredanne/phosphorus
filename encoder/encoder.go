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

package encoder

import (
	"encoding/gob"
	"github.com/wsc/phosphorus/vector"
	"math"
	"os"
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
	if term == "" {
		return
	}
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
		fo := offset
		offset += len(e.Weights[i])
		if field == "" {
			continue
		}
		termId := e.Terms[i][field]
		weight := e.Weights[i][termId]
		v.Components[i] = vector.SparseVectorComponent{fo + termId, weight}
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
