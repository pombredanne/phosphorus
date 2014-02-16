// Copyright 2014 William H. St. Clair

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	cmdHash,
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
	os.Exit(2)
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(maxprocs)

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	args := flag.Args()

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

func init() {
	flag.BoolVar(&noBanner, "nobanner", false, "do not show banner")
	flag.IntVar(&maxprocs, "p", 2, "go MAXPROCS")
}

func msg(resource, disposition string) {
	log.Printf("%s: %s\n", resource, disposition)
}

func errMsg(resource string, err error) {
	log.Printf("%s: %s\n", resource, err)
}
