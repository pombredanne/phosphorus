
package environment

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"github.com/crowdmob/goamz/s3"
	"willstclair.com/phosphorus/config"
	"errors"
	"fmt"
	"time"
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
			ReadCapacityUnits: int64(10),
			WriteCapacityUnits: int64(10)},
		TableName: tableName}
}

type Environment struct {
	token  aws.Auth
	region aws.Region

	DynamoServer *dynamodb.Server
	IndexTableD *dynamodb.TableDescriptionT
	SourceTableD *dynamodb.TableDescriptionT
	IndexTable  *dynamodb.Table
	SourceTable *dynamodb.Table

	s3 *s3.S3
	IndexBucket *s3.Bucket
	IndexPrefix string

	SourceBucket *s3.Bucket
	SourcePrefix string
}

func BucketExists(b *s3.Bucket) (exists bool, err error) {
	_, err = b.List("", "/", "", 10)
	if err == nil {
		exists = true
	} else if err.(*s3.Error).Code == "NoSuchBucket" {
		err = nil
	}
	return
}

// brutal
func NewEnvironment(conf config.Configuration) (env *Environment, err error) {
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
	if err != nil { panic(err) }

	env = &Environment{
		region: region,
		token: token,
		DynamoServer: &dynamodb.Server{token,region},
	}

	//
	// DynamoDB (this is so gross right now)
	//

	// setup for index table
	var err_ error
	env.IndexTableD, err_ = env.DynamoServer.DescribeTable(conf.Index.Table.Name)
	if err_ != nil {
		switch err_.(type) {
		case *dynamodb.Error:
			if err_.(*dynamodb.Error).Code == "ResourceNotFoundException" {
				goto skip
			}
		}
		err = err_
		return
	} else {
		pk, _ := env.IndexTableD.BuildPrimaryKey()
		env.IndexTable = env.DynamoServer.NewTable(conf.Index.Table.Name, pk)
	}
skip:
	env.SourceTableD, err_ = env.DynamoServer.DescribeTable(conf.Source.Table.Name)

	if err_ != nil {
		switch err_.(type) {
		case *dynamodb.Error:
			if err_.(*dynamodb.Error).Code == "ResourceNotFoundException" {
				goto skip2
			}
		}
		err = err_
		return
	} else {
		pk, _ := env.SourceTableD.BuildPrimaryKey()
		env.SourceTable = env.DynamoServer.NewTable(conf.Source.Table.Name, pk)
	}
skip2:

	//
	// S3
	//
	env.s3 = s3.New(env.token, env.region)
	env.IndexBucket = env.s3.Bucket(conf.Index.S3.Bucket)
	env.IndexPrefix = conf.Index.S3.Prefix
	env.SourceBucket = env.s3.Bucket(conf.Source.S3.Bucket)
	env.SourcePrefix = conf.Index.S3.Prefix

	return
}
