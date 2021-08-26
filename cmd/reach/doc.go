// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Command reach provides a cli for binary system reachability.
//  ⎣ ⇨ reach
//  Reach is a finite state reachability tool for binary systems.
//
//  usage: reach [gopts] <command> [args]
//
//  available commands:
//  	iic	iic is an incremental inductive checker.
//  	bmc	bmc performs SAT based bounded model checking.
//  	sim	sim simulates aiger.
//  	ck	ck checks traces and inductive invariants.
//  	stim	stim outputs an aiger stimulus from an output directory.
//  	aag	aag outputs an ascii aiger of the reach internal representation from an output directory.
//  	aig	aig outputs an binary aiger of the Reach internal representation of an aiger.
//  	info	info provides summary information about an aiger or output.
//
//  global options:
//    -cpuprof string
//      	file to output cpu profile
//
//  For help on a command, try "reach <cmd> -h".
//  ⎣ ⇨ reach iic -h
//  reach iic [options] <aiger0> [<aiger1>, ...]
//    -dur duration
//      	timeout (default 30s)
//    -o string
//      	output directory (default ".")
//    -to int
//      	maximum depth (default 1073741824)
//
//  iic runs an incremental inductive checker on the supplied aiger files to find
//  or disprove reachability of bad states. Iic can find and output deep
//  counterexample traces and output inductive invariants as a witness to
//  unreachable bad states.
//
//  iic counterexamples are not necessarily shortest counterexamples. Bad state
//  depths for traces are the trace length itself.  For unknown results, depths
//  represent the depth to which it is known no counterexample trace exists.
//
//  ⎣ ⇨ reach bmc -h
//  reach bmc [opts] <aiger0> <aiger1> ...
//    -dur duration
//      	timeout (default 30s)
//    -o string
//      	output directory (default ".")
//    -to int
//      	maximum depth (default 1073741824)
//
//  bmc does SAT based bounded model checking on aiger files.  Bounded model
//  checking is the most effective way to find or verify the absense of corner case
//  bugs which don't require very many steps of computation.  If no bugs are found,
//  then the depth of the result indicates that there are no reachable bad steps
//  within "depth" steps.
//
//  ⎣ ⇨ reach sim -h
//  reach sim [opts] <aiger>
//    -dur duration
//      	timeout (default 30s)
//    -o string
//      	output directory (default ".")
//    -to int
//      	maximum depth (default 1073741824)
//    -until int
//      	"-until n" will limit sim so that it runs at most
//      	until all bad states have been reached n times. (default 1)
//    -win int
//      	memory for trace gen in steps. (default 1024)
//
//  sim simulates an aiger file with the specified trace.  Simulation does 64
//  Boolean operations in parallel with a single 64-bit word operation.
//
//  Upon completion, any bad states which were visited will cause sim to create
//  a trace.  The trace may be incomplete and only contains the last 'win' steps
//  leading to the bad state.
//
//  Reachable bad states have 'Depth' reported as the true number of steps, which
//  may exceed the trace memory limit.
//
//  ⎣ ⇨ reach ck -h
//  reach ck [opts] <output0> [<output1>, ...]
//    -dur duration
//      	time limit for checking each invariant. (default 5s)
//    -v	verbose, provide more info.
//
//  ck verifies traces and inductive invariants in reach output directories.
//  ck prints out whether or each bad state is verified and any errors.  If
//  there are any bad states which fail verification, then check causes reach
//  to exit with status 1. Otherwise, reach exits with status 0.
//
//  ⎣ ⇨ reach stim -h
//  reach stim [opts] <output>
//    -o string
//      	suffix (after bad.) for aiger stimuli output files.
//
//  stim output saiger stimuli from an output directory.  The output
//  directory should have a .trace file associated with a bad state.
//
//  By default, the output is written to stdout.
//
//  ⎣ ⇨ reach aag -h
//  reach aag [opts] <output>
//    -o string
//      	output path
//
//  aag outputs ascii aiger file of the aig in the specified output directory.  The resulting file
//  is the aag of the Reach internal representation of the aig.
//
//  By default, the output is written to stdout.
//
//  ⎣ ⇨ reach aig -h
//  reach aig [opts] <output>
//    -o string
//      	output path
//
//  aig outputs binary aiger file of the aig in the specified output directory.  The resulting file
//  is the aig of the Reach internal representation of the aig.
//
//  By default, the output is written to stdout.
//
//  ⎣ ⇨ reach info -h
//  reach info [opts] <aiger | output>
//    -v	verbose, provide more info.
//
//  info provides information about an aiger or output directory of reach.
//
package main
