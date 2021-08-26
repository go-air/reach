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
)

var ckCmd = &subCmd{
	Name:  "ck",
	Flags: flag.NewFlagSet("ck", flag.ExitOnError),
	Run:   doCk,
	Init:  initCk,
	Usage: "reach ck [opts] <output0> [<output1>, ...]",
	Short: `ck checks traces and inductive invariants.`,
	Long: `
ck verifies traces and inductive invariants in reach output directories.  
ck prints out whether or each bad state is verified and any errors.  If 
there are any bad states which fail verification, then check causes reach 
to exit with status 1. Otherwise, reach exits with status 0.
`}

var ckOpts = struct {
	Verbose *bool
	Dur     *time.Duration
}{}

func initCk(cmd *subCmd) {
	flags := cmd.Flags
	ckOpts.Verbose = flags.Bool("v", false, "verbose, provide more info.")
	ckOpts.Dur = flags.Duration("dur", 5*time.Second, "time limit for checking each invariant.")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doCk(cmd *subCmd, args []string) {
	flags := cmd.Flags
	flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no output directories specified.\n")
	}
	hasErr := false
	for _, arg := range flags.Args() {
		out, err := reach.OpenOutput(arg)
		if err != nil {
			fmt.Printf("error opening '%s': %s\n", arg, err)
			hasErr = true
			continue
		}
		log.Printf("check %s:\n", arg)
		for i, bad := range out.Results() {
			if !out.IsVerifiable(i) {
				fmt.Printf("\t%s: nothing to check\n", bad)
				continue
			}
			if errs := out.TryVerifyResult(i, *ckOpts.Dur); len(errs) != 0 {
				hasErr = true
				for _, e := range errs {
					fmt.Printf("\terror verifying %s: %s\n", bad, e)
				}
				hasErr = true
			} else {
				fmt.Printf("\tverified %s\n", bad)
			}
		}
	}
	if hasErr {
		os.Exit(1)
	}
}
