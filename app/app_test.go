package app

import (
	"github.com/crowdmob/goamz/dynamodb"
	"testing"
)

func TestDicks(t *testing.T) {
	j := &Job{
		IndexId:  900,
		Id:       5028,
		Type:     "cat",
		Argument: "no thx",
		State:    JOB_LOCK,
	}

	k, attrs := dynamo(j)

	if k.HashKey != "900" {
		t.Fail()
	}
	if k.RangeKey != "5028" {
		t.Fail()
	}
	if attrs[2].Type != dynamodb.TYPE_NUMBER {
		t.Fail()
	}
}

func TestSchlongs(t *testing.T) {
	// k := &dynamodb.Key{"900", "5028"}
	attrMap := map[string]*dynamodb.Attribute{
		"type": &dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "type",
			Value: "sure"},
		"argument": &dynamodb.Attribute{
			Type:  dynamodb.TYPE_STRING,
			Name:  "argument",
			Value: "yep"},
		"state": &dynamodb.Attribute{
			Type:  dynamodb.TYPE_NUMBER,
			Name:  "state",
			Value: "2"}}

	t.Log("8===D")

	j := Job{IndexId: 900, Id: 5028}
	attrs2struct(&j, attrMap)

	t.Log(j)
	t.Fail()
}
