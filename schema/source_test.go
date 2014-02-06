package schema

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

const (
	CSV = `
132794898,APPLE,1
4291953712,APPLE,
2919886652,,1
1637706266,ORANGE,2
`
	CSV2 = `
2706566389,BANANA,4
4188369442,PEAR,3
3153444041,APRICOT,3
`
	CSV3 = `
234594352,ORANGE,3
`
	FILESOURCE = `
{
  "id_column": 1,
  "delimiter": ",",
  "concurrent": 2,
  "fields": [
    {
      "name": "count",
      "column": 3
    },
    {
      "name": "fruit",
      "column": 2
    }
  ]
}
`
)

var expectedFruits = []string{"", "APPLE", "APPLE", "APRICOT", "BANANA", "ORANGE", "ORANGE", "PEAR"}
var expectedCounts = []string{"", "1", "1", "2", "3", "3", "3", "4"}

func dump(name, contents string) {
	csv, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	csv.WriteString(contents)
	csv.Close()
}

func testdir() string {
	tempdir, err := ioutil.TempDir("", "phosphorus")
	if err != nil {
		panic(err)
	}

	dump(filepath.Join(tempdir, "csv"), CSV)
	dump(filepath.Join(tempdir, "csv2"), CSV2)
	dump(filepath.Join(tempdir, "csv3"), CSV3)
	return tempdir
}

func TestSourceFile(t *testing.T) {
	tempdir := testdir()
	defer os.RemoveAll(tempdir)

	s := &FileSource{}
	err := json.Unmarshal([]byte(FILESOURCE), &s)
	if err != nil {
		t.Error(err)
	}
	s.Glob = filepath.Join(tempdir, "*")

	c, err := s.GetChannel()
	if err != nil {
		t.Error(err)
	}

	actualFruits := []string{}
	actualCounts := []string{}

	for r := range c {
		actualFruits = append(actualFruits, r.Attrs["fruit"])
		actualCounts = append(actualCounts, r.Attrs["count"])
	}

	sort.Strings(actualFruits)
	sort.Strings(actualCounts)

	if !reflect.DeepEqual(expectedFruits, actualFruits) {
		log.Println(expectedFruits, actualFruits)
		t.Fail()
	}

	if !reflect.DeepEqual(expectedCounts, actualCounts) {
		log.Println(expectedCounts, actualCounts)
		t.Fail()
	}
}
