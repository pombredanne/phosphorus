package main

import (
	"log"
	"os"
	"willstclair.com/phosphorus/environment"
)

var cmdDestroy = &Command{
	Run:       runDestroy,
	UsageLine: "destroy",
	Short:     "destroy AWS resources",
}

func runDestroy(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	log.Println("Destroying environment")

	if err := env.IndexTable.Destroy(); err == nil {
		msg("IndexTable", "destroyed")
	} else {
		errMsg("IndexTable", err)
	}

	if err := env.IndexBucket.Destroy(); err == nil {
		msg("IndexBucket", "destroyed")
	} else {
		errMsg("IndexBucket", err)
	}

	if err := env.SourceTable.Destroy(); err == nil {
		msg("SourceTable", "destroyed")
	} else {
		errMsg("SourceTable", err)
	}

	if err := env.SourceBucket.Destroy(); err == nil {
		msg("SourceBucket", "destroyed")
	} else {
		errMsg("SourceBucket", err)
	}
}
