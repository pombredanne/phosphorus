package environment

import (
	"github.com/crowdmob/goamz/dynamodb"
	"testing"
	"willstclair.com/phosphorus/schema"
)

var sig1 = []uint32{14, 255, 104, 172, 138, 51, 132, 248}
var sig2 = []uint32{14, 255, 104, 172, 138, 51, 232, 177}
var sig3 = []uint32{14, 255, 104, 197, 20, 149, 132, 62}

var rec1 = &schema.Record{1, map[string]string{"first": "John", "last": "Doe"}}
var rec2 = &schema.Record{2, map[string]string{"first": "Jane", "last": "Roe"}}

type _schema struct {
	fixture []uint32
}

func (s *_schema) Sign(map[string]string, schema.RandomProvider) ([]uint32, error) {
	return s.fixture, nil
}

func (s *_schema) SignatureLen() int {
	return 8
}

func (s *_schema) ChunkBits() int {
	return 8
}

type _random struct{}

func (r *_random) Get(int64) float64 {
	return 0.0
}

func TestBinkey(t *testing.T) {
	bk := binkey(255, 65535)
	if bk != "////" {
		t.Fatalf("%q != %q", bk, "////")
	}
}

func TestDynamoDBIndex(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip()
	// }

	s := &_schema{sig1}
	sourceT := getRandomTable()
	indexT := getRandomTable()
	ix := NewDynamoDBIndex(s, indexT, sourceT)
	r := &_random{}

	err := ix.Write(rec1, r)
	if err != nil {
		t.Error(err)
	}

	s.fixture = sig2
	err = ix.Write(rec2, r)
	if err != nil {
		t.Error(err)
	}

	ix.Flush()

	s.fixture = sig3

	results, err := ix.Query(map[string]string{}, r)
	if len(results) != 2 {
		t.Error("no results")
	}

	if results[0].Matches != 4 {
		t.Fail()
	}
	if results[0].Record.Id != 1 {
		t.Fail()
	}
	if results[0].Record.Attrs["first"] != "John" {
		t.Fail()
	}
}

func getTable(server *dynamodb.Server, name string) *dynamodb.Table {
	td, err := server.DescribeTable(name)
	if err != nil {
		panic(err)
	}

	pk, err := td.BuildPrimaryKey()
	if err != nil {
		panic(err)
	}

	return server.NewTable(name, pk)
}

func getRandomTable() *dynamodb.Table {
	return getTable(dynamo, createTable())
}
