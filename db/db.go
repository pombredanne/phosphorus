package db

import (
	"encoding/binary"
	"log"
	"time"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
)

var SignatureTableDescription = dynamodb.TableDescriptionT{
	AttributeDefinitions: []dynamodb.AttributeDefinitionT{
		dynamodb.AttributeDefinitionT{
			Name: "s",
			Type: dynamodb.TYPE_BINARY}},
	KeySchema: []dynamodb.KeySchemaT{
		dynamodb.KeySchemaT{
			AttributeName: "s",
			KeyType: "HASH"}},
	ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
		ReadCapacityUnits: 10,
		WriteCapacityUnits: 10},
	TableName: "signature"}

var RecordTableDescription = dynamodb.TableDescriptionT{
	AttributeDefinitions: []dynamodb.AttributeDefinitionT{
		dynamodb.AttributeDefinitionT{
			Name: "id",
			Type: dynamodb.TYPE_BINARY}},
	KeySchema: []dynamodb.KeySchemaT{
		dynamodb.KeySchemaT{
			AttributeName: "id",
			KeyType: "HASH"}},
	ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
		ReadCapacityUnits: 10,
		WriteCapacityUnits: 10},
	TableName: "record"}


func NewServer(accessKeyId string, secretAccessKey string, region aws.Region) *dynamodb.Server {
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(accessKeyId, secretAccessKey, "", expires)
	if err != nil { return false }

	return &dynamodb.Server{token, region}
}

func CreateTable() {
	pk, _ := tableDescription.BuildPrimaryKey()
	table, err := Server.CreateTable(tableDescription)
	if err != nil { panic(err) }
	Table = dynamodb.Table{&Server,table,pk}
}

func LoadItUp() {
	signature := make([]byte, 3)
	key := &dynamodb.Key{}
	for i := 0; i < (1 << 23); i++ {
		binary.PutUvarint(signature, uint64(i))
		key.HashKey = string(signature[:2])
		Table.UpdateAttributes(key,
			[]dynamodb.Attribute{*dynamodb.NewStringAttribute("a", "b")})
		if i % 100000 == 0 { log.Println(i) }
	}
}
