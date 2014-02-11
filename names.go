package main

import (
	"fmt"
	"os"
	// "willstclair.com/metaphone3"
	"willstclair.com/phosphorus/schema"
)

var cmdNames = &Command{
	Run:       runNames,
	UsageLine: "names",
	Short:     "names",
}

var (
	namesSchema string // -schema
)

func init() {
	cmdNames.Flag.StringVar(&namesSchema, "schema", "", "")
}

func runNames(cmd *Command, args []string) {
	// load schema
	s := &schema.Schema{}
	file, err := os.Open(namesSchema)
	if err != nil {
		panic(err)
	}

	err = s.Load(file)
	if err != nil {
		panic(err)
	}
	file.Close()

	// mp := metaphone3.NewMetaphone3()

	for k, v := range s.Fields[0].Classifier.(*schema.TfIdfClassifier).Counts {
		fmt.Printf("%s\t%d\n", k, v)
	}

	// metaphone3.DeleteMetaphone3(mp)

	os.Exit(1)

}
