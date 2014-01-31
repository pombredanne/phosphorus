package main

import (
	"os"
	"log"
	"fmt"
	"willstclair.com/phosphorus/config"
	"github.com/crowdmob/goamz/s3"
)

var cmdPrepare = &Command{
	Run: runPrepare,
	UsageLine: "prepare",
	Short: "create AWS resources",
}

func runPrepare(cmd *Command, args []string) {
	fmt.Fprintf(os.Stderr, "Preparing the environment...\n")

	env, err := config.NewEnvironment(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	// index table
	if env.IndexTable == nil {
		log.Println("Creating: Index table")
		_, err := env.DynamoServer.CreateTable(
			*config.TableD(conf.Index.Table.Name, "s",
				conf.Index.Table.ReadCapacityUnits,
				conf.Index.Table.WriteCapacityUnits))

		if err != nil {
			log.Println(err)
			os.Exit(2)
		}

		log.Println("Created: Index table")
	}

	// source table
	if env.SourceTable == nil {
		log.Println("Creating: Source table")
		_, err := env.DynamoServer.CreateTable(
			*config.TableD(conf.Source.Table.Name, "r",
				conf.Source.Table.ReadCapacityUnits,
				conf.Source.Table.WriteCapacityUnits))
		if err != nil {
			log.Println(err)
			os.Exit(2)
		}

		log.Println("Created: Source table")
	}

	// index bucket
	exists, err := config.BucketExists(env.IndexBucket)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	if !exists {
		log.Println("Creating: Index bucket")
		err = env.IndexBucket.PutBucket(s3.Private)
		if err != nil {
			log.Println(err)
			os.Exit(2)
		}
		log.Println("Created: Index bucket")
	}

	// source bucket
	exists, err = config.BucketExists(env.SourceBucket)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	if !exists {
		log.Println("Creating: Source bucket")
		err = env.SourceBucket.PutBucket(s3.Private)
		if err != nil {
			log.Println(err)
			os.Exit(2)
		}
		log.Println("Created: Source bucket")
	}

	log.Println("OK")
}
