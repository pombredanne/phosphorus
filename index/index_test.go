package index

import (
	"encoding/gob"
	"fmt"
	"github.com/crowdmob/goamz/dynamodb"
	"io/ioutil"
	"os"
	"testing"
	"time"
	"willstclair.com/phosphorus/vector"
)

var server *dynamodb.Server
var table *dynamodb.Table

func init() {
	server = NewServer("phosphorustest", "secret")
	cycleTable()
}

func cycleTable() {
	server.DeleteTable(SignatureTableDescription)
	server.CreateTable(SignatureTableDescription)
	pk, _ := SignatureTableDescription.BuildPrimaryKey()
	table = server.NewTable("signature", pk)
}

func TestSignature(t *testing.T) {
	var s Signature
	s[127] = 0xbeef
	if s.Key(127) != 0x7fbeef {
		t.Fail()
	}
}

func TestSignatureKeys(t *testing.T) {
	var s Signature
	s[127] = 0xbeef
	if s.Keys64()[127] != "f77v" {
		t.Fail()
	}
}

func generateZeroTemplate() string {
	tempDir, err := ioutil.TempDir("", "phosphorus_test")
	if err != nil {
		panic(err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 128; i++ {
		filename := fmt.Sprintf("hash_%02x", i)
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		enc := gob.NewEncoder(file)
		for j := 0; j < 16; j++ {
			v := make(vector.HashVector, 4)
			err = enc.Encode(v)
			if err != nil {
				panic(err)
			}
		}
	}
	return tempDir
}

// this is a little duplicatey
func TestTemplateLoad(t *testing.T) {
	tempDir := generateZeroTemplate()
	defer func() { os.RemoveAll(tempDir) }()

	template := Template{
		Directory: tempDir,
		Dimension: 4,
	}
	template.Load()

	for i := 0; i < 128; i++ {
		for _, v := range template.family[i] {
			for j := 0; j < 4; j++ {
				value := v.Component(j)
				if value != -8.0 {
					t.Errorf("Expected -8.0, not %f", value)
				}
			}
		}
	}
}

// mostly just a sanity check
func TestTemplateGenerate(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "phosphorus_test")
	if err != nil {
		panic(err)
	}
	defer func() { os.RemoveAll(tempDir) }()
	err = os.Chdir(tempDir)
	if err != nil {
		panic(err)
	}

	template := Template{
		Directory: tempDir,
		Dimension: 4,
	}

	template.Generate()
	template.Load()

	for i := 0; i < 128; i++ {
		for _, v := range template.family[i] {
			for j := 0; j < 4; j++ {
				value := v.Component(j)
				if value < -8.0 || value > 8.0 {
					t.Errorf("Unexpected value: %f", value)
				}
			}
		}
	}
}

// again, more of a sanity check than anything
func TestTemplateSign(t *testing.T) {
	tempDir := generateZeroTemplate()
	defer func() { os.RemoveAll(tempDir) }()
	template := Template{
		Directory: tempDir,
		Dimension: 4,
	}
	template.Load()

	v0 := vector.Vector{-1.0, -1.0, 1.0, -1.0}
	v1 := vector.Vector{2.0, -1.0, 2.0, -1.0}

	s := template.Sign(v0)
	for _, i := range s {
		if i != 65535 {
			t.Fail()
		}
	}

	s = template.Sign(v1)
	for _, i := range s {
		if i != 0 {
			t.Fail()
		}
	}

	template.Generate()
	template.Load()
	s = template.Sign(v1)

	for _, i := range s {
		if i == 0 || i == 65535 {
			continue
		}
		return
	}
	t.Fail()
}

func TestSaveEntry(t *testing.T) {
	cycleTable()

	e := Entry{0x7fbeef, []uint32{0x00c0ffee}}
	err := e.Save(table)
	if err != nil {
		t.Error(err)
	}

	response, _ := table.GetItem(&dynamodb.Key{HashKey: "f77v"})
	if response["i"].SetValues[0] != "AMD/7g==" {
		t.Fail()
	}
}

func fakeSig() *Signature {
	var sig Signature
	for i := 0; i < 128; i += 2 {
		sig[i] = 0xcafe
		sig[i+1] = 0xbeef
	}
	return &sig
}

func TestIndexManualFlush(t *testing.T) {
	cycleTable()

	sig := fakeSig()
	xr := NewIndex(2, table)

	xr.Add(0xbeefcafe, sig)
	xr.Flush(0x7fbeef)
	response, _ := table.GetItem(&dynamodb.Key{HashKey: "f77v"})
	if len(response["i"].SetValues) != 1 {
		t.Fail()
	}
	time.Sleep(1 * time.Millisecond)
}

func TestIndexAutoFlush(t *testing.T) {
	cycleTable()

	sig := fakeSig()
	xr := NewIndex(2, table)
	xr.Add(0xdeadbeef, sig)
	xr.Add(0x00c0ffee, sig)

	time.Sleep(500 * time.Millisecond) // gross

	response, _ := table.GetItem(&dynamodb.Key{HashKey: "f77v"})
	if len(response["i"].SetValues) != 2 {
		t.Fail()
	}
}

func TestIndexQuery(t *testing.T) {
	cycleTable()

	var sig0 Signature
	sig1 := fakeSig()
	sig2 := fakeSig()
	sig2[0] = 0

	xr := NewIndex(2, table)
	xr.Add(0xdeadbeef, sig1)
	xr.Add(0x00c0ffee, sig2)

	xr.FlushAll()

	x := NewIndex(2, table)
	candidates := x.Query(sig1)
	if candidates[0].Matches != 128 {
		t.Fail()
	}
	if candidates[1].Matches != 127 {
		t.Fail()
	}

	candidates = x.Query(&sig0)
	if candidates[0].Matches != 1 {
		t.Fail()
	}
}
