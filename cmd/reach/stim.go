// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/irifrance/reach"
)

var stimCmd = &subCmd{
	Name:  "stim",
	Flags: flag.NewFlagSet("stim", flag.ExitOnError),
	Run:   doStim,
	Init:  initStim,
	Usage: "reach stim [opts] <output>",
	Short: `stim outputs an aiger stimulus from an output directory.`,
	Long: `
stim output saiger stimuli from an output directory.  The output 
directory should have a .trace file associated with a bad state.

By default, the output is written to stdout.
`}

var stimOpts = struct {
	outPathSuffix *string
}{}

func initStim(cmd *subCmd) {
	flags := cmd.Flags
	stimOpts.outPathSuffix = flags.String("o", "", "suffix (after bad.) for aiger stimuli output files.")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doStim(cmd *subCmd, args []string) {
	flags := cmd.Flags
	flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no output directories specified.\n")
	}
	if err := doStimArg(cmd, flags.Arg(0)); err != nil {
		log.Printf("error writing stimulus for '%s': %s\n", flags.Arg(0), err.Error())
	}
}

func doStimArg(cmd *subCmd, arg string) error {
	flags := cmd.Flags
	st, err := os.Stat(arg)
	if os.IsNotExist(err) {
		return err
	}
	if !st.IsDir() {
		flags.SetOutput(os.Stderr)
		flags.Usage()
		fmt.Fprintf(os.Stderr, `cannot output stimulus, need reach output dir with a trace.\n`)
		os.Exit(2)
	}
	return doStimOutput(arg)
}

func doStimOutput(arg string) error {
	out, err := reach.OpenOutput(arg)
	if err != nil {
		return err
	}
	for i, b := range out.Results() {
		if *stimOpts.outPathSuffix != "-" && *stimOpts.outPathSuffix != "" {
			fmt.Printf("getting stimulus for %s...", b)
		}
		w, err := stimWriter(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening stim output: %s\n", err.Error())
			continue
		}
		trace, err := out.Trace(i)
		if err != nil {
			w.Close()
			fmt.Fprintf(os.Stderr, "error reading trace: %s\n", err.Error())
			continue
		}
		_, err = trace.EncodeAigerStim(w)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing stim: %s\n", err.Error())
		}
		if w != os.Stdout {
			w.Close()
		}
		if *stimOpts.outPathSuffix != "-" && *stimOpts.outPathSuffix != "" {
			fmt.Printf("wrote output to %s.\n", stimOutPath(b))
		}
	}
	return nil
}

func stimWriter(bad *reach.Result) (io.WriteCloser, error) {
	var f *os.File
	var err error
	if *stimOpts.outPathSuffix == "-" || *stimOpts.outPathSuffix == "" {
		f = os.Stdout
	} else {
		pathName := stimOutPath(bad)
		f, err = os.OpenFile(pathName, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
	}
	if _, err := fmt.Fprintf(f, "c (aiger) trace for %s\n", bad); err != nil {
		return nil, err
	}
	return f, nil
}

func stimOutPath(b *reach.Result) string {
	dir, nm := filepath.Split(*stimOpts.outPathSuffix)

	snm := fmt.Sprintf("bad-%d.%s", b.M, nm)
	return filepath.Join(dir, snm)
}
