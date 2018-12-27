# Reach -- A Finite State Reachability tool for Binary Systems

Reach performs finite state reachability on binary systems.  One
could also say that Reach is a safety model checker.

## Install

Reach is written in [Go](http://golang.org) and requires Go to build/install from
source.  To install Go, please see [the installation webpage](http://golang.org/doc/install).

Then all one needs to do is run

```sh
go get -u github.com/irifrance/reach...
```

This will also get and build our only dependency ([gini](http://github.com/irifrance/gini))
apart from Go itself.

Please have a look at our
[releases](http://github.com/irifrance/reach/releases) for the project
status.  If you would like to have binary distributions, please let us know by the issue
tracker.

## Background

As a software tool, Reach works on transition systems either in
[gini/logic](https://godoc.org/github.com/irifrance/gini/logic) form (for
library use) or in
[aiger](http://fmv.jku.at/aiger/) format (for cli use) which specify sequential
synchronous circuits with Boolean/binary state, functions, and I/O.  

Mathematically, often enough we think of these things as
transition systems `(x,I,T,B)` where

1. `x` is a set of binary variables.
1. `I(x)` is a set of initial states.
1. `T(x,x')` is a transition relation over pairs of states.
1. `B(x)` is a set of bad states (e.g. the complement of a safety property).

`I`, `T`, and `B` above are formulas, so the size of the state space is
`2**|x|`.

The reachability problem is to find if there is a sequence of states
such that the first one satisfies `I(x)`, the last one satisfies `B(x)`,
and every adjacent pair in the sequence satisfies `T(x,x')`.

This problem is PSPACE complete. This means the class of problems represented
by possible inputs to Reach are considered intractable (but atleast
decidable), more so than the easiest intractable complexity class (NP-complete).

## Usage

Reach provides a command line tool `reach` and an associated set of libraries.
Reach uses [gini](http://github.com/irifrance/gini) and has no other dependencies
other than the Go standard library (and a Go compiler for Reach source releases).

Below are some quick command line examples.  For library usage, please consult
both [reach godoc](http://godoc.org/github.com/irifrance/reach) and
[gini/logic](http://godoc.org/github.com/irifrance/gini/logic) for sequential
logic data structures that are designed to work well with a SAT solver.

### Checking

Reach performs checking of reachability with various checkers.

```sh
% # bounded model checking to 100 steps, not exceeding
% # 5 seconds
% reach bmc -to 100 -dur 5s file.aig
% # incremental inductive checking
% reach iic -dur 10s file.aig file2.aig
% # 64 bit parallel random simulation
% reach sim -window 100 file.aig rand # random simulation
```

Every Reach checker places output in a directory containing the results. The
results may be a trace to (or a partial trace leading to) a bad state or states,
or an invariant proving unreachability that can be checked for inductiveness.

### Outputs

Reach can be used to verify outputs and give info about aiger files
and check that invariants prove properties.  Here is an example session.

```sh
⎣ ⇨ reach bmc foo.aig
foo.aig: solved 1
    bad[1]: status=reachable depth=0 dur=0s
⎣ ⇨ reach info foo
aig foo/aig:
    40 latches
    39 inputs
    1199 total
    1 bads
bad[1]: status=reachable depth=0 dur=0s
⎣ ⇨ reach ck foo
[reach] ck foo:
[reach]     verified bad[1]: status=reachable depth=0 dur=0s
```

### Help

Reach also has good built-in CLI help

```sh
⎣ ⇨ reach
Reach is a finite state reachability tool for binary systems.

usage: reach [gopts] <command> [args]

available commands:
    iic    iic is an incremental inductive checker.
    bmc    bmc performs SAT based bounded model checking.
    sim    sim simulates aiger with <trace>.
    info   info provides summary information about an aiger or output.
    ck     ck checks traces and inductive invariants.

global options:
  -cpuprof string
        file to output cpu profile

For help on a command, try "reach <cmd> -h".
```

## Performance

We've developed Reach initially with tip and hwmcc benchmarks. Reach can
solve a lot of these problems quickly and is fairly robust in terms of 
different kinds of inputs.  Reach also uses some unique technology, making
it reasonable to try out on inputs for which other methods (smv, abc, etc)
have problems.  It's proof engine is much faster than baseline IC3, but it
is still a bit behind ABC/PDR on many problems. 

We're always improving Reach and welcome feedback, including performance
problems or reports.

## Related Tools

Interested parties are encouraged to check out the aiger tools, ABC,
SMV (nusmv, nuxmv).  Also, Z3 and some SMT solvers contains code for 
solving reachability problems, although such functionality was not
originally in full scope of the SMT agenda. 


