package main

import (
	"fmt"
	"os"
	"willstclair.com/phosphorus/config"
	"github.com/crowdmob/goamz/s3"
	"log"
)

const (
	TPL_NOTABLE = "%s table: <does not exist>\n"
	TPL_TABLE = "%s table:\n\tStatus: %s\n\tItems:  %d\n\tReadCapacityUnits: %d\n\tWriteCapacityUnits: %d\n\tSize: %d B\n\n"

	TPL_NOBUCKET = "%s bucket: <does not exist>\n"
	TPL_BUCKET = "%s bucket: exists\n"
)

var cmdEnv = &Command{
	Run: runEnv,
	UsageLine: "env",
	Short: "print information about your environment",
}

func runEnv(cmd *Command, args []string) {
	fmt.Fprintf(os.Stderr, "configuration path: %s (from %s)\n\n", confPath, confFrom)

	env, err := config.NewEnvironment(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	if env.IndexTable != nil {
		fmt.Fprintf(os.Stderr, TPL_TABLE, "Index",
			env.IndexTableD.TableStatus,
			env.IndexTableD.ItemCount,
			env.IndexTableD.ProvisionedThroughput.ReadCapacityUnits,
			env.IndexTableD.ProvisionedThroughput.WriteCapacityUnits,
			env.IndexTableD.TableSizeBytes)
	} else {
		fmt.Fprintf(os.Stderr, TPL_NOTABLE, "Index")
	}

	if env.SourceTable != nil {
		fmt.Fprintf(os.Stderr, TPL_TABLE, "Source",
 			env.SourceTableD.TableStatus,
			env.SourceTableD.ItemCount,
			env.SourceTableD.ProvisionedThroughput.ReadCapacityUnits,
			env.SourceTableD.ProvisionedThroughput.WriteCapacityUnits,
			env.SourceTableD.TableSizeBytes)
	} else {
		fmt.Fprintf(os.Stderr, TPL_NOTABLE, "Source")
	}

	_, err = env.IndexBucket.List(env.IndexPrefix, "/", "", 10)
	if err != nil && err.(*s3.Error).Code == "NoSuchBucket" {
		fmt.Fprintf(os.Stderr, TPL_NOBUCKET, "Index")
	} else {
		fmt.Fprintf(os.Stderr, TPL_BUCKET, "Index")
	}

	_, err = env.SourceBucket.List(env.SourcePrefix, "/", "", 10)
	if err != nil && err.(*s3.Error).Code == "NoSuchBucket" {
		fmt.Fprintf(os.Stderr, TPL_NOBUCKET, "Source")
	} else {
		fmt.Fprintf(os.Stderr, TPL_BUCKET, "Source")
	}
}
