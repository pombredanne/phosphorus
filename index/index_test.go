package index

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"willstclair.com/phosphorus/environment"
	"willstclair.com/phosphorus/vector"
)

func TestSignature(t *testing.T) {
	var s Signature
	s[127] = 0xbeef
	if s.Key(127) != 0x7fbeef {
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

func fakeSig() *Signature {
	var sig Signature
	for i := 0; i < 128; i += 2 {
		sig[i] = 0xcafe
		sig[i+1] = 0xbeef
	}
	return &sig
}

func TestIndexManualFlush(t *testing.T) {
	wc := make(chan *environment.Item)
	sig := fakeSig()
	xr := NewIndex(2, wc)

	xr.Add(0xbeefcafe, sig)
	go xr.Flush(0x7fbeef)

	item := <-wc

	if environment.Dec64(item.Key.HashKey) != 0x7fbeef {
		t.Fail()
	}

	if environment.Dec64(item.Attributes[0].SetValues[0]) != 0xbeefcafe {
		t.Fail()
	}
}

func TestIndexAutoFlush(t *testing.T) {
	wc := make(chan *environment.Item)
	sig := fakeSig()
	xr := NewIndex(2, wc)
	xr.Add(0xdeadbeef, sig)
	go xr.Add(0xc0ffee, sig)

	// time.Sleep(500 * time.Millisecond) // gross

	item := <-wc
	if len(item.Attributes[0].SetValues) != 2 {
		t.Fail()
	}
}

func TestRank(t *testing.T) {
	c := make(chan uint32)

	go func() {
		for i := 0; i < 4; i++ {
			c <- 0xdeadbeef
			if i%2 == 0 {
				c <- 0xc0ffee
			}
		}
		close(c)
	}()

	candidates := Rank(c)

	if candidates[0].Matches != 4 {
		t.Fail()
	}

	if candidates[0].RecordId != 0xdeadbeef {
		t.Fail()
	}

	if candidates[1].Matches != 2 {
		t.Fail()
	}

	if candidates[1].RecordId != 0xc0ffee {
		t.Fail()
	}
}
