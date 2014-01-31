package main

import (
	"path/filepath"
	"strings"
	"errors"
	"fmt"
	"os"
	"flag"
	"log"
	"willstclair.com/phosphorus/config"
)

const (
	CONFIGFILE     = "config.json"
	BANNER          = `
                           )
                          ) \
      . . . . . . . . . ./ ) (. .
       . . . . . . . . . \(_)/ . .
                         / /
                        / /
    Phosphorus         / /
    match server      / /
                     /_/

`
)

type Command struct {
	Run func(cmd *Command, args []string)
	Flag flag.FlagSet
	UsageLine string
	Short string
}

var commands = []*Command{
	cmdEnv,
	cmdPrepare,
	cmdDestroy,
	cmdSource,
	cmdIndex,
}

var confPath string
var confFrom string
var noBanner bool
var conf config.Configuration

const CONFIGENV = "PHOSPHORUS_CONFIG"

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: phosphorus [ -c config.json ] command\n\n")
	os.Exit(2)
}

func main() {
	flag.Parse()
	flag.Usage = usage
	log.SetFlags(0)

	if err := findConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n\n", err.Error())
		os.Exit(1)
	}

	if err := loadConfig(); err != nil {
		switch err.(type) {
		default:
			log.Printf("error loading configuration file: %s\n", err)
		case *config.Error:
			fmt.Printf("phosphorus: invalid configuration file:\n\n")
			for m := range err.(*config.Error).Messages() {
				fmt.Printf("\t%s\n", m)
			}
			fmt.Printf("\n")

		}
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if !noBanner {
		fmt.Fprintf(os.Stderr, BANNER)
	}
	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			// cmd.Flag.Usage = func() { cmd.Usage() }
			// if cmd.CustomFlags {
			// 	args = args[1:]
			// } else {
			// 	cmd.Flag.Parse(args[1:])
			// 	args = cmd.Flag.Args()
			// }
			cmd.Run(cmd, args)
			os.Exit(2)
			return
		}
	}
	usage()
}

func init() {
	flag.StringVar(&confPath, "c", "", "configuration file")
	flag.BoolVar(&noBanner, "nobanner", false, "do not show banner")
}

func findConfig() (err error) {
	if confPath != "" {
		if confPath, err = filepath.Abs(confPath); err != nil {
			return
		}
		confFrom = "commandline"
		return
	}

	confPath = os.Getenv(CONFIGENV)
	if confPath != "" {
		if confPath, err = filepath.Abs(confPath); err != nil {
			return
		}

		confFrom = "environment"
		return
	}

	err = errors.New(fmt.Sprintf("phosphorus: specify a configuration file with $%s or -c",
		CONFIGENV))
	return
}


func loadConfig() (err error) {
	file, err := os.Open(confPath)
	if err != nil {
		return
	}
	defer file.Close()

	conf = config.Configuration{}
	if err = conf.Load(file); err != nil {
		return
	}

	return
}