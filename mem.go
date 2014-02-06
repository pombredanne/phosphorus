package main

import (
	"log"
	"os"
	"sync"
	"willstclair.com/phosphorus/schema"
)

var cmdMem = &Command{
	Run:       runMem,
	UsageLine: "mem",
	Short:     "index in memory",
}

func runMem(cmd *Command, args []string) {
	log.Println("hi")

	src := &schema.FileSource{
		Fields: schema.SourceFields{
			schema.SourceField{"last_name", 2},
			schema.SourceField{"first_name", 3},
			schema.SourceField{"city", 5}},
		IdColumn:   1,
		Delimiter:  ",",
		Glob:       "/Users/wsc/unindexed/flout/*.csv",
		Concurrent: 4}

	field := &schema.Field{
		Comment: "last_name",
		Attrs:   []string{"last_name"},
		Transforms: []*schema.TransformI{
			&schema.TransformI{Name: "upcase"}},
		Classifier: schema.NewTfIdfClassifier()}

	field2 := &schema.Field{
		Comment: "the city",
		Attrs:   []string{"city"},
		Transforms: []*schema.TransformI{
			&schema.TransformI{Name: "upcase"}},
		Classifier: schema.NewTfIdfClassifier()}

	field3 := &schema.Field{
		Comment: "last name",
		Attrs:   []string{"first_name"},
		Transforms: []*schema.TransformI{
			&schema.TransformI{Name: "upcase"}},
		Classifier: schema.NewTfIdfClassifier()}

	s := &schema.Schema{
		HashCount: 2048,
		Width:     12,
		Fields:    []*schema.Field{field, field2, field3}}
	s.Hyd()

	c, err := src.GetChannel()
	if err != nil {
		panic(err)
	}

	s.LearnRecords(c)

	file, err := os.Create("wscorp.schema")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = s.Save(file)

	if err != nil {
		panic(err)
	}

	// for record := range c {
	// 	log.Println(record)
	// 	err := ix.Write(record)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

}

var cmdMem2 = &Command{
	Run:       runMem2,
	UsageLine: "mem2",
	Short:     "index in memory",
}

func runMem2(cmd *Command, args []string) {
	s := &schema.Schema{}
	file, err := os.Open("wscorp.schema")
	if err != nil {
		panic(err)
	}
	err = s.Load(file)
	if err != nil {
		panic(err)
	}
	file.Close()

	src := &schema.FileSource{
		Fields: schema.SourceFields{
			schema.SourceField{"last_name", 2},
			schema.SourceField{"first_name", 3},
			schema.SourceField{"city", 5}},
		IdColumn:   1,
		Delimiter:  ",",
		Glob:       "/Users/wsc/unindexed/flout/florida_00.csv",
		Concurrent: 4}

	ix := schema.NewMemoryIndex(s)

	c, err := src.GetChannel()
	if err != nil {
		panic(err)
	}

	wait := &sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wait.Add(1)
		go func() {
			for record := range c {
				ix.Write(record)
			}
			wait.Done()
		}()
	}

	wait.Wait()
}
