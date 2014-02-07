package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"willstclair.com/phosphorus/schema"
)

var cmdSchema = &Command{
	Run:       runSchema,
	UsageLine: "schema",
	Short:     "create a schema file",
}

var (
	schemaSchemaDef string // -schemadef flag
	schemaSourceDef string // -sourcedef flag
	schemaIn        string // -in flag
	schemaOut       string // -out flag
)

func init() {
	// cmdSchema.Run = runSchema

	cmdSchema.Flag.StringVar(&schemaSourceDef, "sourcedef", "", "")
	cmdSchema.Flag.StringVar(&schemaSchemaDef, "schemadef", "", "")
	cmdSchema.Flag.StringVar(&schemaIn, "in", "", "")
	cmdSchema.Flag.StringVar(&schemaOut, "out", "", "")
}

func runSchema(cmd *Command, args []string) {
	s := &schema.Schema{}
	sDef, err := ioutil.ReadFile(schemaSchemaDef)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = s.LoadJSON(sDef)
	if err != nil {
		panic(err)
	}

	src := &schema.FileSource{}
	srcDef, err := ioutil.ReadFile(schemaSourceDef)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(srcDef, &src)
	if err != nil {
		panic(err)
	}
	src.Glob = schemaIn
	src.Concurrent = 5

	c, err := src.GetChannel()
	if err != nil {
		panic(err)
	}
	s.LearnRecords(c)

	file, err := os.Create(schemaOut)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = s.Save(file)
	if err != nil {
		panic(err)
	}
}
