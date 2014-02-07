package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	BANNER = `
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
	Run         func(cmd *Command, args []string)
	Flag        flag.FlagSet
	UsageLine   string
	Short       string
	CustomFlags bool
}

var commands = []*Command{
	cmdSchema,
	cmdIndex,
}

var noBanner bool
var maxprocs int

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	// fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// func usage() {
// 	fmt.Fprintf(os.Stderr, "usage: phosphorus [args] command [args]\n\n")
// 	os.Exit(2)
// }

func main() {
	flag.Parse()
	// flag.Usage = usage

	runtime.GOMAXPROCS(5)

	// log.SetFlags(0)

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	args := flag.Args()
	// if len(args) < 1 {
	// 	usage()
	// }

	if !noBanner {
		fmt.Fprintf(os.Stderr, BANNER)
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			os.Exit(2)
			return
		}
	}
}

// func init() {
// 	flag.BoolVar(&noBanner, "nobanner", false, "do not show banner")
// 	flag.IntVar(&maxprocs, "p", 2, "go MAXPROCS")
// }

func msg(resource, disposition string) {
	log.Printf("%s: %s\n", resource, disposition)
}

func errMsg(resource string, err error) {
	log.Printf("%s: %s\n", resource, err)
}
