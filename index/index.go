package index

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"log"
	"os"
	"sort"
	"sync"
	"time"
	"willstclair.com/phosphorus/vector"
)

const (
	THROUGHPUT_EXCEEDED  = "ProvisionedThroughputExceededException"
	TABLE_NAME_SIGNATURE = "signature"
)

var (
	SignatureTableDescription = dynamodb.TableDescriptionT{
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

	LocalRegion = aws.Region{
		"localhost",
		"https://ec2.us-east-1.amazonaws.com",
		"https://s3.amazonaws.com",
		"",
		false,
		false,
		"https://sdb.amazonaws.com",
		"https://sns.us-east-1.amazonaws.com",
		"https://sqs.us-east-1.amazonaws.com",
		// "https://iam.amazonaws.com",
		"http://localhost:8000",
		"https://elasticloadbalancing.us-east-1.amazonaws.com",
		"http://localhost:8000",
		aws.ServiceInfo{"https://monitoring.us-east-1.amazonaws.com", aws.V2Signature},
		"https://autoscaling.us-east-1.amazonaws.com"}
)

func NewServer(accessKeyId string, secretAccessKey string) *dynamodb.Server {
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(accessKeyId, secretAccessKey, "", expires)
	if err != nil {
		panic(err)
	}

	// return &dynamodb.Server{token, aws.USEast}
	return &dynamodb.Server{token, LocalRegion}
}

func NewRealServer(accessKeyId string, secretAccessKey string) *dynamodb.Server {
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)

	token, err := aws.GetAuth(accessKeyId, secretAccessKey, "", expires)
	if err != nil {
		panic(err)
	}

	return &dynamodb.Server{token, aws.USEast}
	// return &dynamodb.Server{token, LocalRegion}
}

type Signature [128]uint16

func (s *Signature) Key(i int) uint32 {
	return uint32(s[i]) | uint32(i<<16)
}

func Key64(k uint32, trim int, buf bytes.Buffer) string {
	binary.Write(&buf, binary.BigEndian, k)
	k64 := base64.StdEncoding.EncodeToString(buf.Bytes()[trim:])
	buf.Reset()
	return k64
}

func (s *Signature) Keys64() []string {
	var buf bytes.Buffer
	keys := make([]string, len(s))
	for i, _ := range s {
		keys[i] = Key64(s.Key(i), 1, buf)
	}
	return keys
}

func (s *Signature) DynamoKeys() []dynamodb.Key {
	dkeys := make([]dynamodb.Key, len(s))
	for i, k := range s.Keys64() {
		dkeys[i] = dynamodb.Key{HashKey: k}
	}
	return dkeys
}

type Template struct {
	Directory string
	Dimension int
	family    [128][16]vector.Interface
}

func (t *Template) Sign(v vector.Interface) *Signature {
	var s Signature

	for i, h := range t.family {
		s[i] = vector.Signature(v, h[:])
	}

	return &s
}

func (t *Template) Load() {
	var wait sync.WaitGroup
	for i, _ := range t.family {
		i := i
		filename := fmt.Sprintf("hash_%02x", i)
		wait.Add(1)
		go func() {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("Error loading template: %s", err)
			}
			defer file.Close()

			dec := gob.NewDecoder(file)
			for j, _ := range t.family[i] {
				var v vector.HashVector
				err = dec.Decode(&v)
				if err != nil {
					log.Fatalf("Error loading template: %s", err)
				}
				t.family[i][j] = vector.Interface(v)
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

func (t *Template) Generate() {
	var wait sync.WaitGroup
	for i, _ := range t.family {
		filename := fmt.Sprintf("hash_%02x", i)
		wait.Add(1)
		go func() {
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal("Error generating template: %s", err)
			}
			defer file.Close()
			enc := gob.NewEncoder(file)
			for _ = range t.family[i] {
				err = enc.Encode(vector.Random(t.Dimension).(vector.HashVector))
				if err != nil {
					log.Fatal("Error generating template: %s", err)
				}
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

type Entry struct {
	Key       uint32
	RecordIds []uint32
}

func (e *Entry) RecordsBase64() []string {
	var buf bytes.Buffer
	r64 := make([]string, len(e.RecordIds))
	for i, r := range e.RecordIds {
		r64[i] = Key64(r, 0, buf)
	}
	return r64
}

func (s *Entry) Save(t *dynamodb.Table) error {
	var buf bytes.Buffer
	k := &dynamodb.Key{
		HashKey: Key64(s.Key, 1, buf),
	}
	a := dynamodb.NewBinarySetAttribute("i", s.RecordsBase64())
	_, err := t.AddAttributes(k, []dynamodb.Attribute{*a})
	return err
}

type Index struct {
	entries   [1 << 23][]uint32
	threshold int
	table     *dynamodb.Table
	flushSem  chan int
}

const FLUSHMAX = 3

func NewIndex(threshold int, table *dynamodb.Table) *Index {
	xr := Index{
		threshold: threshold,
		table:     table,
		flushSem:  make(chan int, FLUSHMAX),
	}
	for i := 0; i < FLUSHMAX; i++ {
		xr.flushSem <- 1
	}

	for i := 0; i < 1<<23; i++ {
		xr.entries[i] = make([]uint32, 0, threshold)
	}
	return &xr
}

func (xr *Index) Add(recordId uint32, s *Signature) {
	for si, v := range s {
		i := int(v) | (si << 16)
		xr.entries[i] = append(xr.entries[i], recordId)
		if len(xr.entries[i]) >= xr.threshold {
			<-xr.flushSem
			go func() { xr.Flush(i); xr.flushSem <- 1 }()
		}
	}
}

func (xr *Index) Flush(i int) {
	if len(xr.entries[i]) > 0 {
		ids := xr.entries[i]
		xr.entries[i] = make([]uint32, 0, xr.threshold)

		e := Entry{uint32(i), ids}

		err_count := 0
		for err := e.Save(xr.table); err != nil; { // should limit retries and back off
			// need a type switch here
			err := err.(*dynamodb.Error)
			if err.Code == THROUGHPUT_EXCEEDED {
				log.Println("dynamo: retrying write...")
				err_count++
				if err_count > 5 { panic(err) }
				time.Sleep(500 * time.Millisecond)
			} else {
				// log.Fatalf("flush error: %s", err)
				panic(err)
			}
		}
		// log.Printf("Flush %06x\n", i)
		// log.Printf("Flush %06x OK: %s\n", i, e)
	}
}

func (xr *Index) FlushAll() {
	log.Println("Flushing all writes")
	for i, e := range xr.entries {
		if len(e) > 0 {
			xr.Flush(i)
		}
	}
}

type Candidate struct {
	RecordId uint32
	Matches  int
}

func (x *Index) Query(s *Signature) []Candidate {
	var wait sync.WaitGroup
	c := make(chan uint32)
	out := make(chan []Candidate)
	go x.rankCandidates(c, out)
	keys := s.DynamoKeys()
	for i := 0; i < 8; i++ {
		i := i
		wait.Add(1)
		go func() {
			x.batchGet(keys[i*16:(i+1)*16], c)
			wait.Done()
		}()
	}
	wait.Wait()
	close(c)
	return <-out
}

func (x *Index) batchGet(keys []dynamodb.Key, c chan uint32) {
	bgi := x.table.BatchGetItems(keys)
	results, err := bgi.Execute()
	// No throughput throttle handling here yet. Need to double-check goamz
	// to see what happens to the UnprocessedKeys field in the response.
	if err != nil {
		log.Fatalf("dynamodb error %s", err)
	}
	// if len(results[TABLE_NAME_SIGNATURE]) < len(keys) {
	// 	log.Printf("Warning: missing keys: len(%d) < len(%d)",
	// 		len(results[TABLE_NAME_SIGNATURE]), len(keys))
	// }
	var recordId uint32
	for _, r := range results[TABLE_NAME_SIGNATURE] {
		for _, k := range r["i"].SetValues {
			b, _ := base64.StdEncoding.DecodeString(k)
			buf := bytes.NewBuffer(b)
			binary.Read(buf, binary.BigEndian, &recordId)

			c <- recordId
		}
	}
}

type ByMatches []Candidate

func (c ByMatches) Len() int           { return len(c) }
func (c ByMatches) Less(i, j int) bool { return c[i].Matches < c[j].Matches }
func (c ByMatches) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func (x *Index) rankCandidates(c chan uint32, out chan []Candidate) {
	counter := make(map[uint32]int)
	var candidates []Candidate

	for r := range c {
		counter[r]++
	}
	for k, v := range counter {
		candidates = append(candidates, Candidate{k, v})
	}
	sort.Sort(sort.Reverse(ByMatches(candidates)))
	out <- candidates
	close(out)
}
