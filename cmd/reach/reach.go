// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
)

var outDir = "."

var subCmds = [...]*subCmd{
	iicCmd,
	bmcCmd,
	simCmd,
	ckCmd,
	stimCmd,
	aagCmd,
	aigCmd,
	infoCmd}

// returns global argument list
func splitArgs(args []string) ([]string, []string) {
	var i int
	var arg string
	var found bool
	for i, arg = range args {
		found = false
		for _, sc := range subCmds {
			if arg == sc.Name {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	return args[:i], args[i:]
}

var reachFlags = flag.NewFlagSet("reach", flag.ExitOnError)
var pprofAddr = reachFlags.String("cpuprof", "", "file to output cpu profile")

var doc = `Reach is a finite state reachability tool for binary systems.

usage: reach [gopts] <command> [args]`

func usage(w io.Writer) {
	fmt.Fprintln(w, doc)
	fmt.Fprintf(w, "\navailable commands:\n")
	for _, cmd := range subCmds {
		fmt.Fprintf(w, "\t%s\t%s\n", cmd.Name, cmd.Short)
	}
	fmt.Fprintf(w, "\nglobal options:\n")
	reachFlags.SetOutput(w)
	reachFlags.PrintDefaults()
	fmt.Fprintf(w, "\nFor help on a command, try \"reach <cmd> -h\".\n")
}

func init() {
	for _, c := range subCmds {
		c.Init(c)
	}
}

func main() {
	log.SetPrefix("[reach] ")
	log.SetFlags(0)
	gargs, largs := splitArgs(os.Args[1:])
	reachFlags.Usage = func() {
		usage(os.Stderr)
	}
	reachFlags.Parse(gargs)
	if len(largs) == 0 {
		usage(os.Stderr)
		os.Exit(1)
	}
	if *pprofAddr != "" {
		f, e := os.Create(*pprofAddr)
		if e != nil {
			log.Fatalf("couldn't create file %s", *pprofAddr)
		}
		if e := pprof.StartCPUProfile(f); e != nil {
			log.Fatalf("couldn't start pprof: %s", e)
		}
		defer pprof.StopCPUProfile()
	}
	sub := largs[0]
	var theCmd *subCmd
	for _, c := range subCmds {
		if c.Name == sub {
			theCmd = c
			break
		}
	}
	if theCmd == nil {
		fmt.Fprintf(os.Stderr, "unknown command: '%s'\n", largs[0])
		usage(os.Stderr)
		os.Exit(1)
	}
	theCmd.Run(theCmd, largs[1:])
}
