// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
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

var aagCmd = &subCmd{
	Name:  "aag",
	Flags: flag.NewFlagSet("aag", flag.ExitOnError),
	Run:   doAag,
	Init:  initAag,
	Usage: "reach aag [opts] <output>",
	Short: `aag outputs an ascii aiger of the Reach internal aig.`,
	Long: `
aag outputs ascii aiger file of the aig in the specified output directory.  The resulting file
is the aag of the Reach internal representation of the aig.

By default, the output is written to stdout.
`}

var aagOpts = struct {
	outPath *string
}{}

func initAag(cmd *subCmd) {
	flags := cmd.Flags
	aagOpts.outPath = flags.String("o", "", "output path")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doAag(cmd *subCmd, args []string) {
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
	if err := doAagArg(cmd, flags.Arg(0)); err != nil {
		log.Printf("error writing aag for '%s': %s\n", flags.Arg(0), err.Error())
	}
}

func doAagArg(cmd *subCmd, arg string) error {
	flags := cmd.Flags
	st, err := os.Stat(arg)
	if os.IsNotExist(err) {
		return err
	}
	if !st.IsDir() {
		flags.SetOutput(os.Stderr)
		flags.Usage()
		fmt.Fprintf(os.Stderr, `cannot output aag, need reach output dir.\n`)
		os.Exit(2)
	}
	return doAagOutput(arg)
}

func doAagOutput(arg string) error {
	out, err := reach.OpenOutput(arg)
	if err != nil {
		return err
	}
	aig, err := out.Aiger()
	if err != nil {
		return err
	}
	w, e := aagWriter()
	if e != nil {
		return e
	}
	var outs []z.Lit
	outs = append(outs, aig.Outputs...)
	outs = append(outs, aig.Bad...)
	trans := aiger.MakeFor(aig.Sys(), outs...)
	return trans.WriteAscii(w)
}

func aagWriter() (io.WriteCloser, error) {
	if *aagOpts.outPath == "-" || *aagOpts.outPath == "" {
		return os.Stdout, nil
	}
	return os.OpenFile(*aagOpts.outPath, os.O_WRONLY|os.O_CREATE, 0644)
}
