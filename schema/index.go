package schema

import (
	"sort"
	"sync"
)

type Signer interface {
	Sign(map[string]string, RandomProvider) ([]uint32, error)
	SignatureLen() int
	ChunkBits() int
}

type Index interface {
	Write(*Record, RandomProvider) error
	Query(map[string]string, RandomProvider) ([]Result, error)
	Flush() error
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
	signer      Signer
	ids         [][][]uint32
	idsLock     sync.RWMutex
	records     map[uint32]map[string]string
	recordsLock sync.RWMutex
}

func (ix *MemoryIndex) put(i, j int, id uint32) {
	ix.idsLock.Lock()
	for len(ix.ids) <= i {
		ix.ids = append(ix.ids, [][]uint32{})
	}

	for len(ix.ids[i]) <= j {
		ix.ids[i] = append(ix.ids[i], []uint32{})
	}

	ix.ids[i][j] = append(ix.ids[i][j], id)
	ix.idsLock.Unlock()
}

func (ix *MemoryIndex) Write(record *Record, r RandomProvider) (err error) {
	sigs, err := ix.signer.Sign(record.Attrs, r)
	if err != nil {
		return
	}

	ix.recordsLock.Lock()
	ix.records[record.Id] = record.Attrs
	ix.recordsLock.Unlock()

	for i, sig := range sigs {
		ix.put(i, int(sig), record.Id)
	}
	return
}

func (ix *MemoryIndex) Flush() error {
	return nil
}

type ByMatches []Result

func (c ByMatches) Len() int           { return len(c) }
func (c ByMatches) Less(i, j int) bool { return c[i].Matches < c[j].Matches }
func (c ByMatches) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func (ix *MemoryIndex) Query(record map[string]string, r RandomProvider) (results []Result, err error) {
	sigs, err := ix.signer.Sign(record, r)
	if err != nil {
		return
	}

	counter := make(map[uint32]int)
	ix.idsLock.RLock()
	for i, sig := range sigs {
		for _, id := range ix.ids[i][int(sig)] {
			counter[id]++
		}
	}
	ix.idsLock.RUnlock()

	for k, v := range counter {
		ix.recordsLock.RLock()
		attrs := ix.records[k]
		ix.recordsLock.RUnlock()
		results = append(results, Result{
			&Record{k, attrs}, v})
	}

	sort.Sort(sort.Reverse(ByMatches(results)))

	return
}

func NewMemoryIndex(s Signer) Index {
	ix := &MemoryIndex{signer: s}
	ix.records = make(map[uint32]map[string]string)

	ix.ids = make([][][]uint32, s.SignatureLen())
	for i, _ := range ix.ids {
		ix.ids[i] = make([][]uint32, 1<<uint(s.ChunkBits()))
	}

	return ix
}
