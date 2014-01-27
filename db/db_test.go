package db

import (
	// "testing"
	// "time"
	// "os"
	"os/exec"
)

var dynamoDBServer *exec.Cmd

const DDBLOCAL = "/Users/wsc/repo/dynamodb_local_2014-01-08"

var DummyRegion = aws.Region{
	"dummy-region",
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

// func startDynamoDB() {
// 	dynamoDBServer = exec.Command(
// 		"java",
// 		"-Djava.library.path=" + DDBLOCAL + "/DynamoDBLocal_lib",
// 		"-jar",
// 		DDBLOCAL + "/DynamoDBLocal.jar",
// 		"-inMemory")
// 	dynamoDBServer.Stdout = os.Stdout
// 	dynamoDBServer.Stderr = os.Stderr
// 	err := dynamoDBServer.Start()
// 	if err != nil { panic(err) }

// 	for {
// 		log.Println("attempting...")
// 		time.Sleep(250 * time.Millisecond)
// 		if !Connect() { continue }
// 		_, err := Server.ListTables()
// 		if err == nil { break }
// 	}
// }

// func stopDynamoDB() {
// 	dynamoDBServer.Process.Kill()
// }

// func TestCreateTable(t *testing.T) {
// 	startDynamoDB()
// 	defer stopDynamoDB()
// 	CreateTable()
// 	// table, err := CreateTable()
// 	// if err != nil { panic(err) }
// }

// func TestUpdateTable(t *testing.T) {
// 	startDynamoDB()
// 	defer stopDynamoDB()
// 	CreateTable()
// 	LoadItUp()
// }
