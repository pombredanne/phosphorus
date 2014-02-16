// Copyright 2014 William H. St. Clair

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	crand "crypto/rand"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

type _test struct {
	ParentId int64    `dynamodb:"_hash"`
	Id       int64    `dynamodb:"_range"`
	Animal   string   `dynamodb:"animal"`
	I        int      `dynamodb:"i"`
	I8       int8     `dynamodb:"i8"`
	I16      int16    `dynamodb:"i16"`
	I32      int32    `dynamodb:"i32"`
	I64      int64    `dynamodb:"i64"`
	U        uint     `dynamodb:"u"`
	U8       uint8    `dynamodb:"u8"`
	U16      uint16   `dynamodb:"u16"`
	U32      uint32   `dynamodb:"u32"`
	U64      uint64   `dynamodb:"u64"`
	IS       []int16  `dynamodb:"is16"`
	US       []uint16 `dynamodb:"us16"`
	SS       []string `dynamodb:"ss"`
}

var dynamo *dynamodb.Server
var region aws.Region
var token *aws.Auth

// seed the RNG with an actually random value
func seedRandom() {
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}
	rand.Seed(seed.Int64())
}

func init() {
	seedRandom()
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(randomString(), "secret", "", expires)
	if err != nil {
		panic(err)
	}

	region = aws.Region{
		Name:             "test",
		DynamoDBEndpoint: "http://localhost:8000",
	}

	dynamo = &dynamodb.Server{token, region}
}

// generate a dynamodb table description
func tableD(tableName string) *dynamodb.TableDescriptionT {
	return &dynamodb.TableDescriptionT{
		AttributeDefinitions: []dynamodb.AttributeDefinitionT{
			dynamodb.AttributeDefinitionT{
				Name: "myhash",
				Type: dynamodb.TYPE_NUMBER},
			dynamodb.AttributeDefinitionT{
				Name: "myrange",
				Type: dynamodb.TYPE_NUMBER}},
		KeySchema: []dynamodb.KeySchemaT{
			dynamodb.KeySchemaT{
				AttributeName: "myhash",
				KeyType:       "HASH"},
			dynamodb.KeySchemaT{
				AttributeName: "myrange",
				KeyType:       "RANGE"}},
		ProvisionedThroughput: dynamodb.ProvisionedThroughputT{
			ReadCapacityUnits:  int64(10),
			WriteCapacityUnits: int64(10)},
		TableName: tableName}
}

func randomString() string {
	return fmt.Sprintf("%x", rand.Uint32())
}

func createTable() string {
	name := randomString()
	_, err := dynamo.CreateTable(*tableD(name))
	if err != nil {
		panic(err)
	}

	return name
}

func getTable(server *dynamodb.Server, name string) *dynamodb.Table {
	td, err := server.DescribeTable(name)
	if err != nil {
		panic(err)
	}

	pk, err := td.BuildPrimaryKey()
	if err != nil {
		panic(err)
	}

	return server.NewTable(name, pk)
}

func getRandomTable() *dynamodb.Table {
	return getTable(dynamo, createTable())
}

func TestGetPut(t *testing.T) {
	s := &_test{1000, 2000, "dog", 2147483647, 127, 32767, 2147483647, 9223372036854775807, 4294967295, 255, 65535, 4294967295, 18446744073709551615, []int16{1, 2, 3}, []uint16{32768, 32769, 32770}, []string{"apple", "orange", "banana"}}

	tbl := getRandomTable()
	err := OverwriteItem(tbl, s)
	if err != nil {
		t.Error(err)
	}

	s2 := &_test{
		ParentId: 1000,
		Id:       2000}
	err = GetItem(tbl, s2)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(s, s2) {
		t.Fail()
	}
}

type _state struct {
	ParentId  int64 `dynamodb:"_hash"`
	Id        int64 `dynamodb:"_range"`
	State     int   `dynamodb:"state"`
	Timestamp int64 `dynamodb:"timestamp"`
}

func TestCreate(t *testing.T) {
	tbl := getRandomTable()
	s0 := &_state{1000, 2000, 1, 1}
	err := CreateItem(tbl, s0)
	if err != nil {
		t.Error(err)
	}
	s0.State = 4
	err = CreateItem(tbl, s0)
	if err == nil {
		t.Log("Permitted overwrite!")
		t.Fail()
	}
}

func TestConditionalUpdate(t *testing.T) {
	tbl := getRandomTable()
	s0 := &_state{1000, 2000, 1, 1392049274}
	err := CreateItem(tbl, s0)
	if err != nil {
		t.Error(err)
	}

	s1 := &_state{1000, 2000, 2, 1392049399}
	success, err := ConditionalUpdate(tbl, s1, s0)
	if err != nil {
		t.Error(err)
	}

	if !success {
		t.Fail()
	}

	s2 := &_state{1000, 2000, 3, 1392049461}
	success, err = ConditionalUpdate(tbl, s2, s0)
	if err == nil || success {
		t.Fail()
	}
}

type _set struct {
	ParentId  int64    `dynamodb:"_hash"`
	Id        int64    `dynamodb:"_range"`
	RecordIds []uint32 `dynamodb:"record_ids"`
}

func TestAddAttributes(t *testing.T) {
	tbl := getRandomTable()
	s0 := &_set{1, 2, []uint32{3, 4, 5}}

	err := AddAttributes(tbl, s0)
	if err != nil {
		t.Error(err)
	}

	s1 := &_set{1, 2, []uint32{6, 7}}
	err = AddAttributes(tbl, s1)
	if err != nil {
		t.Error(err)
	}

	s2 := &_set{1, 2, []uint32{}}
	err = GetItem(tbl, s2)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(s2.RecordIds, []uint32{3, 4, 5, 6, 7}) {
		t.Fail()
	}
}
