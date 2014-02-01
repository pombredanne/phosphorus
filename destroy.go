package main

import (
	"fmt"
	"os"
	"willstclair.com/phosphorus/environment"
	"log"
)

var cmdDestroy = &Command{
	Run: runDestroy,
	UsageLine: "destroy",
	Short: "destroy AWS resources",
}

func destroyMsg(resource, disposition string) {
	log.Printf("%s: %s\n", resource, disposition)
}

func destroyErr(resource string, err error) {
	log.Printf("%s: %s\n", resource, err)
}

func runDestroy(cmd *Command, args []string) {
	fmt.Fprintf(os.Stderr, "configuration path: %s (from %s)\n\n", confPath, confFrom)

	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	log.Println("destroying environment...")

	if err := env.IndexTable.Destroy(); err == nil {
		destroyMsg("IndexTable", "destroyed")
	} else {
		destroyErr("IndexTable", err)
	}

	if err := env.IndexBucket.Destroy(); err == nil {
		destroyMsg("IndexBucket", "destroyed")
	} else {
		destroyErr("IndexBucket", err)
	}

	if err := env.SourceTable.Destroy(); err == nil {
		destroyMsg("SourceTable", "destroyed")
	} else {
		destroyErr("SourceTable", err)
	}

	if err := env.SourceBucket.Destroy(); err == nil {
		destroyMsg("SourceBucket", "destroyed")
	} else {
		destroyErr("SourceBucket", err)
	}
}
