package data

import (
	"time"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"github.com/crowdmob/goamz/s3/s3test"
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

// func TestSlurp(t *testing.T) {
// 	d := NewData(tempDir, 2)

// 	var c encoder.Counter

// 	err := d.Slurp(func(records chan *Record) {
// 		for r := range records {
// 			c.Count(r.Fields)
// 		}
// 	})

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if c.Fields[0].Counts[c.Fields[0].Terms["ORANGE"]] != 2 {
// 		log.Println(c.Fields[0])
// 		t.Fail()
// 	}

// 	if c.Fields[1].Counts[c.Fields[1].Terms["3"]] != 3 {
// 		t.Fail()
// 	}
// }

func pie(in interface{}) (out interface{}, err error) {
	record := in.([]string)
	out = record[1] + " PIE"
	return
}

func TestFile(t *testing.T) {
	ch := make(chan interface{})

	f := File{
		Path: "csv",
		Mappers: []Mapper{pie},
		Stream: ch,
	}

	go func() {
		err := f.Load()
		if err != nil { t.Error(err) }
	}()

	expected := []string{"APPLE PIE", "APPLE PIE", " PIE", "ORANGE PIE"}

	count := 0
	for _, exp := range expected {
		actual := <-ch
		if actual.(string) != exp {
			t.FailNow()
		}
		count++
	}

	if count < len(expected) {
		t.Fail()
	}
}

func TestS3(t *testing.T) {
	s3server, err := s3test.NewServer(&s3test.Config{})
	if err != nil { panic(err) }

	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)
	token, err := aws.GetAuth("phosphorus-test", "secret", "", expires)
	if err != nil { panic(err) }

	region := aws.Region{
		Name: "test",
		S3Endpoint: s3server.URL(),
		S3LocationConstraint: true }

	mys3 := s3.New(token, region)
	log.Println(mys3)
	bucket := mys3.Bucket("mybucket")
	err = bucket.PutBucket(s3.Private)
	if err != nil { panic(err) }
	log.Println(bucket)


	buf := bytes.NewBufferString("test")
	err = bucket.Put("files/test.dat", buf.Bytes(), "text/plain",
		s3.Private, s3.Options{})

	if err != nil { panic(err) }
	data, err := bucket.Get("files/test.dat")
	if err != nil { panic(err) }

	log.Println(bytes.NewBuffer(data).String())


	t.Fail()
}
