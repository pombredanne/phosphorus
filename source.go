package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sync"
	"willstclair.com/phosphorus/environment"
	"willstclair.com/phosphorus/data"
)

var cmdSource = &Command{
	Run:       runSource,
	UsageLine: "source",
	Short:     "populate the source table",
}

func runSource(cmd *Command, args []string) {
	// set up environment
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	// set up field descriptions
	var schema data.SourceSchema
	for _, field := range conf.Source.SourceFields {
		schema = append(schema, data.Attribute{
			field.Name, field.Column, field.ShortName})
	}

	// begin slurping
	log.Println("Loading source data")

	// create a channel for filehandles from S3
	c := make(chan io.ReadCloser)
	go func() {
		env.SourceBucket.OpenAll(c)
		close(c)
	}()

	// get our channel for saving items
	err = env.SourceTable.Load()
	if err != nil {
		panic(err)
	}
	wc := env.SourceTable.PutChannel()

	// limit the total number of simultaneous downloads
	sem := make(chan int, 5)
	for i := 0; i < 5; i++ {
		sem <- 1
	}
	// parse each record and add to the putchannel
	var wait sync.WaitGroup
	for rc := range c {
		rc := rc
		<-sem
		wait.Add(1)
		go func() {
			log.Println("starting...")
			defer rc.Close()
			r := csv.NewReader(rc)
			for line, err := r.Read(); err != io.EOF; line, err = r.Read() {
				if err != nil {
					panic(err)
				}
				record := schema.Parse(line)
				// log.Println(record)
				wc <- *record.ToItem()
			}
			wait.Done()
			log.Println("done.")
			sem <- 1
		}()
	}

	log.Println("waiting")
	wait.Wait()
	log.Println("closing")
	close(wc)
	log.Println("goodbye")
}
