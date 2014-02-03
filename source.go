package main

import (
	"log"
	"os"
	"willstclair.com/phosphorus/environment"
)

var cmdSource = &Command{
	Run:       runSource,
	UsageLine: "source",
	Short:     "populate the source table",
}

func runSource(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	log.Println("Loading source data")
}
