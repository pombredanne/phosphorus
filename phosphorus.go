package main

import (
	"encoding/json"
	"fmt"
	log_ "log"
	"os"
	"runtime"
	"willstclair.com/phosphorus/data"
	"willstclair.com/phosphorus/encoder"
	"willstclair.com/phosphorus/index"
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
	MaxProcs        int
	AccessKeyId     string
	SecretAccessKey string
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
	d := data.NewData("./data", 4)
	var c encoder.Counter

	err := d.Slurp(func(records chan *data.Record) {
		for r := range records {
			c.Count(r.Fields)
		}
	})

	if err != nil {
		panic(err)
	}

	os.Chdir(working)

	e := encoder.NewEncoder(&c)
	e.Path = "./encoder"
	err = e.Save()
	if err != nil {
		panic(err)
	}

	log.Printf("Encoding complete; dimension: %d\n", e.Dimension)

	log.Println("Generating hash template.")
	t := index.Template{
		Directory: "./hash",
		Dimension: e.Dimension,
	}
	t.Generate()

	log.Println("uh welp")
}

func info() {
	var e encoder.Encoder
	e.Path = "./encoder"
	err := e.Load()
	if err == nil {
		log.Printf("Encoder found; dimension: %d\n", e.Dimension)
	} else {
		log.Println("Encoder not found.")
		return
	}

	t := index.Template{
		Directory: "./hash",
		Dimension: e.Dimension}
	t.Load()
	log.Printf("Uh, template found?")
}

func createTable() {
	server := index.NewRealServer(config.AccessKeyId, config.SecretAccessKey)
	_, err := server.CreateTable(index.SignatureTableDescription)
	if err != nil {
		panic(err)
	}
}

type RecordSig struct {
	Id        uint32
	Signature *index.Signature
}

func runindex() {
	var e encoder.Encoder
	e.Path = "./encoder"
	err := e.Load()
	if err != nil {
		panic(err)
	}

	t := index.Template{
		Directory: "./hash",
		Dimension: e.Dimension}
	t.Load()

	server := index.NewRealServer(config.AccessKeyId, config.SecretAccessKey)
	pk, _ := index.SignatureTableDescription.BuildPrimaryKey()
	table := server.NewTable("signature", pk)

	xr := index.NewIndex(5, table)

	d := data.NewData("./data", 4)

	records2 := make(chan *RecordSig)
	done := make(chan int)
	go func() {
		for r := range records2 {
			xr.Add(r.Id, r.Signature)
		}
		done <- 1
	}()

	err = d.Slurp(func(records chan *data.Record) {
		for r := range records {
			v := e.Encode(r.Fields)
			signature := t.Sign(v)
			records2 <- &RecordSig{r.RecordId, signature}
		}

	})
	if err != nil {
		log.Fatal(err)
	}

	close(records2)

	<-done
	xr.FlushAll()

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
	case "info":
		info()
	case "createtable":
		createTable()
	case "index":
		runindex()
	case "version":
		fmt.Println(VERSION)
	default:
		fmt.Printf("Unrecognized action %q", action)
		log.Fatalln(USAGE)
	}
}
