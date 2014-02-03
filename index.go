package main

import (
	"log"
	"willstclair.com/phosphorus/environment"
)

var cmdIndex = &Command{
	Run: runIndex,
	UsageLine: "index",
	Short: "build the index",
}

func runIndex(cmd *Command, args []string) {
	log.Printf("configuration path: %s (from %s)\n\n", confPath, confFrom)
	log.Println("Running indexer...")

	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
}
