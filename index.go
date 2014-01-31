package main

import ()

var cmdIndex = &Command{
	Run: runIndex,
	UsageLine: "index",
	Short: "build the index",
}

func runIndex(cmd *Command, args []string) {
}
