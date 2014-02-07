package schema

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
)

type TransformF func(string) string

type Transform struct {
	Name        string
	Description string
	Instance    func(map[string]interface{}) (TransformF, error)
}

type TransformI struct {
	Name      string                 `json:"function"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Fn        TransformF             `json:"-"`
}

func (ti *TransformI) hydrate() (err error) {
	for _, xf := range Transforms {
		if ti.Name == xf.Name {
			ti.Fn, err = xf.Instance(ti.Arguments)
			return
		}
	}
	err = fmt.Errorf("function not found: %s", ti.Name)
	return
}

type Field struct {
	Comment    string        `json:"comment"`
	Attrs      []string      `json:"attrs"`
	Transforms []*TransformI `json:"transforms"`
	Classifier Classifier    `json:"-"`
}

func (d *Field) hydrate() (err error) {
	if d.Classifier == nil {
		d.Classifier = NewTfIdfClassifier()
	}
	for _, t := range d.Transforms {
		err = t.hydrate()
	}
	return
}

func (d *Field) Load(data []byte) (err error) {
	err = json.Unmarshal(data, &d)
	if err != nil {
		return
	}

	d.hydrate()

	return
}

func (d *Field) xform(term string) (out string) {
	out = term
	for _, t := range d.Transforms {
		out = t.Fn(out)
	}
	return
}

func (d *Field) pick(record map[string]string) string {
	var buf bytes.Buffer
	for _, attr := range d.Attrs {
		buf.WriteString(record[attr])
	}
	return d.xform(buf.String())
}

func (d *Field) Learn(record map[string]string) {
	d.Classifier.Learn(d.pick(record))
}

func (d *Field) Signature(record map[string]string, n int) (s []float64, err error) {
	s, err = d.Classifier.Signature(d.pick(record), n)
	return
}

type Schema struct {
	HashCount int      `json:"hash_count"`
	Width     int      `json:"chunk_size"`
	Fields    []*Field `json:"fields"`
}

func (s *Schema) SignatureLen() int {
	return s.HashCount / s.Width
}

func (s *Schema) ChunkBits() int {
	return s.Width
}

func (s *Schema) LoadJSON(data []byte) (err error) {
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}
	s.hydrate()
	return
}

func (s *Schema) Hyd() {
	s.hydrate()
}

func (s *Schema) hydrate() {
	for _, d := range s.Fields {
		d.hydrate()
	}
}

func (s *Schema) Learn(c chan map[string]string) {
	for record := range c {
		for _, d := range s.Fields {
			d.Learn(record)
		}
	}
}

func (s *Schema) LearnRecords(c chan *Record) {
	for record := range c {
		for _, d := range s.Fields {
			d.Learn(record.Attrs)
		}
	}
}

func (s *Schema) Sign(record map[string]string) ([]uint32, error) {
	var raw [][]float64
	var signatures []uint32

	for _, d := range s.Fields {
		sig, err := d.Signature(record, s.HashCount)
		if err != nil {
			return nil, err
		}
		raw = append(raw, sig)
	}

	chunks := s.HashCount / s.Width
	for i := 0; i < chunks; i++ {
		var chunk uint32
		for j := 0; j < s.Width; j++ {
			sum := 0.0
			for _, v := range raw {
				sum += v[(i*s.Width)+j]
			}
			if sum >= 0.0 {
				chunk |= (1 << uint(j))
			}
		}

		signatures = append(signatures, chunk)
	}
	return signatures, nil
}

func (s *Schema) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	err = enc.Encode(s)
	return
}

func (s *Schema) Load(r io.Reader) (err error) {
	dec := gob.NewDecoder(r)
	err = dec.Decode(s)
	s.hydrate()
	return
}
