package db

import (
	// "testing"
	// "time"
	// "os"
	"os/exec"
)

var dynamoDBServer *exec.Cmd

const DDBLOCAL = "/Users/wsc/repo/dynamodb_local_2014-01-08"


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
