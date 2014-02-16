// Copyright 2014 William H. St. Clair

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"github.com/wsc/phosphorus/schema"
	"io/ioutil"
	"log"
	"os"
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
