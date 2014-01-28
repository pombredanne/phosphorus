package db

import (
	// "encoding/binary"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	// "log"
	"time"
)

var DummyRegion = aws.Region{
	"us-east-1",
	"https://ec2.us-east-1.amazonaws.com",
	"https://s3.amazonaws.com",
	"",
	false,
	false,
	"https://sdb.amazonaws.com",
	"https://sns.us-east-1.amazonaws.com",
	"https://sqs.us-east-1.amazonaws.com",
	"https://iam.amazonaws.com",
	"https://elasticloadbalancing.us-east-1.amazonaws.com",
	"http://localhost:8000",
	aws.ServiceInfo{"https://monitoring.us-east-1.amazonaws.com", aws.V2Signature},
	"https://autoscaling.us-east-1.amazonaws.com"}

var SignatureTableDescription = dynamodb.TableDescriptionT{
	AttributeDefinitions: []dynamodb.AttributeDefinitionT{
		dynamodb.AttributeDefinitionT{
			Name: "s",
			Type: dynamodb.TYPE_BINARY}},
	KeySchema: []dynamodb.KeySchemaT{
		dynamodb.KeySchemaT{
			AttributeName: "s",
			KeyType:       "HASH"}},
	ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
		ReadCapacityUnits:  10,
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
			KeyType:       "HASH"}},
	ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
		ReadCapacityUnits:  10,
		WriteCapacityUnits: 10},
	TableName: "record"}

func NewServer(accessKeyId string, secretAccessKey string) *dynamodb.Server {
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(accessKeyId, secretAccessKey, "", expires)
	if err != nil {
		panic(err)
	}

	return &dynamodb.Server{token, DummyRegion}
}

func CreateTable(server *dynamodb.Server, t dynamodb.TableDescriptionT) *dynamodb.Table {
	pk, _ := t.BuildPrimaryKey()
	table, err := server.CreateTable(t)
	if err != nil {
		panic(err)
	}
	return &dynamodb.Table{server, table, pk}
}
