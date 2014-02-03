package main

import (
	"log"
	"os"
	"willstclair.com/phosphorus/environment"
)

var cmdThroughput = &Command {
	Run: runThroughput,
	UsageLine: "throughput -t ( Source | Index ) ( -w 100 | -r 100 )",
	Short: "Change table throughput to the target level",
}

func runThroughput(cmd *Command, args []string) {
	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	defer env.Cleanup()

	// goamz doesn't support UpdateTable.
}
