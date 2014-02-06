package schema

import (
	"bytes"
	"reflect"
	"testing"
)

var field = &Field{
	Comment: "myfield",
	Attrs:   []string{"name"},
	Transforms: []*TransformI{
		&TransformI{Name: "upcase"}},
	Classifier: NewTfIdfClassifier()}

var animal = &Field{
	Comment: "animals",
	Attrs:   []string{"animal"},
	Transforms: []*TransformI{
		&TransformI{Name: "upcase"}},
	Classifier: NewTfIdfClassifier()}

var schema = &Schema{
	HashCount: 2048,
	Width:     12,
	Fields:    []*Field{field, animal}}

var fieldJs = `{"comment":"myfield","attrs":["name"],"transforms":[{"function":"substr","arguments":{"begin":0,"end":3}},{"function":"upcase"}]}`

var schemaJs = `{"hash_count":2048,"chunk_size":16,"fields":[{"comment":"myfield","attrs":["name"],"transforms":[{"function":"substr","arguments":{"begin":0,"end":3}},{"function":"upcase"}]}]}`

func TestLoadField(t *testing.T) {
	f := &Field{}
	err := f.Load([]byte(fieldJs))
	if err != nil {
		t.Error(err)
	}

	if f.Comment != "myfield" {
		t.Fail()
	}
}

func TestLoadSchema(t *testing.T) {
	s := &Schema{}
	err := s.LoadJSON([]byte(schemaJs))
	if err != nil {
		t.Error(err)
	}

	if s.HashCount != 2048 {
		t.Fail()
	}
}

func TestPersistSchema(t *testing.T) {
	schema.hydrate()

	c := make(chan map[string]string, 5)
	fruits := []string{"apple", "orange", "banana", "apple", "pear"}
	animals := []string{"dog", "cat", "mouse", "cat", "bird"}
	appleDog := map[string]string{"name": "apple", "animal": "dog"}
	for i := 0; i < 5; i++ {
		c <- map[string]string{"name": fruits[i], "animal": animals[i]}
	}
	close(c)

	schema.Learn(c)
	sig1, _ := schema.Sign(appleDog)

	buf := &bytes.Buffer{}
	err := schema.Save(buf)
	if err != nil {
		t.Error(err)
	}

	s2 := &Schema{}
	err = s2.Load(buf)
	if err != nil {
		t.Error(err)
	}

	sig2, err := s2.Sign(appleDog)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(sig1, sig2) {
		t.Fail()
	}
}