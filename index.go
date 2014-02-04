package main

import (
	"log"
	"io"
	"sync"
	"encoding/csv"
	"bytes"
	// "runtime/pprof"
	"os"
	"path"
	"path/filepath"
	"willstclair.com/phosphorus/data"
	"willstclair.com/phosphorus/encoder"
	"willstclair.com/phosphorus/environment"
	"willstclair.com/phosphorus/index"
)

var cmdIndex = &Command{
	Run:       runIndex,
	UsageLine: "index",
	Short:     "prepare the index encoder and hash template",
}

func runIndex(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	// set up source field descriptions
	var schema data.SourceSchema
	for _, field := range conf.Source.SourceFields {
		schema = append(schema, data.Attribute{
			field.Name, field.Column, field.ShortName})
	}

	// set up index field descriptions
	var indexSchema [][]string
	for _, field := range conf.Index.IndexFields {
		var shortNames []string
		for _, name := range field.Names {
			for _, sourceField := range schema {
				if name == sourceField.Name {
					shortNames = append(shortNames, sourceField.ShortName)
					break
				}
			}
		}
		indexSchema = append(indexSchema, shortNames)
	}

	log.Println("Running indexer")

	// create a channel for filehandles from S3
	c := make(chan io.ReadCloser)
	go func() {
		env.SourceBucket.OpenAll(c)
		close(c)
	}()


	// limit the total number of simultaneous downloads
	sem := make(chan int, 5)
	for i := 0; i < 5; i++ {
		sem <- 1
	}

	ctr := &encoder.Counter{}
	// count each record
	// var wait sync.WaitGroup
	for rc := range c {
		rc := rc
		// <-sem
		// wait.Add(1)
		// go func() {
		log.Println("starting...")
		defer rc.Close()
		r := csv.NewReader(rc)
		for line, err := r.Read(); err != io.EOF; line, err = r.Read() {
			if err != nil {
				panic(err)
			}
			record := schema.Parse(line)
			ctr.Count(flatten(indexSchema, record.Fields))
		}

		// wait.Done()
		log.Println("done.")
			// sem <- 1
		// }()
	}

	log.Println("waiting")
	// wait.Wait()

	log.Println("creating encoder")
	enc := encoder.NewEncoder(ctr)
	enc.Path = filepath.Join(env.TempDir, "encoder")
	err = enc.Save()
	if err != nil {
		panic(err)
	}

	log.Println("saving encoder")
	err = env.IndexBucket.Put(enc.Path, "encoder")
	if err != nil {
		panic(err)
	}


	log.Println("generating hash family")
	hashDir := filepath.Join(env.TempDir, "hash")
	err = os.Mkdir(hashDir, os.FileMode(os.FileMode(0777) | os.ModeDir))
	if err != nil {
		panic(err)
	}

	temp := &index.Template{
		Directory: hashDir,
		Dimension: enc.Dimension,
	}

	temp.Generate()

	log.Println("saving hash family")

	hashFiles, err := filepath.Glob(filepath.Join(hashDir, "hash_*"))
	if err != nil {
		panic(err)
	}

	for _, hashFile := range hashFiles {
		log.Printf("\tsaving %s\n", hashFile)
		err = env.IndexBucket.Put(hashFile, path.Base(hashFile))
		if err != nil {
			panic(err)
		}
	}

	log.Println("goodbye")

}

func flatten(indexSchema [][]string, keys map[string]string) (fields []string) {
	var buf bytes.Buffer
	for _, indexFields := range indexSchema {
		for _, shortName := range indexFields {
			buf.WriteString(fixDataHack(keys[shortName]))
		}
		fields = append(fields, buf.String())
		buf.Reset()
	}
	return
}

// don't want to reformat my data right now so here's a hack
func fixDataHack(s string) string {
	if len(s) == 1 && s[0] >= 48 && s[0] <= 57 {
		return "0" + s
	}
	return s
}


var cmdIndexData = &Command{
	Run:       runIndexData,
	UsageLine: "indexdata",
	Short:     "populate the index with our data",
}

func runIndexData(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	// set up source field descriptions
	var schema data.SourceSchema
	for _, field := range conf.Source.SourceFields {
		schema = append(schema, data.Attribute{
			field.Name, field.Column, field.ShortName})
	}

	// set up index field descriptions
	var indexSchema [][]string
	for _, field := range conf.Index.IndexFields {
		var shortNames []string
		for _, name := range field.Names {
			for _, sourceField := range schema {
				if name == sourceField.Name {
					shortNames = append(shortNames, sourceField.ShortName)
					break
				}
			}
		}
		indexSchema = append(indexSchema, shortNames)
	}

	log.Println("Running indexer")

	// create a channel for filehandles from S3
	c := make(chan io.ReadCloser)
	go func() {
		env.SourceBucket.OpenAll(c)
		close(c)
	}()


	// limit the total number of simultaneous downloads
	sem := make(chan int, 5)
	for i := 0; i < 5; i++ {
		sem <- 1
	}

	// download and restore our encoder
	log.Println("Downloading encoder")
	encoderPath := filepath.Join(env.TempDir, "encoder")
	err = env.IndexBucket.Get("encoder", encoderPath)
	if err != nil {
		panic(err)
	}

	log.Println("Creating encoder")
	enc := &encoder.Encoder{Path: encoderPath}
	err = enc.Load()
	if err != nil {
		panic(err)
	}

	// download and restore our hash template
	log.Println("Downloading hash template")
	hashDir := filepath.Join(env.TempDir, "hash")
	err = os.Mkdir(hashDir, os.FileMode(os.FileMode(0777) | os.ModeDir))
	if err != nil {
		panic(err)
	}

	err = env.IndexBucket.GetAll("hash", hashDir)
	if err != nil {
		panic(err)
	}

	log.Println("Loading hash template")
	temp := &index.Template{
		Directory: hashDir,
		Dimension: enc.Dimension}

	temp.Load()
	log.Println("Loaded hash template")

	// get our flush channel
	wc := env.IndexTable.AddChannel()
	xr := index.NewIndex(64, wc)

	// initialize our stupid table (d'oh)
	err = env.IndexTable.Load()
	if err != nil {
		panic(err)
	}

	var wait sync.WaitGroup
	for rc := range c {
		rc := rc
		<-sem
		wait.Add(1)
		go func() {
			log.Println("starting...")
			defer func() {
				rc.Close()
				wait.Done()
				log.Println("done.")
				sem <- 1
			}()

			r := csv.NewReader(rc)
			for line, err := r.Read(); err != io.EOF; line, err = r.Read() {
				if err != nil {
					panic(err)
				}
				record := schema.Parse(line)
				v := enc.Encode(flatten(indexSchema, record.Fields))
				sig := temp.Sign(v)
				xr.Add(record.RecordId, sig)
			}

		}()
	}

	log.Println("waiting")
	wait.Wait()
	log.Println("flushing")
	// file, err := os.Create("/home/ubuntu/phosphorus.prof")
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()
	// pprof.StartCPUProfile(file)
	xr.FlushAll()
	close(wc)
	log.Println("goodbye")
	// pprof.StopCPUProfile()
}
