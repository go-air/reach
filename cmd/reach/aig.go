// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-air/gini/logic/aiger"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
)

var aigCmd = &subCmd{
	Name:  "aig",
	Flags: flag.NewFlagSet("aig", flag.ExitOnError),
	Run:   doAig,
	Init:  initAig,
	Usage: "reach aig [opts] <output>",
	Short: `aig outputs an binary aiger of the Reach internal aig.`,
	Long: `
aig outputs binary aiger file of the aig in the specified output directory.  The resulting file
is the aig of the Reach internal representation of the aig.

By default, the output is written to stdout.
`}

var aigOpts = struct {
	outPath *string
}{}

func initAig(cmd *subCmd) {
	flags := cmd.Flags
	aigOpts.outPath = flags.String("o", "", "output path")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doAig(cmd *subCmd, args []string) {
	flags := cmd.Flags
	flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no output directories specified.\n")
		return
	}
	if flags.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "too many output directories specified.\n")
		return
	}
	if err := doAigArg(cmd, flags.Arg(0)); err != nil {
		log.Printf("error writing aig for '%s': %s\n", flags.Arg(0), err.Error())
	}
}

func doAigArg(cmd *subCmd, arg string) error {
	flags := cmd.Flags
	st, err := os.Stat(arg)
	if os.IsNotExist(err) {
		return err
	}
	if !st.IsDir() {
		flags.SetOutput(os.Stderr)
		flags.Usage()
		fmt.Fprintf(os.Stderr, `cannot output aig, need reach output dir.\n`)
		os.Exit(2)
	}
	return doAigOutput(arg)
}

func doAigOutput(arg string) error {
	out, err := reach.OpenOutput(arg)
	if err != nil {
		return err
	}
	aig, err := out.Aiger()
	if err != nil {
		return err
	}
	w, e := aigWriter()
	if e != nil {
		return e
	}
	var outs []z.Lit
	outs = append(outs, aig.Outputs...)
	outs = append(outs, aig.Bad...)
	trans := aiger.MakeFor(aig.Sys(), outs...)
	return trans.WriteBinary(w)
}

func aigWriter() (io.WriteCloser, error) {
	if *aigOpts.outPath == "-" || *aigOpts.outPath == "" {
		return os.Stdout, nil
	}
	return os.OpenFile(*aigOpts.outPath, os.O_WRONLY|os.O_CREATE, 0644)
}
