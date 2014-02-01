package main

import (
	"os"
	"log"
	"fmt"
	"willstclair.com/phosphorus/environment"
)

var cmdPrepare = &Command{
	Run: runPrepare,
	UsageLine: "prepare",
	Short: "create AWS resources",
}

func createMsg(resource, disposition string) {
	log.Printf("%s: %s\n", resource, disposition)
}

func createErr(resource string, err error) {
	log.Printf("%s: %s\n", resource, err)
}

func runPrepare(cmd *Command, args []string) {
	fmt.Fprintf(os.Stderr, "configuration path: %s (from %s)\n\n", confPath, confFrom)

	fmt.Fprintf(os.Stderr, "Preparing the environment...\n")

	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	if err := env.IndexTable.Create(); err == nil {
		createMsg("IndexTable", "created")
	} else {
		createErr("IndexTable", err)
	}

	if err := env.IndexBucket.Create(); err == nil {
		createMsg("IndexBucket", "created")
	} else {
		createErr("IndexBucket", err)
	}

	if err := env.SourceTable.Create(); err == nil {
		createMsg("SourceTable", "created")
	} else {
		createErr("SourceTable", err)
	}

	if err := env.SourceBucket.Create(); err == nil {
		createMsg("SourceBucket", "created")
	} else {
		createErr("SourceBucket", err)
	}

	log.Println("OK")
}
