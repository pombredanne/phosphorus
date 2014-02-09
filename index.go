package main

import (
	"encoding/json"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
	"willstclair.com/phosphorus/environment"
	"willstclair.com/phosphorus/random"
	"willstclair.com/phosphorus/schema"
)

var cmdIndex = &Command{
	Run:       runIndex,
	UsageLine: "index",
	Short:     "index",
}

var (
	indexDir         string
	indexSchema      string // -schema flag
	indexSourceDef   string // -sourcedef flag
	indexIn          string // -in flag
	indexSourceTable string
	indexIndexTable  string
)

func init() {
	cmdIndex.Flag.StringVar(&indexDir, "dir", "", "")
	cmdIndex.Flag.StringVar(&indexSchema, "schema", "", "")
	cmdIndex.Flag.StringVar(&indexSourceDef, "sourcedef", "", "")
	cmdIndex.Flag.StringVar(&indexIn, "in", "", "")
	cmdIndex.Flag.StringVar(&indexSourceTable, "sourcetable", "", "")
	cmdIndex.Flag.StringVar(&indexIndexTable, "indextable", "", "")
}

func runIndex(cmd *Command, args []string) {
	log.Println("hello")
	// get randomstore
	rs := random.NewRandomStore(indexDir)
	log.Println("randomstore")

	// load schema
	s := &schema.Schema{}
	file, err := os.Open(indexSchema)
	if err != nil {
		panic(err)
	}

	err = s.Load(file)
	if err != nil {
		panic(err)
	}
	file.Close()
	log.Println("schema")

	// sum := 0
	// for _, f := range s.Fields {
	// 	sum += f.Classifier.Dimension()
	// 	log.Printf("%s %d\n", f.Comment, f.Classifier.Dimension())
	// 	log.Println(f.Classifier.(*schema.TfIdfClassifier).Counts)
	// }

	// log.Println("Dimension: ", sum)
	// os.Exit(1)

	src := &schema.FileSource{}
	srcDef, err := ioutil.ReadFile(indexSourceDef)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(srcDef, &src)
	if err != nil {
		panic(err)
	}
	src.Glob = indexIn
	src.Concurrent = 5
	log.Println("filesource")

	c, err := src.GetChannel()
	if err != nil {
		panic(err)
	}
	log.Println("getchannel")

	// aws :(
	now := time.Now()
	expires := now.Add(time.Duration(60) * time.Minute)
	auth, err := aws.EnvAuth()

	if err != nil {
		auth, err = aws.GetAuth("", "", "", expires)
		if err != nil {
			panic(err)
		}
		panic(err)
	}
	log.Println("envauth")

	dynamo := &dynamodb.Server{auth, aws.USEast}
	sourceT := dynTable(dynamo, indexSourceTable)
	indexT := dynTable(dynamo, indexIndexTable)
	log.Println("tables")

	ix := environment.NewDynamoDBIndex(s, indexT, sourceT)
	log.Println("newindex")

	log.Println("go")

	var wait sync.WaitGroup

	for i := 0; i < 128; i++ {
		wait.Add(1)
		go func() {
			for record := range c {
				ix.Write(record, rs)
			}
			wait.Done()
		}()
	}
	log.Println("wait")
	wait.Wait()
	log.Println("goodbye")

}

func dynTable(s *dynamodb.Server, name string) *dynamodb.Table {
	td, err := s.DescribeTable(name)
	if err != nil {
		panic(err)
	}
	pk, err := td.BuildPrimaryKey()
	if err != nil {
		panic(err)
	}

	t := s.NewTable(name, pk)
	return t
}
