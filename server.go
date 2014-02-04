package main

import (
	"log"
	"strings"
	"net/http"
	"encoding/json"
	// "fmt"
	"bytes"
	"path/filepath"
	"os"
	"willstclair.com/phosphorus/data"
	"willstclair.com/phosphorus/index"
	"willstclair.com/phosphorus/encoder"
	"willstclair.com/phosphorus/environment"
)

var cmdServer = &Command{
	Run: runServer,
	UsageLine: "server",
	Short: "run the match server",
}

func runServer(cmd *Command, args []string) {
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

	// initialize our stupid table (d'oh)
	err = env.IndexTable.Load()
	if err != nil {
		panic(err)
	}

	err = env.SourceTable.Load()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/match", func(w http.ResponseWriter, r *http.Request) {
		var fields []string
		var buf bytes.Buffer
		for _, field := range conf.Index.IndexFields {
			for _, fieldName := range field.Names {
				val := r.FormValue(fieldName)
				if val != "" {
					buf.WriteString(fixDataHack(strings.ToUpper(val)))
				}
			}
			fields = append(fields, buf.String())
			buf.Reset()
		}

		v := enc.Encode(fields)
		sig := temp.Sign(v)

		// ugh
		var keys = make([]uint32, 128)
		for i, _ := range sig {
			keys[i] = sig.Key(i)
		}

		qc := make(chan uint32)
		go env.IndexTable.MultiGet(keys, qc)
		candidates := index.Rank(qc)
		log.Println(candidates)

		var results []*QueryResult

		for _, c := range candidates[:10] {
			record, err := env.SourceTable.Get(c.RecordId)
			if err != nil {
				panic(err)
			}
			results = append(results, &QueryResult{c,record})
		}

		jenc := json.NewEncoder(w)
		jenc.Encode(results)

		// fmt.Fprintf(w, "%s", record)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type QueryResult struct {
	Candidate index.Candidate
	Record map[string]string
}
