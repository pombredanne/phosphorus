package schema

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
)

type Classifier interface {
	Learn(string)
	Signature(string, int) ([]float64, error)
	Dimension() int
}

type TfIdfClassifier struct {
	Counts  map[string]int
	dirty   bool
	total   int
	terms   []string
	seedgen *rand.Rand
	gens    []*rand.Rand
	hash    int64
	vector  [][]uint16
	seed    []int64
	lock    sync.RWMutex
}

func NewTfIdfClassifier() Classifier {
	c := &TfIdfClassifier{
		Counts: make(map[string]int)}

	return c
}

func (c *TfIdfClassifier) Dimension() int {
	return len(c.Counts)
}

func (c *TfIdfClassifier) Learn(term string) {
	if term == "" {
		return
	}
	c.dirty = true
	c.Counts[term]++
	return
}

func (c *TfIdfClassifier) weight(term string) float64 {
	return math.Log(float64(c.total) / float64(c.Counts[term]))
}

func (c *TfIdfClassifier) seedGen() *rand.Rand {
	if c.seedgen == nil {
		c.seedgen = rand.New(rand.NewSource(c.Hash()))
	}
	return c.seedgen
}

func (c *TfIdfClassifier) getGen(i int) *rand.Rand {
	for len(c.gens) <= i {
		r := rand.New(rand.NewSource(c.seedGen().Int63()))
		c.gens = append(c.gens, r)
	}

	return c.gens[i]
}

func (c *TfIdfClassifier) hashValue(i, j int) (v float64) {
	c.lock.Lock()
	for len(c.vector) <= i {
		c.vector = append(c.vector, []uint16{})
	}

	for len(c.vector[i]) <= j {
		c.vector[i] = append(c.vector[i], Compact(c.getGen(i).NormFloat64()))
	}
	c.lock.Unlock()

	c.lock.RLock()
	v = Uncompact(c.vector[i][j])
	c.lock.RUnlock()

	return
}

func (c *TfIdfClassifier) Signature(term string, n int) (s []float64, err error) {
	c.Clean()
	termIndex := sort.SearchStrings(c.terms, term)
	if termIndex == len(c.terms) {
		err = fmt.Errorf("Term not found: %s", term)
		return
	}

	termWeight := c.weight(term)

	for i := 0; i < n; i++ {
		multiplier := c.hashValue(i, termIndex)
		s = append(s, termWeight*multiplier)
	}

	return
}

func (c *TfIdfClassifier) Clean() {
	if !c.dirty && len(c.terms) > 0 {
		return
	}

	c.terms = make([]string, 0, len(c.Counts))
	c.total = 0

	for term, count := range c.Counts {
		c.terms = append(c.terms, term)
		c.total += count
	}

	sort.Strings(c.terms)
	return
}

func (c *TfIdfClassifier) genHash() {
	h := fnv.New64a()
	enc := gob.NewEncoder(h)

	for _, t := range c.terms {
		enc.Encode(fmt.Sprintf("%s%d", t, c.Counts[t]))
	}
	buf := bytes.NewBuffer(h.Sum([]byte{}))
	binary.Read(buf, binary.LittleEndian, &c.hash)
}

func (c *TfIdfClassifier) Hash() int64 {
	c.Clean()
	if c.hash == 0 {
		c.genHash()
	}
	return c.hash
}

func Compact(x float64) uint16 {
	return uint16(math.Floor((x + 8.0) * 4096.0))
}

func Uncompact(x uint16) float64 {
	return (float64(x) / 4096.0) - 8.0
}

func init() {
	gob.Register(&TfIdfClassifier{})
}
