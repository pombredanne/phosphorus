package main

import (
	// "bufio"
	// "encoding/gob"
	// "encoding/base64"
	"encoding/json"
	// "encoding/binary"
	// "math/rand"
	// crand "crypto/rand"
	// "math"
	// "math/big"
	// "bytes"
	"flag"
	// "fmt"
	"log"
	"os"
	// "path/filepath"
	"runtime"
	// "sync"

	// "willstclair.com/metaphone3"
	"willstclair.com/phosphorus/index"
	"willstclair.com/phosphorus/encoder"
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
	c lib.Counter
	e lib.Encoder
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


func main() {
	flag.Parse()
	readConfiguration()

	runtime.GOMAXPROCS(Config.MaxProcs)
	log.Println(BANNER)
	switch action {
	// case "version":
	// 	flag.PrintDefaults()
	// case "prepare":
	// 	Counts()
	// case "genhash":
	// 	GenerateHashFamily()
	case "tables":
		CreateTables()
	case "testdynamo":
		TestTables()
	// case "index":
	// 	Index()
	default:
		log.Printf("Unrecognized action %q", action)
	}
	log.Println("Goodbye")
}
