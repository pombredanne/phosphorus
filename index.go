package main

import (
	"log"
	"os"
	"willstclair.com/phosphorus/environment"
)

var cmdIndex = &Command{
	Run: runIndex,
	UsageLine: "index",
	Short: "build the index",
}

func runIndex(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	log.Println("Running indexer")

}
