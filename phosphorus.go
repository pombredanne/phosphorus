package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"encoding/binary"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"willstclair.com/phosphorus/lib"
	"willstclair.com/phosphorus/metaphone3"
	"willstclair.com/phosphorus/db"
	"github.com/crowdmob/goamz/dynamodb"
)

const BANNER = `
                           )
                          ) \
        . . . . . . . . ./ ) (. . .
       . . . . . . . . . \(_)/ . .
                         / /
                        / /
    Phosphorus         / /
    match server      / /
                     /_/
`

type _Aws struct {
	AccessKeyID      string
	SecretAccessKey  string
	DynamoDBLocalDir string
}

type Configuration struct {
	MaxProcs         int
	WorkingDirectory string
	SignatureCount   int
	SignatureBits    int
	AWS              _Aws
}

var (
	configPath string
	action string
	inGlob string
	force bool
	Config Configuration
	records chan *IDRecord
	c lib.Counter
	e lib.Encoder
	HashFamily [][]lib.HashVector
)

const (
	encFile = "encoder"
)

func init() {
	defaultConfigPath := os.Getenv("PHOSPHORUS_CONFIG")
	if len(defaultConfigPath) == 0 {
		defaultConfigPath = os.Getenv("HOME") + "/.phosphorus"
	}

	flag.StringVar(&configPath, "config", defaultConfigPath, "path to configuration file")
	flag.StringVar(&action, "action", "version", "action to run")
	flag.StringVar(&inGlob, "in", "", "files to process")
	flag.BoolVar(&force, "f", false, "force overwrite of existing workspace data")

	records = make(chan *IDRecord)
}

func readConfiguration() {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Could not open %s: %s", configPath, err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&Config)
	if err != nil {
		log.Fatalf("Error parsing configuration file %s: %s", configPath, err)
	}
}

func loadEncoder() error {
	file, err := os.Open(Config.WorkingDirectory + "/" + encFile)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewDecoder(file).Decode(&e)
}

func saveEncoder() error {
	file, err := os.Create(Config.WorkingDirectory + "/" + encFile)
	if err != nil {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(&e)
}

func Counts() {
	if !force && loadEncoder() == nil {
		log.Fatalf("encoder already exists; use -f to overwrite")
	}

	files, err := filepath.Glob(inGlob)
	if err != nil {
		log.Fatalf("could not glob input files %s: %s", inGlob, err)
	}

	go func() {
		for record := range records {
			c.Count(record.Fields)
		}
	}()

	var wait sync.WaitGroup
	for _, filename := range files {
		wait.Add(1)
		filename := filename
		// err := readRecordFile(filename, records)
		// if err != nil {
		// 	panic(err)
		// }
		go func() {
			err := readRecordFile(filename, records)
			if err != nil {
				panic(err)
			}
			wait.Done()
		}()
	}


	wait.Wait()

	log.Printf("File load complete.")

	e = *lib.NewEncoder(&c)
	err = saveEncoder()
	if err != nil {
		log.Fatalf("saving encoder failed: %s", err)
	}
}

type IDRecord struct {
	Id uint
	Fields []interface{}
}

func readRecordFile(filename string, records chan *IDRecord) error {
	mp := metaphone3.NewMetaphone3()
	log.Printf("Starting %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		record, err := lib.ParseRecord(line)
		if err != nil {
			log.Println(err)
			continue
		}
		f := record.Fields(mp)
		records <- &IDRecord{record.RecordId, f}
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	log.Printf("Finished %s", filename)
	return nil
}

func GenerateHashFamily() {
	var wait sync.WaitGroup

	err := loadEncoder()
	if err != nil {
		log.Fatalf("could not load encoder: %s", err)
	}

	dim := e.Dimension

	log.Printf("Generating hash family with dimension: %d", dim)

	for i := 0; i < Config.SignatureCount; i++ {
		path := Config.WorkingDirectory + fmt.Sprintf("/hash_%02x", i)
		wait.Add(1)
		go func() {
			err := generateHashFile(path, Config.SignatureBits, dim)
			wait.Done()
			if err != nil {
				panic(err)
			}
		}()
	}
	wait.Wait()

	log.Println("Hash family created.")
}

func generateHashFile(filename string, count int, dimension int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	for i := 0; i < count; i++ {
		err := enc.Encode(lib.Random(dimension).(lib.HashVector))
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadHashFamily() {
	var wait sync.WaitGroup

	HashFamily = make([][]lib.HashVector, Config.SignatureCount)
	for i := 0; i < Config.SignatureCount; i++ {
		i := i
		path := fmt.Sprintf(Config.WorkingDirectory+"/hash_%02x", i)
		HashFamily[i] = make([]lib.HashVector, Config.SignatureBits)
		wait.Add(1)
		go func() {
			err := loadHashFile(path, &HashFamily[i])
			wait.Done()
			if err != nil {
				panic(err)
			}
		}()
	}

	wait.Wait()
	log.Println("Hash family loaded.")
}

func loadHashFile(filename string, hf *[]lib.HashVector) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	dec := gob.NewDecoder(file)

	for j := 0; j < Config.SignatureBits; j++ {
		if err = dec.Decode(&(*hf)[j]); err != nil {
			return err
		}
	}

	return nil
}

func ddbFlush() {
	buckets := new([1<<23][]uint)
	server := db.NewServer(Config.AWS.AccessKeyID, Config.AWS.SecretAccessKey)
	pk, _ := db.SignatureTableDescription.BuildPrimaryKey()
	table := server.NewTable("signature", pk)
	for record := range records {
		v := e.Encode(record.Fields)
		s, _ := lib.SignatureSet(v, HashFamily)
		for i, t := range s {
			x := uint64(t)
			x |= (uint64(i) << 16)

			buckets[x] = append(buckets[x], record.Id)
			// log.Printf("%x %d", x, len(buckets[x]))
			sig := make([]byte, 5)
			recordId := make([]byte, 5)
			if len(buckets[x]) > 0 {
				log.Printf("FLUSH: %x", x)
				binary.PutUvarint(sig, uint64(x))
				recordIds := make([]string, 10)
				for _, rid := range buckets[x] {

					binary.PutUvarint(recordId, uint64(rid))
					recordIds = append(recordIds, string(recordId[:5]))
				}
				attr := dynamodb.NewBinarySetAttribute("s", recordIds)
				// log.Printf("%x %s", sig, recordIds)
				_, err := table.PutItem(string(sig[:5]), "",
					[]dynamodb.Attribute{*attr})
				if err != nil { panic(err) }
				buckets[x] = make([]uint, 0, 10)
			}
		}
	}
}

func Index() {
	err := loadEncoder()
	if err != nil {
		log.Fatalf("could not load encoder: %s", err)
	}
	log.Println(e.Dimension)
	LoadHashFamily()

	files, _ := filepath.Glob(inGlob)

	for i := 0; i < 4; i++ {
		go ddbFlush()
	}

	var wait sync.WaitGroup
	for _, filename := range files {
		wait.Add(1)
		filename := filename
		go func() {
			err := readRecordFile(filename, records)
			if err != nil { panic(err) }
			wait.Done()
		}()
	}

	wait.Wait()
	log.Printf("Lol.")
}

func CreateTables() {
	// server := db.NewServer(Config.AWS.AccessKeyID, Config.AWS.SecretAccessKey)
	// sig := db.CreateTable(server, db.SignatureTableDescription)
	// log.Println(sig)

	// pk, _ := db.SignatureTableDescription.BuildPrimaryKey()
	// table := server.NewTable("signature", pk)

	// attr := dynamodb.NewBinarySetAttribute("s", recordIds)

	buf := new(bytes.Buffer)
	beef := uint64(0x7fbeef)
	err := binary.Write(buf, binary.BigEndian, beef)
	if err != nil { log.Fatal(err) }
	log.Printf("%x", buf.Bytes())


	// beefBuf := make([]byte, 5)

	// binary.PutUvarint(beefBuf, beef)



	// _, err := table.PutItem(string(sig[:5]), "",
	// 	[]dynamodb.Attribute{*attr})

	// db.CreateTable(server, db.RecordTableDescription)
}

func main() {
	flag.Parse()
	readConfiguration()

	runtime.GOMAXPROCS(Config.MaxProcs)
	log.Println(BANNER)
	switch action {
	case "version":
		flag.PrintDefaults()
	case "prepare":
		Counts()
	case "genhash":
		GenerateHashFamily()
	case "tables":
		CreateTables()
	case "index":
		Index()
	default:
		log.Printf("Unrecognized action %q", action)
	}
	log.Println("Goodbye")
}
