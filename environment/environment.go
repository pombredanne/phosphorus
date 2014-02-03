package environment

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"os"
	"sync"
	// "log"
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	s3_ "github.com/crowdmob/goamz/s3"
	"time"
	"willstclair.com/phosphorus/config"
	// "log"
)

type table struct {
	server *dynamodb.Server
	name   string
	key    string
	table  *dynamodb.Table
	// description *dynamodb.TableDescriptionT
}

func (t *table) Exists() (exists bool, err error) {
	if t.server == nil {
		panic(errors.New("dynamodb server not specified"))
	}

	if t.name == "" || t.key == "" {
		panic(errors.New("table key or name not specified"))
	}

	_, err_ := t.server.DescribeTable(t.name)
	if err_ != nil {
		switch err_.(type) {
		case *dynamodb.Error:
			switch err_.(*dynamodb.Error).Code {
			case "ResourceNotFoundException": // does not exist, no error
				exists = false
			case "ResourceInUseException": // does exist, no error
				exists = true
			default: // some other dynamo error
				err = err_
			}
		default: // non-dynamo error
			err = err_
		}
		return
	}
	exists = true
	return
}

func (t *table) Create() (err error) {
	exists, err := t.Exists()
	if err != nil {
		return
	}

	if exists {
		err = errors.New("table already exists")
	} else {
		_, err = t.server.CreateTable(*tableD(t.name, t.key))
	}
	return
}

func (t *table) Destroy() (err error) {
	exists, err := t.Exists()
	if err != nil {
		return
	}

	if !exists {
		err = errors.New("table does not exist")
		return
	}

	td, err := t.server.DescribeTable(t.name)
	if err != nil {
		return
	}
	_, err = t.server.DeleteTable(*td)

	return
}

func (t *table) Load() (err error) {
	exists, err := t.Exists()
	if err != nil {
		return
	}

	if !exists {
		err = errors.New("table does not exist")
		return
	}

	td, err := t.server.DescribeTable(t.name)
	if err != nil {
		return
	}

	pk, err := td.BuildPrimaryKey()
	if err != nil {
		return
	}

	t.table = t.server.NewTable(t.name, pk)
	return
}

// I loathe the DynamoDB API
func (t *table) BatchPut(items [][]dynamodb.Attribute) error {
	remainingKeys := make(map[string]int)

	if len(items) > 25 {
		return errors.New("too many items")
	}

	for i, item := range items {
		for _, attr := range item {
			if attr.Name == t.key {
				remainingKeys[attr.Value] = i
			}
		}
	}

	for attempt := 0; len(remainingKeys) > 0; attempt++ {
		putItems := make([][]dynamodb.Attribute, 0, 25)
		for _, i := range remainingKeys {
			putItems = append(putItems, items[i])
		}
		put := make(map[string][][]dynamodb.Attribute)
		put["Put"] = putItems
		bwi := t.table.BatchWriteItems(put)
		unprocessed, err := bwi.Execute()
		if unprocessed != nil {
			stillRemainingKeys := make(map[string]int)
			u := unprocessed[t.name].([]interface{})
			for _, i := range u {
				i = i.(map[string]interface{})["PutRequest"].(map[string]interface{})["Item"].(map[string]interface{})[t.key].(map[string]interface{})["B"]
				stillRemainingKeys[i.(string)] = remainingKeys[i.(string)]
			}
			remainingKeys = stillRemainingKeys
			continue
		} else if err != nil {
			switch err.(type) {
			case *dynamodb.Error:
				if err.(*dynamodb.Error).Code == "ProvisionedThroughputExceededException" {
					time.Sleep((1 << uint(attempt)) * time.Second)
					continue
				} else {
					return err
				}
			default:
				return err
			}
		} else {
			break
		}
	}
	return nil
}

type Item struct {
	Key        dynamodb.Key
	Attributes []dynamodb.Attribute
}

func (i *Item) ToAttributes(keyName string) (attrs []dynamodb.Attribute) {
	attrs = make([]dynamodb.Attribute, len(i.Attributes)+1)
	attrs[0] = *dynamodb.NewBinaryAttribute(keyName, i.Key.HashKey)
	for i, attr := range i.Attributes {
		attrs[i+1] = attr
	}
	return
}

func Dec64(e string) (i uint32) {
	b, err := base64.StdEncoding.DecodeString(e)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(b)
	binary.Read(buf, binary.BigEndian, &i)
	return
}

func Enc64(i uint32) (e string) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, i)
	e = base64.StdEncoding.EncodeToString(buf.Bytes())
	return
}

func NewSetItem(key uint32, members []uint32) (item *Item) {
	var buf bytes.Buffer
	var encodedSet []string

	for _, m := range members {
		encodedSet = append(encodedSet, Enc64(m))
	}

	buf.Reset()

	item = &Item{
		dynamodb.Key{Enc64(key), ""},
		[]dynamodb.Attribute{
			*dynamodb.NewBinarySetAttribute("i", encodedSet)}}

	return
}

func (t *table) PutChannel() (c chan Item) {
	c = make(chan Item)

	go func() {
		items := make([][]dynamodb.Attribute, 0, 25)
		for item := range c {
			items = append(items, item.ToAttributes(t.key))
			if len(items) == 25 {
				err := t.BatchPut(items)
				if err != nil {
					panic(err)
				}
				items = make([][]dynamodb.Attribute, 0, 25)
			}
		}
		err := t.BatchPut(items)
		if err != nil {
			panic(err)
		}
	}()

	return
}

func (t *table) AddChannel() (c chan Item) {
	c = make(chan Item)

	go func() {
		for item := range c {
			_, err := t.table.AddAttributes(&item.Key, item.Attributes)
			if err != nil {
				panic(err)
			}
		}
	}()

	return
}

func dynamoKeys(keys []uint32) (dkeys []dynamodb.Key) {
	for _, k := range keys {
		dkeys = append(dkeys, dynamodb.Key{HashKey: Enc64(k)})
	}
	return
}

func (t *table) BatchGet(keys []uint32, c chan uint32) (err error) {
	bgi := t.table.BatchGetItems(dynamoKeys(keys))
	results, err := bgi.Execute()
	// No throughput throttle handling here yet. Need to double-check goamz
	// to see what happens to the UnprocessedKeys field in the response.
	if err != nil {
		return
	}

	for _, r := range results[t.name] {
		for _, k := range r["i"].SetValues {
			c <- Dec64(k)
		}
	}
	return
}

func (t *table) MultiGet(keys []uint32, c chan uint32) {
	var wait sync.WaitGroup
	for i := 0; i < 8; i++ {
		i := i
		wait.Add(1)
		go func() {
			t.BatchGet(keys[i*16:(i+1)*16], c)
			wait.Done()
		}()
	}
	wait.Wait()
	close(c)
	return
}

// generate a dynamodb table description
func tableD(tableName, keyName string) *dynamodb.TableDescriptionT {
	return &dynamodb.TableDescriptionT{
		AttributeDefinitions: []dynamodb.AttributeDefinitionT{
			dynamodb.AttributeDefinitionT{
				Name: keyName,
				Type: dynamodb.TYPE_BINARY}},
		KeySchema: []dynamodb.KeySchemaT{
			dynamodb.KeySchemaT{
				AttributeName: keyName,
				KeyType:       "HASH"}},
		ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
			ReadCapacityUnits:  int64(10),
			WriteCapacityUnits: int64(10)},
		TableName: tableName}
}

type bucket struct {
	server *s3_.S3
	name   string
	prefix string
	bucket *s3_.Bucket
}

func (b *bucket) Exists() (exists bool, err error) {
	if b.bucket == nil {
		b.bucket = b.server.Bucket(b.name)
	}

	_, err = b.bucket.List(b.prefix, "/", "", 10)
	if err != nil && err.(*s3_.Error).Code == "NoSuchBucket" {
		err = nil
	} else if err == nil {
		exists = true
	}

	return
}

func (b *bucket) Create() (err error) {
	if b.bucket == nil {
		b.bucket = b.server.Bucket(b.name)
	}
	err = b.bucket.PutBucket(s3_.Private)
	return
}

func (b *bucket) Destroy() (err error) {
	if b.bucket == nil {
		b.bucket = b.server.Bucket(b.name)
	}
	err = b.bucket.DelBucket()
	return
}

type Environment struct {
	token  aws.Auth
	region aws.Region

	dynamo *dynamodb.Server
	s3     *s3_.S3

	IndexTable  *table
	SourceTable *table

	IndexBucket  *bucket
	SourceBucket *bucket

	TempDir string
}

func (e *Environment) Cleanup() error {
	return os.RemoveAll(e.TempDir)
}

func New(conf config.Configuration) (env *Environment, err error) {
	// create a temporary directory
	tempdir, err := ioutil.TempDir("", "phosphorus")
	if err != nil {
		return
	}

	// check our region
	region, exists := aws.Regions[conf.AWSRegion]
	if !exists {
		err = errors.New(fmt.Sprintf("unknown AWS region: %q", conf.AWSRegion))
		return
	}

	//
	// Credentials (not sure all of this ceremony is necessary)
	//
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)
	token, err := aws.GetAuth(conf.AccessKeyId, conf.SecretAccessKey, "", expires)
	if err != nil {
		return
	}

	dynamo := &dynamodb.Server{token, region}
	s3 := s3_.New(token, region)
	env = &Environment{
		region:      region,
		token:       token,
		dynamo:      dynamo,
		s3:          s3,
		IndexTable:  &table{dynamo, conf.Index.Table.Name, "s", nil},
		SourceTable: &table{dynamo, conf.Source.Table.Name, "r", nil},
		IndexBucket: &bucket{s3,
			conf.Index.S3.Bucket, conf.Index.S3.Prefix, nil},
		SourceBucket: &bucket{s3,
			conf.Source.S3.Bucket, conf.Source.S3.Prefix, nil},
		TempDir: tempdir,
	}

	return
}
