package environment

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/crowdmob/goamz/dynamodb"
	"log"
	"math"
	"sort"
	"sync"
	"time"
	"willstclair.com/phosphorus/schema"
)

const SET_ATTR = "ids"

type DynamoDBIndex struct {
	indexT    *dynamodb.Table
	sourceT   *dynamodb.Table
	indexM    *WriteSem
	sourceM   *WriteSem
	signer    schema.Signer
	buckets   [][]map[uint32]bool
	locks     [][]sync.Mutex
	threshold int
}

func NewDynamoDBIndex(s schema.Signer, indexT *dynamodb.Table, sourceT *dynamodb.Table) schema.Index {
	ix := &DynamoDBIndex{
		indexT:    indexT,
		sourceT:   sourceT,
		indexM:    NewWriteSem(5000),
		sourceM:   NewWriteSem(1000),
		signer:    s,
		threshold: 64}

	numChunks := s.SignatureLen()
	sigValues := 1 << uint(s.ChunkBits())

	ix.buckets = make([][]map[uint32]bool, numChunks)
	ix.locks = make([][]sync.Mutex, numChunks)
	for i := 0; i < numChunks; i++ {
		ix.buckets[i] = make([]map[uint32]bool, sigValues)
		ix.locks[i] = make([]sync.Mutex, sigValues)
	}

	return ix
}

const (
	DDB_TOOMUCH = "ProvisionedThroughputExceededException"
	DDB_ISE     = "InternalServerError"
)

func (ix *DynamoDBIndex) addAttrsIndex(key *dynamodb.Key, attrs []dynamodb.Attribute) error {
	for {
		ix.indexM.Write(1)
		_, err := ix.indexT.AddAttributes(key, attrs)
		if err == nil {
			break
		}
		switch err.(type) {
		case *dynamodb.Error:
			if err.(*dynamodb.Error).Code == DDB_TOOMUCH {
				ix.indexM.Backoff()
				continue
			} else if err.(*dynamodb.Error).Code == DDB_ISE {
				time.Sleep(1000 * time.Millisecond)
				continue
			}
		}
		return err
	}
	return nil
}

func (ix *DynamoDBIndex) flush(sIdx, sVal int) error {
	l := len(ix.buckets[sIdx][sVal])
	if l == 0 {
		return nil
	}
	ids := make([]string, 0, l)
	for k, _ := range ix.buckets[sIdx][sVal] {
		ids = append(ids, uint32ToBase64String(k))
	}
	key := &dynamodb.Key{binkey(sIdx, sVal), ""}
	attrs := []dynamodb.Attribute{*dynamodb.NewBinarySetAttribute(SET_ATTR, ids)}
	err := ix.addAttrsIndex(key, attrs)
	if err != nil {
		return err
	}
	ix.buckets[sIdx][sVal] = make(map[uint32]bool)

	return nil
}

func (ix *DynamoDBIndex) insertId(id uint32, sIdx, sVal int) error {
	ix.locks[sIdx][sVal].Lock()
	defer ix.locks[sIdx][sVal].Unlock()

	bucket := ix.buckets[sIdx][sVal]
	if bucket == nil {
		ix.buckets[sIdx][sVal] = make(map[uint32]bool)
	}

	ix.buckets[sIdx][sVal][id] = true

	if len(ix.buckets[sIdx][sVal]) >= ix.threshold {
		err := ix.flush(sIdx, sVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ix *DynamoDBIndex) insertRecord(id uint32, attrs map[string]string) error {
	dynamoAttrs := []dynamodb.Attribute{}
	for k, v := range attrs {
		dynamoAttrs = append(dynamoAttrs, *dynamodb.NewStringAttribute(k, v))
	}

	for {
		ix.sourceM.Write(1)
		_, err := ix.sourceT.PutItem(uint32ToBase64String(id), "", dynamoAttrs)
		if err == nil {
			break
		}
		switch err.(type) {
		case *dynamodb.Error:
			if err.(*dynamodb.Error).Code == DDB_TOOMUCH {
				ix.sourceM.Backoff()
				continue
			} else if err.(*dynamodb.Error).Code == DDB_ISE {
				time.Sleep(1000 * time.Millisecond)
				continue
			}
		}
		return err
	}
	return nil
}

func (ix *DynamoDBIndex) Write(record *schema.Record, r schema.RandomProvider) error {
	sigs, err := ix.signer.Sign(record.Attrs, r)
	if err != nil {
		return err
	}

	err = ix.insertRecord(record.Id, record.Attrs)
	if err != nil {
		return err
	}
	for sigIdx, sigVal := range sigs {
		err := ix.insertId(record.Id, sigIdx, int(sigVal))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ix *DynamoDBIndex) lockAndFlush(sIdx, sVal int) error {
	ix.locks[sIdx][sVal].Lock()
	defer ix.locks[sIdx][sVal].Unlock()
	return ix.flush(sIdx, sVal)
}

func (ix *DynamoDBIndex) Flush() error {
	for sIdx, sVals := range ix.buckets {
		for sVal, _ := range sVals {
			err := ix.lockAndFlush(sIdx, sVal)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func batchGet(table *dynamodb.Table, keys []dynamodb.Key, chunk int) ([]map[string]*dynamodb.Attribute, error) {
	items := []map[string]*dynamodb.Attribute{}

	for i := 0; i < len(keys); i += chunk {
		ub := i + chunk
		if ub > len(keys) {
			ub = len(keys)
		}
		bgi := table.BatchGetItems(keys[i:ub])
		results, err := bgi.Execute()
		if err != nil {
			return nil, err
		}

		for _, r := range results[table.Name] {
			items = append(items, r)
		}
	}
	return items, nil
}

const INDEX_BATCH_GET_CHUNK = 16

func (ix *DynamoDBIndex) batchGetKeys(sig []uint32) ([]uint32, error) {
	items, err := batchGet(ix.indexT, dynamokeys(sig), INDEX_BATCH_GET_CHUNK)
	if err != nil {
		return nil, err
	}

	recordIds := []uint32{}

	for _, item := range items {
		for _, setVal := range item[SET_ATTR].SetValues {
			recordIds = append(recordIds, base64StringToUint32(setVal))
		}
	}

	return recordIds, nil
}

func (ix *DynamoDBIndex) batchGetRecords(ids []uint32) ([]*schema.Record, error) {
	items, err := batchGet(ix.sourceT, recordkeys(ids), 16)
	if err != nil {
		return nil, err
	}

	sourceTHashKeyName := ix.sourceT.Key.KeyAttribute.Name

	records := make([]*schema.Record, 0, len(ids))

	for _, item := range items {
		record := &schema.Record{Attrs: make(map[string]string)}
		for _, attr := range item {
			if attr.Name == sourceTHashKeyName {
				record.Id = base64StringToUint32(attr.Value)
			} else {
				record.Attrs[attr.Name] = attr.Value
			}
		}
		records = append(records, record)
	}

	return records, nil
}

func (ix *DynamoDBIndex) Query(attrs map[string]string, r schema.RandomProvider) (results []schema.Result, err error) {
	sigs, err := ix.signer.Sign(attrs, r)
	if err != nil {
		return
	}

	ids, err := ix.batchGetKeys(sigs)
	if err != nil {
		return
	}

	counter := make(map[uint32]int)
	for _, id := range ids {
		counter[id]++
	}
	dedupedIds := make([]uint32, 0, len(counter))
	for id, _ := range counter {
		dedupedIds = append(dedupedIds, id)
	}
	records, err := ix.batchGetRecords(dedupedIds)
	if err != nil {
		return
	}
	results = make([]schema.Result, 0, len(counter))

	for _, record := range records {
		results = append(results, schema.Result{record, counter[record.Id]})
	}

	sort.Sort(sort.Reverse(schema.ByMatches(results)))

	return
}

func dynamokeys(sigs []uint32) []dynamodb.Key {
	keys := make([]dynamodb.Key, len(sigs))
	for sigIdx, sigVal := range sigs {
		keys[sigIdx] = dynamodb.Key{binkey(sigIdx, int(sigVal)), ""}
	}
	return keys
}

func binkey(sIdx, sVal int) string {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, uint8(sIdx))
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, binary.BigEndian, uint16(sVal))
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func recordkeys(ids []uint32) []dynamodb.Key {
	keys := make([]dynamodb.Key, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, dynamodb.Key{uint32ToBase64String(id), ""})
	}
	return keys
}

func uint32ToBase64String(i uint32) string {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, i)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func base64StringToUint32(e string) (i uint32) {
	b, err := base64.StdEncoding.DecodeString(e)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(b)
	binary.Read(buf, binary.BigEndian, &i)
	return
}

const REPLENISH_INTERVAL = 5000

type WriteSem struct {
	fill       int
	dead       bool
	bucket     int
	bucketLock sync.Mutex
	bucketCond *sync.Cond
	fails      int
	failsLock  sync.Mutex
}

func (ws *WriteSem) trickle() {
	for !ws.dead {
		ws.bucketLock.Lock()
		ws.failsLock.Lock()

		if ws.fails > 0 {
			ws.fill -= (ws.fails / 2)
		} else if ws.bucket == 0 {
			f := float64(ws.fill) * 1.1
			ws.fill = int(math.Ceil(f))
		}
		ws.fails = 0
		ws.bucket |= ws.fill
		ws.bucketCond.Broadcast()
		ws.failsLock.Unlock()
		ws.bucketLock.Unlock()
		log.Println("New fill: ", ws.fill)
		time.Sleep(REPLENISH_INTERVAL * time.Millisecond)
	}
}

func (ws *WriteSem) Write(n int) {
	ws.bucketLock.Lock()
	for ws.bucket < n {
		ws.bucketCond.Wait()
	}
	ws.bucket -= n
	ws.bucketLock.Unlock()
}

func (ws *WriteSem) Backoff() {
	ws.failsLock.Lock()
	ws.fails++
	ws.failsLock.Unlock()
}

func (ws *WriteSem) Kill() {
	ws.dead = true
}

func NewWriteSem(throughput int) *WriteSem {
	ws := &WriteSem{
		fill: throughput}
	ws.bucketCond = sync.NewCond(&ws.bucketLock)
	go ws.trickle()
	return ws
}
