// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-air/reach"
	"github.com/go-air/reach/iic"
)

var iicCmd = &subCmd{
	Name:  "iic",
	Flags: flag.NewFlagSet("iic", flag.ExitOnError),
	Run:   doIic,
	Init:  initIic,
	Usage: "reach iic [options] <aiger0> [<aiger1>, ...]",
	Short: "iic is an incremental inductive checker.",

	Long: `
iic runs an incremental inductive checker on the supplied aiger files to find
or disprove reachability of bad states. Iic can find and output deep
counterexample traces and output inductive invariants as a witness to
unreachable bad states.

iic counterexamples are not necessarily shortest counterexamples. Bad state
depths for traces are the trace length itself.  For unknown results, depths
represent the depth to which it is known no counterexample trace exists.
`}

var iicOpts = struct {
	Dur            *time.Duration
	MaxDepth       *int
	Verbose        *bool
	Justify        *bool
	ConsecSift     *bool
	ConsecSiftPull *bool
	FilterObs      *bool
	Preprocess     *bool
}{}

func initIic(cmd *subCmd) {
	flags := cmd.Flags
	iicOpts.Dur = flags.Duration("dur", 30*time.Second, "timeout.")
	iicOpts.MaxDepth = flags.Int("to", 1<<30, "maximum depth.")
	iicOpts.Verbose = flags.Bool("v", false, "run with verbosity.")
	iicOpts.Justify = flags.Bool("justify", true, "justify proof obligations.")
	iicOpts.ConsecSift = flags.Bool("csift", true, "do consecutive sifting.")
	iicOpts.ConsecSiftPull = flags.Bool("pull", true, "do pulling with consecutive sifting.")
	iicOpts.FilterObs = flags.Bool("filter", true, "filter proof obligations.")
	iicOpts.Preprocess = flags.Bool("pp", true, "pre-process aig.")
	flags.StringVar(&outDir, "o", ".", "output directory")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doIic(cmd *subCmd, args []string) {
	flags := cmd.Flags
	flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no aigs specified.\n")
	}
	for i := 0; i < flags.NArg(); i++ {
		arg := flags.Arg(i)
		if err := doIicAiger(arg, *iicOpts.Dur); err != nil {
			fmt.Fprintf(os.Stderr, "error doing '%s': %s\n", arg, err)
			continue
		}
	}
}

func doIicAiger(fn string, dur time.Duration) error {
	start := time.Now()
	aig, err := readAiger(fn)
	if err != nil {
		return err
	}
	fmt.Printf("read %s in %s\n", fn, time.Since(start))
	bad := aigerBad(aig)
	if len(bad) == 0 {
		return fmt.Errorf("ErrNoBads")
	}
	trans := aig.S
	for _, b := range bad {
		mc := iic.New(trans, b)
		if *iicOpts.Verbose {
			fmt.Printf("created mc in %s\n", time.Since(start))
		}
		opts := mc.Options()
		opts.Verbose = *iicOpts.Verbose
		opts.Justify = *iicOpts.Justify
		opts.Preprocess = *iicOpts.Preprocess
		opts.Duration = *iicOpts.Dur
		opts.ConsecuSift = *iicOpts.ConsecSift
		opts.ConsecuSiftPull = *iicOpts.ConsecSiftPull
		opts.FilterObs = *iicOpts.FilterObs
		opts.MaxDepth = *iicOpts.MaxDepth

		switch mc.Try() {
		case 1:
			fmt.Printf("%s: cex found.\n", fn)
		case -1:
			fmt.Printf("%s: inv found.\n", fn)
		case 0:
			fmt.Printf("%s: timeout.\n", fn)
		default:
			panic("unreachable")
		}
		out, err := reach.MakeOutput(fn, outDir)
		if err != nil {
			log.Printf("%s: error making output: %s", fn, err)
			continue
		}
		mc.FillOutput(out)
		if err := out.Store(); err != nil {
			log.Printf("error storing output: %s", err)
		}
		fmt.Printf("wrote results in %s.\n", out.RootDir())
	}
	return nil
}
