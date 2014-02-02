package main

import ()

var cmdSource = &Command{
	Run: runSource,
	UsageLine: "source",
	Short: "populate the source table",
}

func runSource(cmd *Command, args []string) {
	log.Printf("configuration path: %s (from %s)\n\n", confPath, confFrom)
	log.Println("Loading source data...")

	env, err := environment.New(conf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

}
