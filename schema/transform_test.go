package schema

import (
	"encoding/json"
	"reflect"
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

	if xf([]string{"apple"})[0] != "app" {
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

	if ti.Fn([]string{"apple"})[0] != "app" {
		t.Fail()
	}
}

type _c struct {
	input    string
	expected []string
}

var testNames = []_c{
	_c{"ST CLAIR", []string{"ST CLAIR"}},
	_c{"ST CLAIR-JONES", []string{"ST CLAIR", "JONES"}},
	_c{"DEL RAY", []string{"DEL RAY"}},
	_c{"DUFF HESTER", []string{"DUFF", "HESTER"}},
	_c{"ROSADO DE GRACIA", []string{"ROSADO", "DE GRACIA"}},
	_c{"CONNOLLY-MC LEISH", []string{"CONNOLLY", "MC LEISH"}},
	_c{"DU BRUCQ", []string{"DU BRUCQ"}},
	_c{"VAN NUYS- CRUZ", []string{"VAN NUYS", "CRUZ"}},
	_c{"OSBORNE - BARTON", []string{"OSBORNE", "BARTON"}},
	_c{"SAINT VIL-COACHY", []string{"SAINT VIL", "COACHY"}},
	_c{"SIMO D' OLEO", []string{"SIMO", "D' OLEO"}},
	_c{"DE LA RENTA", []string{"DE LA RENTA"}},
	_c{"DEL-PILAR", []string{"DEL PILAR"}},
	// _c{"", []string{"", ""}},
}

func TestSplitNames(t *testing.T) {
	for _, pair := range testNames {
		actual := normalizeNames(pair.input)
		if !reflect.DeepEqual(actual, pair.expected) {
			t.Fail()
			t.Logf("%s != %s", actual, pair.expected)
		}
	}
}

// ugh todo another time: strip JR, SR, III, etc etc
