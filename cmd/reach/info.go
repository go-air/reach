// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/irifrance/gini/logic/aiger"
	"github.com/irifrance/reach"
)

var infoCmd = &subCmd{
	Name:  "info",
	Flags: flag.NewFlagSet("info", flag.ExitOnError),
	Run:   doInfo,
	Init:  initInfo,
	Usage: "reach info [opts] <aiger | output>",
	Short: `info provides summary information about an aiger or output.`,
	Long: `
info provides information about an aiger or output directory of reach.
`}

var infoOpts = struct {
	Verbose *bool
	Format  *string
}{}

func initInfo(cmd *subCmd) {
	flags := cmd.Flags
	infoOpts.Verbose = flags.Bool("v", false, "verbose, provide more info.")
	infoOpts.Format = flags.String("f", "", "format for bad state json.")
	flags.Usage = func() {
		fmt.Println(cmd.Usage)
		flags.PrintDefaults()
		fmt.Println(cmd.Long)
	}
}

func doInfo(cmd *subCmd, args []string) {
	flags := cmd.Flags
	flags.Parse(args)
	if flags.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "no aigs or output directories specified.\n")
	}
	for _, arg := range flags.Args() {
		if err := doInfoArg(cmd, arg); err != nil {
			log.Printf("error doing '%s': %s", arg, err.Error())
		}
	}
}

func doInfoArg(cmd *subCmd, arg string) error {
	st, err := os.Stat(arg)
	if os.IsNotExist(err) {
		return err
	}
	if st.IsDir() {
		return doInfoOutput(arg)
	}
	return doInfoAig(arg)
}

func doInfoOutput(arg string) error {
	out, err := reach.OpenOutput(arg)
	if err != nil {
		return err
	}
	var tmpl *template.Template
	if *infoOpts.Format != "" {
		tmpl, err = template.New("reach").Parse(*infoOpts.Format)
		if err != nil {
			return fmt.Errorf("invalid template: %s\n", err)
		}
	}
	for _, b := range out.Results() {
		if *infoOpts.Verbose {
			fmt.Printf("%s ", out.AigerPath())
		}

		if *infoOpts.Format == "" {
			fmt.Printf("%s\n", b)
			continue
		}
		if err = tmpl.Execute(os.Stdout, b); err != nil {
			return fmt.Errorf("unable to execute template: %s", err.Error())
		}
		fmt.Printf("\n")
	}
	return nil
}

func doInfoAig(arg string) error {
	f, e := os.Open(arg)
	if e != nil {
		return e
	}
	defer f.Close()
	aig, err := aiger.ReadBinary(f)
	if err != nil {
		return err
	}
	fmt.Printf("aig %s:\n", arg)
	fmt.Printf("\t%d latches\n\t%d inputs\n\t%d total\n\t%d bads\n", len(aig.Sys().Latches),
		len(aig.Inputs), aig.Sys().Len(), len(aigerBad(aig)))
	return nil
}
