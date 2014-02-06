package schema

import (
	"encoding/json"
	"testing"
)

const JS = `{"function":"substr","arguments":{"begin":0,"end":3}}`

func TestTransform(t *testing.T) {
	var args = map[string]interface{}{
		"begin": 0,
		"end":   3}

	xf, err := xformSubstr.Instance(args)
	if err != nil {
		t.Error(err)
	}

	if xf("apple") != "app" {
		t.Fail()
	}
}

func TestTransformSerialize(t *testing.T) {
	var args = map[string]interface{}{
		"begin": 0,
		"end":   3}

	fn, err := xformSubstr.Instance(args)
	if err != nil {
		t.Error(err)
	}

	tf := &TransformI{
		Name:      "substr",
		Arguments: args,
		Fn:        fn}

	js, err := json.Marshal(tf)
	if err != nil {
		t.Error(err)
	}

	// log.Println(string(js))

	if string(js) != JS {
		t.Fail()
	}
}

func TestTransformDeserialize(t *testing.T) {
	var ti = &TransformI{}
	err := json.Unmarshal([]byte(JS), &ti)
	if err != nil {
		t.Error(err)
	}

	ti.hydrate()

	if ti.Fn("apple") != "app" {
		t.Fail()
	}
}
