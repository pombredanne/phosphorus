package schema

import (
	// "log"
	"fmt"
	"strings"
)

var Transforms = []*Transform{
	xformSubstr,
	xformUpcase,
}

var xformSubstr = &Transform{
	Name: "substr",
	Description: "(begin:int end:int) substring of a string",
	Instance: xformSubstrF,
}

func xformSubstrF(args map[string]interface{}) (tf TransformF, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid args")
		}
	}()

	b, exists := args["begin"]
	if !exists {
		err = fmt.Errorf("begin is required")
		return
	}
	begin := fuckJSON(b)

	e, exists := args["end"]
	if !exists {
		err = fmt.Errorf("end is required")
		return
	}
	end := fuckJSON(e)

	tf = func(input string) string {
		return input[begin:end]}

	return
}

var xformUpcase = &Transform{
	Name: "upcase",
	Description: "make a string upper case",
	Instance: xformUpcaseF,
}

func xformUpcaseF(args map[string]interface{}) (tf TransformF, err error) {
	tf = func(input string) string {
		return strings.ToUpper(input)}
	return
}

// find a less unsavory name for this
func fuckJSON(d interface{}) (int) {
	switch d.(type) {
	case int:
		return d.(int)
	case float64:
		return int(d.(float64))
	default:
		panic("not an int")
	}
}
