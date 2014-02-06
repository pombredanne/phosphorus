package schema

import (
	"sort"
)

type Index interface {
	Write(*Record) error
	Query(map[string]string) ([]Result, error)
}

type Record struct {
	Id    uint32
	Attrs map[string]string
}

type Result struct {
	Record  *Record
	Matches int
}

type MemoryIndex struct {
	schema  *Schema
	ids     [][][]uint32
	records map[uint32]map[string]string
}

func (ix *MemoryIndex) put(i, j int, id uint32) {
	for len(ix.ids) <= i {
		ix.ids = append(ix.ids, [][]uint32{})
	}

	for len(ix.ids[i]) <= j {
		ix.ids[i] = append(ix.ids[i], []uint32{})
	}

	ix.ids[i][j] = append(ix.ids[i][j], id)
}
func (ix *MemoryIndex) Write(record *Record) (err error) {
	sigs, err := ix.schema.Sign(record.Attrs)
	if err != nil {
		return
	}

	ix.records[record.Id] = record.Attrs

	for i, sig := range sigs {
		ix.put(i, int(sig), record.Id)
	}
	return
}

type byMatches []Result

func (c byMatches) Len() int           { return len(c) }
func (c byMatches) Less(i, j int) bool { return c[i].Matches < c[j].Matches }
func (c byMatches) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func (ix *MemoryIndex) Query(record map[string]string) (results []Result, err error) {
	sigs, err := ix.schema.Sign(record)
	if err != nil {
		return
	}

	counter := make(map[uint32]int)
	for i, sig := range sigs {
		for _, id := range ix.ids[i][int(sig)] {
			counter[id]++
		}
	}

	for k, v := range counter {
		results = append(results, Result{
			&Record{k, ix.records[k]}, v})
	}

	sort.Sort(sort.Reverse(byMatches(results)))

	return
}

func NewMemoryIndex(s *Schema) Index {
	i := &MemoryIndex{schema: s}
	i.records = make(map[uint32]map[string]string)
	return i
}
