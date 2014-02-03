package main

import (
	"os"
	"log"
	"willstclair.com/phosphorus/environment"
)

var cmdPrepare = &Command{
	Run: runPrepare,
	UsageLine: "prepare",
	Short: "create AWS resources",
}

func runPrepare(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	log.Println("Creating environment")

	if err := env.IndexTable.Create(); err == nil {
		msg("IndexTable", "created")
	} else {
		errMsg("IndexTable", err)
	}

	if err := env.IndexBucket.Create(); err == nil {
		msg("IndexBucket", "created")
	} else {
		errMsg("IndexBucket", err)
	}

	if err := env.SourceTable.Create(); err == nil {
		msg("SourceTable", "created")
	} else {
		errMsg("SourceTable", err)
	}

	if err := env.SourceBucket.Create(); err == nil {
		msg("SourceBucket", "created")
	} else {
		errMsg("SourceBucket", err)
	}
}
