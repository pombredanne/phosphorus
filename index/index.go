package index

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"willstclair.com/phosphorus/vector"
	"willstclair.com/phosphorus/environment"
)

type Signature [128]uint16

func (s *Signature) Key(i int) uint32 {
	return uint32(s[i]) | uint32(i<<16)
}

type Template struct {
	Directory string
	Dimension int
	family    [128][16]vector.Interface
}

func (t *Template) Sign(v vector.Interface) *Signature {
	var s Signature

	for i, h := range t.family {
		s[i] = vector.Signature(v, h[:])
	}

	return &s
}

func (t *Template) Load() {
	var wait sync.WaitGroup
	for i, _ := range t.family {
		i := i
		filename := fmt.Sprintf("hash_%02x", i)
		wait.Add(1)
		go func() {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("Error loading template: %s", err)
			}
			defer file.Close()

			dec := gob.NewDecoder(file)
			for j, _ := range t.family[i] {
				var v vector.HashVector
				err = dec.Decode(&v)
				if err != nil {
					log.Fatalf("Error loading template: %s", err)
				}
				t.family[i][j] = vector.Interface(v)
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

func (t *Template) Generate() {
	var wait sync.WaitGroup
	for i, _ := range t.family {
		filename := fmt.Sprintf("hash_%02x", i)
		wait.Add(1)
		go func() {
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal("Error generating template: %s", err)
			}
			defer file.Close()
			enc := gob.NewEncoder(file)
			for _ = range t.family[i] {
				err = enc.Encode(vector.Random(t.Dimension).(vector.HashVector))
				if err != nil {
					log.Fatal("Error generating template: %s", err)
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

type Index struct {
	entries   [1 << 23][]uint32
	threshold int
	writeChannel chan *environment.Item
}

func NewIndex(threshold int, writeChannel chan *environment.Item) (xr *Index) {
	xr = &Index{
		threshold: threshold,
		writeChannel: writeChannel,
	}

	for i := 0; i < 1<<23; i++ {
		xr.entries[i] = make([]uint32, 0, threshold)
	}

	return
}

func (xr *Index) Add(recordId uint32, s *Signature) {
	for si, v := range s {
		i := int(v) | (si << 16)
		xr.entries[i] = append(xr.entries[i], recordId)
		if len(xr.entries[i]) >= xr.threshold {
			xr.Flush(i)
		}
	}
}

func (xr *Index) Flush(i int) {
	if len(xr.entries[i]) > 0 {
		ids := xr.entries[i]
		xr.entries[i] = make([]uint32, 0, xr.threshold)
		xr.writeChannel <- environment.NewSetItem(uint32(i), ids)
	}
}

func (xr *Index) FlushAll() {
	log.Println("Flushing all writes")
	for i, e := range xr.entries {
		if len(e) > 0 {
			xr.Flush(i)
		}
	}
}

type Candidate struct {
	RecordId uint32
	Matches  int
}

type ByMatches []Candidate

func (c ByMatches) Len() int           { return len(c) }
func (c ByMatches) Less(i, j int) bool { return c[i].Matches < c[j].Matches }
func (c ByMatches) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func Rank(c chan uint32) (out []Candidate) {
	counter := make(map[uint32]int)

	for r := range c {
		counter[r]++
	}
	for k, v := range counter {
		out = append(out, Candidate{k, v})
	}
	sort.Sort(sort.Reverse(ByMatches(out)))

	return
}
