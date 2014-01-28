package lib

import (
	// "fmt"
	// "math"
	// "math/rand"
)

type Field struct {
	RecordCount int
	Terms       map[interface{}]int
	Counts      []int
}

func NewField() *Field {
	var f Field
	f.Terms = make(map[interface{}]int)
	f.Counts = make([]int, 0, 1024)
	return &f
}

func (f *Field) Add(term interface{}) {
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

func (c *Counter) Count(fields []interface{}) {
	for i, term := range fields {
		if len(c.Fields) <= i {
			c.Fields = append(c.Fields, NewField())
		}
		c.Fields[i].Add(term)
	}
}
