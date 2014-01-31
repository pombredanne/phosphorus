package main

import ()

var cmdDestroy = &Command{
	Run: runDestroy,
	UsageLine: "destroy",
	Short: "destroy AWS resources",
}

func runDestroy(cmd *Command, args []string) {
}
