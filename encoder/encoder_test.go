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
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"testing"
)

var records = [][]string{
	[]string{"a", "1", ""},
	[]string{"b", "1", ""},
	[]string{"c", "2", "x"}}

var f *Field
var c Counter

func init() {
	f = NewField()
	f.Add("apple")
	f.Add("apple")
	f.Add("orange")

	for _, record := range records {
		c.Count(record)
	}
}

func TestFieldRecordCount(t *testing.T) {
	if f.RecordCount != 3 {
		t.Fail()
	}
}

func TestFieldTermId(t *testing.T) {
	if f.Terms["apple"] != 0 {
		t.Fail()
	}
}

func TestFieldCounts(t *testing.T) {
	if f.Counts[0] != 2 {
		t.Fail()
	}
}

func TestFieldCountResizing(t *testing.T) {
	for i := 0; i < 10000; i++ {
		f.Add(string(i))
	}
}

func TestCounterTerms(t *testing.T) {
	if c.Fields[0].Terms["c"] != 2 {
		t.Fail()
	}

	if c.Fields[1].Terms["2"] != 1 {
		t.Fail()
	}
}

func TestCounterCounts(t *testing.T) {
	if !reflect.DeepEqual(c.Fields[0].Counts, []int{1, 1, 1}) {
		t.Fail()
	}
	if !reflect.DeepEqual(c.Fields[1].Counts, []int{2, 1}) {
		t.Fail()
	}
}

func TestEncoder(t *testing.T) {
	e := NewEncoder(&c)
	v := e.Encode(records[0])

	if math.Abs(1.0986122886681096-v.Component(0)) > 0.00001 {
		t.Fail()
	}

	if math.Abs(0.4054651081081644-v.Component(3)) > 0.00001 {
		t.Fail()
	}
}

func TestPersistEncoder(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "phosphorus_test")
	if err != nil {
		panic(err)
	}
	defer func() { os.RemoveAll(tempDir) }()
	err = os.Chdir(tempDir)
	if err != nil {
		panic(err)
	}

	e := NewEncoder(&c)
	e.Path = "encoder"
	e.Save()

	e2 := Encoder{Path: "encoder"}
	e2.Load()

	v := e2.Encode(records[0])
	if math.Abs(1.0986122886681096-v.Component(0)) > 0.00001 {
		t.Fail()
	}

	if math.Abs(0.4054651081081644-v.Component(3)) > 0.00001 {
		t.Fail()
	}
}

func TestDoNotCountEmptyFields(t *testing.T) {
	e := NewEncoder(&c)
	for _, termMap := range e.Terms {
		_, exists := termMap[""]
		if exists {
			t.Fail()
		}
	}
}
