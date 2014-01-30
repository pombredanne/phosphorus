package data

import (
	"log"
	"io/ioutil"
	"os"
	"testing"
	"willstclair.com/phosphorus/encoder"
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
)

var tempDir string

func init() {
	tempDir, _ := ioutil.TempDir("", "phosphorus_test")
	os.Chdir(tempDir)
	csv, _ := os.Create("csv")
	csv.WriteString(CSV)
	csv.Close()

	csv, _ = os.Create("csv2")
	csv.WriteString(CSV2)
	csv.Close()

	csv, _ = os.Create("csv3")
	csv.WriteString(CSV3)
	csv.Close()
}

func TestSlurp(t *testing.T) {
	d := NewData(tempDir, 2)

	var c encoder.Counter

	err := d.Slurp(func(records chan *Record) {
		for r := range records {
			c.Count(r.Fields)
		}
	})

	if err != nil {
		t.Error(err)
	}

	if c.Fields[0].Counts[c.Fields[0].Terms["ORANGE"]] != 2 {
		log.Println(c.Fields[0])
		t.Fail()
	}

	if c.Fields[1].Counts[c.Fields[1].Terms["3"]] != 3 {
		t.Fail()
	}
}
