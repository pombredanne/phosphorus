package schema

import (
	"fmt"
	"strings"
)

var Transforms = []*Transform{
	xformSubstr,
	xformUpcase,
	xformSplit,
	xformTrim,
	xformKillAfter,
}

var xformSubstr = &Transform{
	Name:        "substr",
	Description: "(begin:int end:int) substring of a string",
	Instance:    xformSubstrF,
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

	tf = func(input []string) []string {
		for i, t := range input {
			input[i] = t[begin:end]
		}
		return input
	}

	return
}

var xformUpcase = &Transform{
	Name:        "upcase",
	Description: "make a string uppercase",
	Instance:    xformUpcaseF,
}

func xformUpcaseF(args map[string]interface{}) (tf TransformF, err error) {
	tf = func(input []string) []string {
		for i, t := range input {
			input[i] = strings.ToUpper(t)
		}
		return input
	}
	return
}

var xformTrim = &Transform{
	Name:        "trim",
	Description: "trim a string",
	Instance:    xformTrimF,
}

func xformTrimF(args map[string]interface{}) (tf TransformF, err error) {
	tf = func(input []string) []string {
		for i, t := range input {
			input[i] = strings.TrimSpace(t)
		}
		return input
	}
	return
}

var xformSplit = &Transform{
	Name:        "split",
	Description: "split a term",
	Instance:    xformSplitF,
}

func xformSplitF(args map[string]interface{}) (tf TransformF, err error) {
	tf = func(input []string) []string {
		out := make([]string, 0, len(input)*2)
		for _, t := range input {
			out = append(out, normalizeNames(t)...)
		}
		return out
	}
	return
}

var xformKillAfter = &Transform{
	Name:        "killafter",
	Description: "kill after char",
	Instance:    xformKillAfterF,
}

func killafter(s string, sep string) string {
	i := strings.Index(s, sep)
	if i == -1 {
		return s
	}
	return s[0:i]
}

func xformKillAfterF(args map[string]interface{}) (tf TransformF, err error) {
	sep, exists := args["sep"]
	if !exists {
		return nil, fmt.Errorf("missing arg for killafter")
	}

	tf = func(input []string) []string {
		for i, t := range input {
			input[i] = killafter(t, sep.(string))
		}
		return input
	}
	return
}

var prefixList = []string{"DE", "DEL", "LO", "MC", "MAC", "ST", "DU", "VAN", "SAINT", "D'", "L'", "O'", "LE", "LA", "VON", "O", "DI", "LI"}
var prefixSet = make(map[string]bool)

func init() {
	for _, p := range prefixList {
		prefixSet[p] = true
	}
}

func normalizeNames(name string) []string {
	out := []string{}
	sp := strings.Split(strings.Replace(strings.TrimSpace(name), "-", " ", -1), " ")
	spt := make([]string, 0, len(sp))
	for _, s := range sp {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		spt = append(spt, t)
	}

	start := 0
	for i := 0; i < len(spt); i++ {
		_, exists := prefixSet[spt[i]]
		if !exists {
			newS := strings.Join(spt[start:i+1], " ")
			out = append(out, newS)
			start = i + 1
		}
	}

	return out
}

// find a less unsavory name for this
func fuckJSON(d interface{}) int {
	switch d.(type) {
	case int:
		return d.(int)
	case float64:
		return int(d.(float64))
	default:
		panic("not an int")
	}
}
