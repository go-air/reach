// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-air/reach"
	"github.com/go-air/reach/sim"
)

var simCmd = &subCmd{
	Name:  "sim",
	Flags: flag.NewFlagSet("sim", flag.ExitOnError),
	Run:   doSim,
	Init:  initSim,
	Usage: "reach sim [opts] <aiger>",
	Short: `sim simulates aiger.`,
	Long: `
sim simulates an aiger file with the specified trace.  Simulation does 64
Boolean operations in parallel with a single 64-bit word operation. 

Upon completion, any bad states which were visited will cause sim to create
a trace.  The trace may be incomplete and only contains the last 'win' steps
leading to the bad state.

Reachable bad states have 'Depth' reported as the true number of steps, which
may exceed the trace memory limit.
`}

var simOpts = struct {
	Trace         *bool
	Dur           *time.Duration
	MaxDepth      *int64
	MaxWatchCount *int
	WindowMax     *int
	RestartFactor *int
	Verbose       *bool
	N             *int
	Seed          *int64
}{}

var untilDoc = `"-until n" will limit sim so that it runs at most
until all bad states have been reached n times.`

func initSim(cmd *subCmd) {
	flags := cmd.Flags
	simOpts.Trace = flags.Bool("trace", true, "generate traces.")
	simOpts.Dur = flags.Duration("dur", 30*time.Second, "timeout.")
	simOpts.MaxDepth = flags.Int64("to", 1<<30, "stop after reaching the specified depth (if -restart==0).")
	simOpts.N = flags.Int("n", 1, "repeat n times until stopping condition.")
	simOpts.MaxWatchCount = flags.Int("until", 1, untilDoc)
	simOpts.WindowMax = flags.Int("win", 1024, "memory for trace gen in steps.")
	simOpts.RestartFactor = flags.Int("restart", 0, "restart factor for Luby series restarts (default 0).")
	simOpts.Seed = flags.Int64("seed", 44, "random seed.")
	simOpts.Verbose = flags.Bool("v", false, "verbosity.")
	flags.StringVar(&outDir, "o", ".", "output directory")

	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doSim(cmd *subCmd, args []string) {
	flags := cmd.Flags
	cmd.Flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no aigs specified.\n")
	}
	for i := 0; i < flags.NArg(); i++ {
		if err := doSimArg(flags.Arg(i)); err != nil {
			log.Printf("%s", err)
		}
	}
}

func doSimArg(fn string) error {
	deadLine := time.Now().Add(*simOpts.Dur)
	aig, err := readAiger(fn)
	if err != nil {
		return err
	}
	bad := aigerBad(aig)
	if len(bad) == 0 {
		return fmt.Errorf("ErrNoBads")
	}
	opts := sim.NewOptions()
	opts.WatchUntil = *simOpts.MaxWatchCount
	opts.MaxDepth = int64(*simOpts.MaxDepth)
	opts.Duration = time.Until(deadLine)
	opts.Verbose = *simOpts.Verbose
	opts.N = *simOpts.N
	opts.RestartFactor = *simOpts.RestartFactor
	opts.Seed = *simOpts.Seed

	ck := sim.New(aig.Sys(), bad...)
	ck.SetOptions(opts)
	n := ck.Simulate()
	if opts.Verbose {
		fmt.Printf("[sim] did %d steps for 64 traces\n", n)
	}
	out, err := reach.MakeOutput(fn, outDir)
	if err != nil {
		return err
	}
	ck.FillOutput(out)
	for _, b := range out.Results() {
		fmt.Printf("\t%s\n", b)
	}
	return out.Store()
}
