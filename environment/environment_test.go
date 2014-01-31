package environment

import (
	"fmt"
	"testing"
	"time"
	"bytes"
	"encoding/binary"
	"encoding/base64"
	"math"
	"log"
	"math/rand"
	"math/big"
	crand "crypto/rand"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
)

var LocalRegion = aws.Region{
	"localhost",
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
	"https://autoscaling.us-east-1.amazonaws.com",
	aws.ServiceInfo{"https://rds.us-east-1.amazonaws.com", aws.V2Signature}}

var dynamo *dynamodb.Server

func init() {
	seedRandom()
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(randomString(), "secret", "", expires)
	if err != nil {
		panic(err)
	}

	// localRegion := aws.Reg
	dynamo = &dynamodb.Server{token, LocalRegion}
}

func seedRandom() {
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}
	rand.Seed(seed.Int64())
}

func randomString() string {
	return fmt.Sprintf("%x", rand.Uint32())
}

func createTable() string {
	name := randomString()
	_, err := dynamo.CreateTable(*tableD(name, "k"))
	if err != nil {
		panic(err)
	}
	return name
}

func TestTableDoesNotExist(t *testing.T) {
	tbl := &table{dynamo, "does-not-exist", "k", nil}
	exists, err := tbl.Exists()
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Fail()
	}
}

func TestTableExists(t *testing.T) {
	name := createTable()
	tbl := &table{dynamo, name, "k", nil}
	exists, err := tbl.Exists()
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Fail()
	}
}

func TestCreateNewTable(t *testing.T) {
	name := randomString()
	tbl := &table{dynamo, name, "k", nil}
	err := tbl.Create()
	if err != nil {
		t.Error(err)
	}

	exists, err := tbl.Exists()
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Fail()
	}
}

func TestCreateExtantTable(t *testing.T) {
	name := createTable()
	tbl := &table{dynamo, name, "k", nil}
	err := tbl.Create()
	if err == nil {
		t.Fail()
	}
}

func TestDestroyTable(t *testing.T) {
	name := createTable()
	tbl := &table{dynamo, name, "k", nil}

	err := tbl.Destroy()
	if err != nil {
		t.Error(err)
	}

	exists, err := tbl.Exists()
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Fail()
	}
}

func binKey(iKey int64) (bKey string) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, iKey)
	bKey = base64.StdEncoding.EncodeToString(buf.Bytes())
	return
}

func testItems(key string) (start int64, items [][]dynamodb.Attribute) {
	start = time.Now().UnixNano()

	for i := 0; i < 25; i++ {
		j := int64(i) + start
		item := make([]dynamodb.Attribute, 2)
		item[0] = *dynamodb.NewBinaryAttribute(key, binKey(j))
		item[1] = *dynamodb.NewStringAttribute("test", "test")
		items = append(items, item)
	}
	return
}

func TestBatchPut(t *testing.T) {
	t.Skip("Skipping TestBatchPut")
	accessKey := ""
	secret := ""
	tableName := ""
	tableKey := ""

	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(
		accessKey, secret, "", expires)
	if err != nil {
		panic(err)
	}

	server := &dynamodb.Server{token, aws.USEast}

	tbl := &table{server, tableName, tableKey, nil}
	err = tbl.Load()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		_, items := testItems(tableKey)
		err = tbl.BatchPut(items)
		if err != nil {
			log.Println("error!")
			log.Println(err)
			t.FailNow()
		}
		// time.Sleep(2000 * time.Millisecond)
		// for j := int64(0); j < 25; j++ {
		// 	itemKey := &dynamodb.Key{binKey(start + j), ""}
		// 	_, err = tbl.table.GetItemConsistent(itemKey, true)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }
	}
}
