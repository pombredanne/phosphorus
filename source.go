package main

import ()

var cmdSource = &Command{
	Run: runSource,
	UsageLine: "source",
	Short: "populate the source table",
}

func runSource(cmd *Command, args []string) {
}
