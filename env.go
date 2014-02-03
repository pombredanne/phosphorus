package main

import (
	"fmt"
	"os"
	"willstclair.com/phosphorus/environment"
	"log"
)

const (
	TPL_NOTABLE = "%s table: <does not exist>\n"
	TPL_TABLE = "%s table:\n\tStatus: %s\n\tItems:  %d\n\tReadCapacityUnits: %d\n\tWriteCapacityUnits: %d\n\tSize: %d B\n\n"

	TPL_NOBUCKET = "%s bucket: <does not exist>\n"
	TPL_BUCKET = "%s bucket: exists\n"
)

var cmdEnv = &Command{
	Run: runEnv,
	UsageLine: "env",
	Short: "print information about your environment",
}

func runEnv(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	log.Println("Environment info\n")

	if exists, err := env.IndexTable.Exists(); err == nil && exists {
		fmt.Fprintf(os.Stderr, "IndexTable: exists\n")
	} else if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "IndexTable: does not exist\n")
	}

	if exists, err := env.IndexBucket.Exists(); err == nil && exists {
		fmt.Fprintf(os.Stderr, "IndexBucket: exists\n")
	} else if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "IndexBucket: does not exist\n")
	}

	if exists, err := env.SourceTable.Exists(); err == nil && exists {
		fmt.Fprintf(os.Stderr, "SourceTable: exists\n")
	} else if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "SourceTable: does not exist\n")
	}

	if exists, err := env.SourceBucket.Exists(); err == nil && exists {
		fmt.Fprintf(os.Stderr, "SourceBucket: exists\n")
	} else if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "SourceBucket: does not exist\n")
	}
}
