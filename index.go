package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"willstclair.com/phosphorus/schema"
)

var cmdIndex = &Command{
	Run:       runIndex,
	UsageLine: "index",
	Short:     "index",
}

var (
	indexSchema      string // -schema flag
	indexSourceDef   string // -sourcedef flag
	indexIn          string // -in flag
	indexSourceTable string
	indexIndexTable  string
)

func init() {
	cmdIndex.Flag.StringVar(&indexSchema, "schema", "", "")
	cmdIndex.Flag.StringVar(&indexSourceDef, "sourcedef", "", "")
	cmdIndex.Flag.StringVar(&indexIn, "in", "", "")
	cmdIndex.Flag.StringVar(&indexSourceTable, "sourcetable", "", "")
	cmdIndex.Flag.StringVar(&indexIndexTable, "indextable", "", "")
}

func runIndex(cmd *Command, args []string) {
	log.Println("hello")
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

	sum := 0
	for _, f := range s.Fields {
		sum += f.Classifier.Dimension()
		log.Printf("%s %d\n", f.Comment, f.Classifier.Dimension())
		log.Println(f.Classifier.(*schema.TfIdfClassifier).Counts)
	}

	log.Println("Dimension: ", sum)
	os.Exit(1)

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

	c, err := src.GetChannel()
	if err != nil {
		panic(err)
	}

	// ix := schema.NewMemoryIndex(s)

	var wait sync.WaitGroup

	for i := 0; i < 5; i++ {
		wait.Add(1)
		go func() {
			for record := range c {
				s.Sign(record.Attrs)
			}
			wait.Done()
		}()
	}
	wait.Wait()

}
