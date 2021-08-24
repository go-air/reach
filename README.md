# ⊧ Reach -- A Finite State Reachability Tool

Reach is a symbolic finite state reachability checker.  One could also say that
Reach is a safety model checker.

## Install

Reach is written in [Go](http://golang.org) and requires Go to build/install from
source.  To install Go, please see [the installation webpage](http://golang.org/doc/install).

Then all one needs to do is run

```sh
go install github.com/go-air/reach/...@latest
```

This will also get and build our only dependency ([gini](http://github.com/go-air/gini))
apart from Go itself.

Please have a look at our
[releases](http://github.com/go-air/reach/releases) for the project status.
If you would like to have binary distributions, please let us know by the issue
tracker.

## Background

As a software tool, Reach works on transition systems either in
[gini/logic](https://godoc.org/github.com/go-air/gini/logic#S) form (for
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

## Documentation
The [doc](https://github.com/go-air/reach/tree/master/doc) directory contains some high
level documentation.  [Godoc](https://godoc.org/github.com/go-air/reach) is
available for library reference.


## Usage
```sh
⎣ ⇨ reach
Reach is a finite state reachability tool for binary systems.

usage: reach [gopts] <command> [args]

available commands:
	iic	iic is an incremental inductive checker.
	bmc	bmc performs SAT based bounded model checking.
	sim	sim simulates aiger.
	ck	ck checks traces and inductive invariants.
	stim	stim outputs an aiger stimulus from an output directory.
	aag	aag outputs an ascii aiger of the Reach internal aig.
	aig	aig outputs an binary aiger of the Reach internal aig.
	info	info provides summary information about an aiger or output.

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
have problems.  Its proof engine is much faster than baseline IC3, but it
is still behind ABC/PDR on many problems. 

## Citing Reach

DOI based citations and downloads:
[![DOI](https://zenodo.org/badge/163344002.svg)](https://zenodo.org/badge/latestdoi/163344002)

BibTeX:
```
@misc{scott_cotton_2019_2554423,
  author       = {Scott  Cotton},
  title        = {go-air/reach: Tenu},
  month        = jan,
  year         = 2019,
  doi          = {10.5281/zenodo.2554423},
  url          = {https://doi.org/10.5281/zenodo.2554423}
}
```


