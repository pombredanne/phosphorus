package config

import (
	"encoding/json"
	"io"
	"fmt"
	"strings"
)

// the boilerplate in here seems gratuitous and i'm sure there's some
// way to clean it up, but leaving that for another time

type Error struct {
	init bool
	messages map[string]bool
}

func (e *Error) Error() string {
	return strings.Join(e.Messages(), "; ")
}

func (e *Error) Add(s string) {
	if !e.init {
		e.messages = make(map[string]bool)
	}
	e.messages[s] = true
}

func (e *Error) Merge(f error) {
	for message, _ := range f.(*Error).messages {
		e.Add(message)
	}
}

func (e *Error) Messages() (msgs []string) {
	for m, _ := range e.messages {
		msgs = append(msgs, m)
	}
	return
}

type Configuration struct {
	MaxProcs int
	AWSRegion string
	AccessKeyId string
	SecretAccessKey string
	Source Source
	Index Index
}

func (c *Configuration) Load(r io.Reader) (err error) {
	var conf Configuration
	err = json.NewDecoder(r).Decode(&conf)
	if err != nil {
		return
	}

	err = conf.Validate()
	if err != nil {
		return
	}

	*c = conf
	return
}

func (c *Configuration) Validate() (err error) {
	err = &Error{}
	if c.MaxProcs < 1 {
		err.(*Error).Add("MaxProcs must be >0")
	}

	if len(c.AccessKeyId) < 1 {
		err.(*Error).Add("AccessKeyId is empty")
	}

	if len(c.SecretAccessKey) < 1 {
		err.(*Error).Add("SecretAccessKey is empty")
	}

	srcErr := c.Source.Validate()
	if srcErr != nil {
		err.(*Error).Merge(srcErr)
	}

	idxErr := c.Index.Validate()
	if idxErr != nil {
		err.(*Error).Merge(idxErr)
	}

	if srcErr == nil || idxErr == nil {
		for k, _ := range c.Index.nameSet {
			_, exists := c.Source.nameSet[k]
			if !exists {
				err.(*Error).Add(fmt.Sprintf("Index.IndexFields unknown field %q", k))
			}
		}
	}

	if len(err.(*Error).messages) < 1 {
		err = nil
	}

	return
}

type Source struct {
	S3 S3
	Table DynamoTable
	IdColumn int
	SourceFields []SourceField
	Delimiter string
	nameSet map[string]bool
	shortNameSet map[string]bool
}

func (s *Source) Validate() (err error) {
	err = &Error{}

	if s.IdColumn < 1 {
		err.(*Error).Add("Source.IdColumn must be >1")
	}

	if len(s.Delimiter) < 1 {
		err.(*Error).Add("Source.Delimiter must be >1")
	}

	s.nameSet = make(map[string]bool)
	s.shortNameSet = make(map[string]bool)

	for _, field := range s.SourceFields {
		_, exists := s.nameSet[field.Name]
		if exists {
			err.(*Error).Add(fmt.Sprintf("SourceFields.Name duplicate %q", field.Name))
		} else {
			s.nameSet[field.Name] = true
		}

		_, exists = s.shortNameSet[field.ShortName]
		if exists {
			err.(*Error).Add(fmt.Sprintf("SourceFields.ShortName duplicate %q", field.ShortName))
		} else {
			s.shortNameSet[field.ShortName] = true
		}

		childErr := field.Validate()
		if childErr != nil {
			err.(*Error).Merge(childErr)
		}
	}

	childErr := s.Table.Validate()
	if childErr != nil {
		err.(*Error).Merge(childErr)
	}

	if len(err.(*Error).messages) < 1 {
		err = nil
	}
	return
}

type Index struct {
	S3 S3
	Table DynamoTable
	IndexFields []IndexField
	nameSet map[string]bool
}

func (i *Index) Validate() (err error) {
	err = &Error{}
	i.nameSet = make(map[string]bool)

	childErr := i.S3.Validate()
	if childErr != nil {
		err.(*Error).Merge(childErr)
	}

	childErr = i.Table.Validate()
	if childErr != nil {
		err.(*Error).Merge(childErr)
	}

	for _, field := range i.IndexFields {
		for _, n := range field.Names {
			i.nameSet[n] = true
		}
		childErr = field.Validate()
		if childErr != nil {
			err.(*Error).Merge(childErr)
		}
	}

	if len(err.(*Error).messages) < 1 {
		err = nil
	}
	return
}

type S3 struct {
	Bucket string
	Prefix string
}

func (s *S3) Validate() (err error) {
	if len(s.Bucket) < 3 || len(s.Bucket) > 255 {
		err = &Error{}
		err.(*Error).Add("S3.Bucket must be 3-255 characters long")
	}
	return
}

type DynamoTable struct {
	Name string
	ReadCapacityUnits int
	WriteCapacityUnits int
}

func (t *DynamoTable) Validate() (err error) {
	err = &Error{}

	if t.ReadCapacityUnits < 1 {
		err.(*Error).Add("Table.ReadCapacityUnits must be > 0")
	}

	if t.WriteCapacityUnits < 1 {
		err.(*Error).Add("Table.WriteCapacityUnits must be > 0")
	}

	if len(t.Name) < 3 || len(t.Name) > 255 {
		err.(*Error).Add("Table.Name must be 3-255 chars long")
	}

	if len(err.(*Error).messages) < 1 {
		err = nil
	}
	return
}

type SourceField struct {
	Name string
	Column int
	ShortName string
}

func (f *SourceField) Validate() (err error) {
	err = &Error{}

	if len(f.Name) < 1 {
		err.(*Error).Add("SourceField.Name is required")
	}

	if len(f.ShortName) < 1 {
		err.(*Error).Add("SourceField.ShortName is required")
	}

	if f.Column < 1 {
		err.(*Error).Add("SourceField.Column must be > 0")
	}

	if len(err.(*Error).messages) < 1 {
		err = nil
	}
	return
}

type IndexField struct {
	Names []string
}

func (f *IndexField) Validate() (err error) {
	if len(f.Names) < 1 {
		err = &Error{}
		err.(*Error).Add("IndexField.Names must not be empty")
	}

	return
}
