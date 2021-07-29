// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-air/reach"
	"github.com/go-air/reach/bmc"
)

var bmcCmd = &subCmd{
	Name:  "bmc",
	Flags: flag.NewFlagSet("bmc", flag.ExitOnError),
	Run:   doBmc,
	Init:  initBmc,
	Usage: "reach bmc [opts] <aiger0> <aiger1> ...",
	Short: `bmc performs SAT based bounded model checking.`,
	Long: `
bmc does SAT based bounded model checking on aiger files.  Bounded model
checking is the most effective way to find or verify the absense of corner case
bugs which don't require very many steps of computation.  If no bugs are found,
then the depth of the result indicates that there are no reachable bad steps
within "depth" steps.
`}

var bmcOpts = struct {
	Dur      *time.Duration
	MaxDepth *int
}{}

func initBmc(cmd *subCmd) {
	flags := cmd.Flags
	bmcOpts.Dur = flags.Duration("dur", 30*time.Second, "timeout")
	bmcOpts.MaxDepth = flags.Int("to", 1<<30, "maximum depth")
	flags.StringVar(&outDir, "o", ".", "output directory")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doBmc(cmd *subCmd, args []string) {
	flags := cmd.Flags
	flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no aigs specified.\n")
	}
	for i := 0; i < flags.NArg(); i++ {
		arg := flags.Arg(i)
		if err := doBmcAiger(arg, *bmcOpts.Dur, *bmcOpts.MaxDepth); err != nil {
			fmt.Fprintf(os.Stderr, "error doing '%s': %s\n", arg, err)
			continue
		}
	}
}

func doBmcAiger(fn string, dur time.Duration, to int) error {
	deadLine := time.Now().Add(dur)
	aig, err := readAiger(fn)
	if err != nil {
		return err
	}
	bad := aigerBad(aig)
	if len(bad) == 0 {
		return fmt.Errorf("ErrNoBads")
	}
	mc := bmc.New(aig.S, bad...)
	mc.SetMaxDepth(to)
	n := mc.Try(time.Until(deadLine))
	fmt.Printf("%s: solved %d\n", fn, n)
	out, err := reach.MakeOutput(fn, outDir)
	if err != nil {
		return err
	}
	mc.FillOutput(out)
	for _, b := range out.Results() {
		fmt.Printf("\t%s\n", b)
	}
	return out.Store()
}
