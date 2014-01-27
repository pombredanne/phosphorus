package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"fmt"
	"bufio"
	"encoding/gob"
	"willstclair.com/phosphorus/classifier"
	"willstclair.com/phosphorus/florida"
	"willstclair.com/phosphorus/vector"
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
	AccessKeyID     string
	SecretAccessKey string
	DynamoDBLocalDir string
}

type Configuration struct {
	MaxProcs int
	WorkingDirectory string
	SignatureCount int
	SignatureBits int
	AWS    _Aws
}

var configPath string
var action string
var inGlob string
var force bool
var Config Configuration
var records chan classifier.Record
var c classifier.Classifier
var HashFamily [][]vector.HashVector

func init() {
	defaultConfigPath := os.Getenv("PHOSPHORUS_CONFIG")
	if len(defaultConfigPath) == 0 {
		defaultConfigPath = os.Getenv("HOME") + "/.phosphorus"
	}

	flag.StringVar(&configPath, "config", defaultConfigPath, "path to configuration file")
	flag.StringVar(&action, "action", "version", "action to run")
	flag.StringVar(&inGlob, "in", "", "files to process")
	flag.BoolVar(&force, "f", false, "force overwrite of existing workspace data")

	records = make(chan classifier.Record)
}

func readConfiguration() {
	file, err := os.Open(configPath)
	if err != nil { log.Fatalf("Could not open %s: %s", configPath, err) }
	defer file.Close()

	err = json.NewDecoder(file).Decode(&Config)
	if err != nil {
		log.Fatalf("Error parsing configuration file %s: %s", configPath, err)
	}
}

func loadClassifier() error {
	file, err := os.Open(Config.WorkingDirectory + "/classifier")
	if err != nil { return err }
	defer file.Close()
	return gob.NewDecoder(file).Decode(&c)
}

func saveClassifier() error {
	file, err := os.Create(Config.WorkingDirectory + "/classifier")
	if err != nil { return err }
	defer file.Close()
	return gob.NewEncoder(file).Encode(&c)
}

func TrainClassifier() {
	if !force && loadClassifier() == nil {
		log.Fatalf("classifier already exists; use -f to overwrite")
	}

	files, err := filepath.Glob(inGlob)
	if err != nil {
		log.Fatalf("could not glob input files %s: %s", inGlob, err)
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

	go c.Listen(records)
	wait.Wait()

	log.Printf("File load complete. Dimensions: %d", c.Dimension())
	err = saveClassifier()
	if err != nil { log.Fatalf("saving classifier failed: %s", err) }
}

func readRecordFile(filename string, records chan classifier.Record) error {
	file, err := os.Open(filename)
	if err != nil { return err }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		record, err := florida.ParseRecord(line)
		if err != nil {
			log.Println(err)
			continue
		}
		records <- record
	}
	if err = scanner.Err(); err != nil { return err }
	return nil
}

func GenerateHashFamily() {
	var wait sync.WaitGroup

	err := loadClassifier()
	if err != nil { log.Fatalf("could not load classifier file: %s", err) }
	dim := c.Dimension()

	log.Printf("Generating hash family with dimension: %d", dim)

	for i := 0; i < Config.SignatureCount; i++ {
		path := Config.WorkingDirectory + fmt.Sprintf("/hash_%02x", i)
		wait.Add(1)
		go func() {
			err := generateHashFile(path, Config.SignatureBits, dim)
			wait.Done()
			if err != nil { panic(err) }
		}()
	}
	wait.Wait()

	log.Println("Hash family created.")
}

func generateHashFile (filename string, count int, dimension int) error {
	file, err := os.Create(filename)
	if err != nil { return err }
	defer file.Close()

	enc := gob.NewEncoder(file)
	for i := 0; i < count; i++ {
		err := enc.Encode(vector.Random(dimension).(vector.HashVector))
		if err != nil { return err }
	}
	return nil
}

func LoadHashFamily () {
	var wait sync.WaitGroup

	HashFamily = make([][]vector.HashVector, Config.SignatureCount)
	for i := 0; i < Config.SignatureCount; i++ {
		i := i
		path := fmt.Sprintf(Config.WorkingDirectory + "/hash_%02x", i)
		HashFamily[i] = make([]vector.HashVector, Config.SignatureBits)
		wait.Add(1)
		go func() {
			err := loadHashFile(path, &HashFamily[i])
			wait.Done()
			if err != nil { panic(err) }
		}()
	}

	wait.Wait()
	log.Println("Hash family loaded.")
}

func loadHashFile(filename string, hf *[]vector.HashVector) error {
	file, err := os.Open(filename)
	if err != nil { return err }

	defer func() { if err := file.Close(); err != nil { panic(err) }}()

	dec := gob.NewDecoder(file)

	for j := 0; j < Config.SignatureBits; j++ {
		if err = dec.Decode(&(*hf)[j]); err != nil { return err }
	}

	return nil
}

func Index() {
	err := loadClassifier()
	if err != nil { log.Fatalf("could not load classifier: %s", err) }
	LoadHashFamily()
}

func main() {
	flag.Parse()
	readConfiguration()

	runtime.GOMAXPROCS(Config.MaxProcs)
	log.Println(BANNER)
	switch action {
	case "version":
		flag.PrintDefaults()
	case "classify":
		TrainClassifier()
	case "genhash":
		GenerateHashFamily()
	case "loadhash":
		LoadHashFamily()
	case "index":
		Index()
	default:
		log.Println("Unrecognized action %q", action)
	}
	log.Println("Goodbye")
}
