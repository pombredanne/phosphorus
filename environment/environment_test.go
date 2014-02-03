package environment

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	s3_ "github.com/crowdmob/goamz/s3"
	"github.com/crowdmob/goamz/s3/s3test"
	"log"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

var region aws.Region
var dynamo *dynamodb.Server
var s3server *s3_.S3
var token *aws.Auth

func init() {
	seedRandom()
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(randomString(), "secret", "", expires)
	if err != nil {
		panic(err)
	}

	s3testserver, err := s3test.NewServer(&s3test.Config{})
	if err != nil {
		panic(err)
	}

	region = aws.Region{
		Name:                 "test",
		S3Endpoint:           s3testserver.URL(),
		S3LocationConstraint: true,
		DynamoDBEndpoint:     "http://localhost:8000",
	}

	dynamo = &dynamodb.Server{token, region}
	s3server = s3_.New(token, region)
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

func TestPutChannel(t *testing.T) {
	name := createTable()
	tbl := &table{dynamo, name, "k", nil}
	tbl.Load()

	c := tbl.PutChannel()
	for i := 0; i < 30; i++ {
		k := dynamodb.Key{binKey(int64(i)), ""}
		c <- Item{k, []dynamodb.Attribute{
			*dynamodb.NewStringAttribute("test", "test")}}
	}

	close(c)

	for i := 0; i < 30; i++ {
		k := dynamodb.Key{binKey(int64(i)), ""}
		_, err := tbl.table.GetItem(&k)
		if err != nil {
			panic(err)
		}
	}
}

func TestAddChannel(t *testing.T) {
	name := createTable()
	tbl := &table{dynamo, name, "k", nil}
	tbl.Load()

	k := dynamodb.Key{binKey(92825), ""}

	c := tbl.AddChannel()
	c <- Item{k, []dynamodb.Attribute{
		*dynamodb.NewStringSetAttribute("colors", []string{"red"})}}
	c <- Item{k, []dynamodb.Attribute{
		*dynamodb.NewStringSetAttribute("colors", []string{"blue"})}}

	close(c)

	time.Sleep(100 * time.Millisecond)

	item, err := tbl.table.GetItem(&k)
	if err != nil {
		panic(err)
	}

	if len(item["colors"].SetValues) != 2 {
		t.Fail()
	}
}

func TestBucketDoesNotExist(t *testing.T) {
	bkt := &bucket{s3server, "does-not-exist", "", nil}
	exists, err := bkt.Exists()
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Fail()
	}
}

func TestBucketExists(t *testing.T) {
	preExisting := s3server.Bucket("exists")
	err := preExisting.PutBucket(s3_.Private)
	if err != nil {
		panic(err)
	}

	bkt := &bucket{s3server, "exists", "", nil}
	exists, err := bkt.Exists()
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Fail()
	}
}

func TestCreateBucket(t *testing.T) {
	bkt := &bucket{s3server, "new-bucket", "", nil}
	err := bkt.Create()
	if err != nil {
		t.Error(err)
	}

	bkt = &bucket{s3server, "new-bucket", "", nil}
	exists, err := bkt.Exists()
	if err != nil {
		panic(err)
	}
	if !exists {
		t.Fail()
	}
}

func TestDestroyBucket(t *testing.T) {
	bkt := &bucket{s3server, "new-bucket", "", nil}
	err := bkt.Create()
	if err != nil {
		t.Error(err)
	}

	bkt = &bucket{s3server, "new-bucket", "", nil}
	err = bkt.Destroy()
	if err != nil {
		panic(err)
	}

	exists, err := bkt.Exists()
	if exists {
		t.Fail()
	}
}
