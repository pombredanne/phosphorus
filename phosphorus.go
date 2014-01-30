package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	log_ "log"
	"os"
	"runtime"
)

const (
	WORKING_DIR     = "/var/lib/phosphorus"
	WORKING_DIR_ENV = "PHOSPHORUS_DIR"
	CONFIG_FILE     = "config.json"
	LOG_FLAGS       = log_.Ldate | log_.Ltime | log_.Lshortfile
	VERSION         = "1.0.0-dev+deadbeef"
	USAGE           = "usage: phosphorus ( prepare | version )"
	BANNER          = `
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
)

var (
	working string
	log     log_.Logger
	config  Configuration
)

type Configuration struct {
	MaxProcs           int
	AWSSecretAccessKey string
	AWSAccessKeyId     string
}

func init() {
	var err error

	log = *log_.New(os.Stdout, "", LOG_FLAGS)

	working = os.Getenv(WORKING_DIR_ENV)
	if working == "" {
		working = WORKING_DIR
	}
	_, err = os.Stat(working)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("PHOSPHORUS_DIR does not exist: %s", working)
		}
		log.Fatalln(err)
	}

	if err = os.Chdir(working); err != nil {
		log.Fatalln(err)
	}

	file, err := os.Open(CONFIG_FILE)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if json.NewDecoder(file).Decode(&config) != nil {
		log.Fatalln(err)
	}

	runtime.GOMAXPROCS(config.MaxProcs)
}

func prepare() {
	var lines [][]string

	file, err := os.Open("/Users/wsc/unindexed/VoterExtract/DAD_20130903.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r := csv.NewReader(file)

	for line, err := r.Read(); err != io.EOF; line, err = r.Read() {
		if err != nil {
			log.Println(err)
		}
		lines = append(lines, line)
	}

	log.Println("goodbye?")
}

func main() {
	fmt.Println(BANNER)

	if len(os.Args) < 2 {
		log.Fatalln(USAGE)
	}
	action := os.Args[1]
	switch action {
	case "prepare":
		prepare()
	case "version":
		fmt.Println(VERSION)
	default:
		fmt.Printf("Unrecognized action %q", action)
		log.Fatalln(USAGE)
	}
}
