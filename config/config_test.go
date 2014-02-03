package config

import (
	"bytes"
	"testing"
)

const JSON = `{"Source":{"SourceFields":[{"Column":2,"Name":"last_name","ShortName":"ln"},{"Column":3,"Name":"first_name","ShortName":"fn"},{"Column":4,"Name":"birth_year","ShortName":"by"},{"Column":5,"Name":"birth_month","ShortName":"bm"},{"Column":6,"Name":"birth_day","ShortName":"bd"}],"IdColumn":1,"S3":{"Bucket":"phosphorus","Prefix":"source/"},"Delimiter":"\t","Table":{"ReadCapacityUnits":10,"Name":"phosphorus-source","WriteCapacityUnits":10}},"Index":{"S3":{"Bucket":"phosphorus","Prefix":"index/"},"Table":{"ReadCapacityUnits":100,"Name":"phosphorus-index","WriteCapacityUnits":100},"IndexFields":[{"Names":["last_name"]},{"Names":["first_name"]},{"Names":["birth_year"]},{"Names":["birth_month","birth_day"]}]},"MaxProcs":3,"SecretAccessKey":"secret","AccessKeyId":"key"}`

func TestS3Validate(t *testing.T) {
	s := &S3{"bucketname", "prefix/"}
	err := s.Validate()
	if err != nil {
		t.Error(err)
	}

	s.Bucket = "x"
	err = s.Validate()
	if err == nil {
		t.Fail()
	}
}

func sourceFixture() Source {
	return Source{
		S3:       S3{"bucketname", "prefix/"},
		Table:    DynamoTable{"name", 10, 10},
		IdColumn: 1,
		SourceFields: []SourceField{
			SourceField{"first", 2, "f"},
			SourceField{"last", 3, "l"},
			SourceField{"birthdate", 4, "bd"}},
		Delimiter: ",",
	}
}

func TestSourceValidate(t *testing.T) {
	s := sourceFixture()
	err := s.Validate()
	if err != nil {
		t.Error(err)
	}
}

func TestDuplicateSourceFields(t *testing.T) {
	var s Source
	var err error

	s = sourceFixture()
	s.SourceFields[1].Name = "first"
	err = s.Validate()
	if err == nil {
		t.Fail()
	}

	s = sourceFixture()
	s.SourceFields[1].Column = 2 // duplicate cols are OK
	err = s.Validate()
	if err != nil {
		t.Error(err)
	}

	s = sourceFixture()
	s.SourceFields[1].ShortName = "f"
	err = s.Validate()
	if err == nil {
		t.Fail()
	}
}

func configFixture() (conf Configuration, err error) {
	reader := bytes.NewBufferString(JSON)
	err = conf.Load(reader)
	if err != nil {
		return
	}

	return
}

func TestInvalidIndexFields(t *testing.T) {
	c, err := configFixture()
	if err != nil {
		t.Error(err)
	}
	c.Index.IndexFields[0].Names[0] = "city"

	err = c.Validate()
	if err == nil {
		t.Fail()
	}
}
