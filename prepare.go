package main

import ()

var cmdPrepare = &Command{
	Run: runPrepare,
	UsageLine: "prepare",
	Short: "create AWS resources",
}

func runPrepare(cmd *Command, args []string) {
}

var cmdDestroy = &Command{
	Run: runDestroy,
	UsageLine: "destroy",
	Short: "destroy AWS resources",
}

func runDestroy(cmd *Command, args []string) {
}

var cmdSource = &Command{
	Run: runSource,
	UsageLine: "source",
	Short: "populate the source table",
}

func runSource(cmd *Command, args []string) {
}

var cmdIndex = &Command{
	Run: runIndex,
	UsageLine: "index",
	Short: "build the index",
}

func runIndex(cmd *Command, args []string) {
}
