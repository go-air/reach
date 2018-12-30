# Reach Overview

## The Problem
Reach computes the reachability problem in a directed graph represented symbolically
as Boolean formulas/circuits.  

Mathematically, often enough we think of the reachability problem in terms of 
transition systems $(x,I,T,B)$ where

1. $x$ is a set of binary variables.
1. $I(x)$ is a set of initial states.
1. $T(x,x')$ is a transition relation over pairs of states.
1. $B(x)$ is a set of bad states (e.g. the complement of a safety property).

$I$, $T$, and $B$ above are formulas, so the size of the state space is
$2^{|x|}$.

The reachability problem is to find if there is a sequence of states
such that the first one satisfies $I(x)$, the last one satisfies $B(x)$,
and every adjacent pair in the sequence satisfies $T(x,x')$.

This problem is PSPACE complete. This means the class of problems represented
by possible inputs to Reach are considered intractable (but atleast
decidable), more so than the easiest intractable complexity class (NP-complete).

### Symbolic Representation in Reach

In software terms, the Reach library uses the
[gini/logic.S](http://godoc.org/github.com/irifrance/gini/logic#S) structure to
represent graphs.  The CLI uses [the aiger format](http://fmv.jku.at/aiger)
[1].

Both of these are and-inverter graph with `latches`, which correspond to state
variables $x$.  Both of these representations logic.S allow input variables as
well, let us call them $y$. The latch next states are functions of $x$ and $y$,
which can be set and retrieved via logic.S objects. Let us call these
functions, under one name, $\Phi: (x \times y) \to x'$.

The relation $T(x,x')$ is then defined as

$$T(x,x') \circeq \exists y \ . \ x' = \Phi(x,y)$$

Likewise, the bad states are defined as

$$B(x) \circeq \exists y \ . \ b(x, y)$$

where $b$ is a literal (of type github.com/irifrance/gini/z.Lit) found in a 
circuit of type \*github.com/irifrance/gini/logic.S.

The initial states in a circuit are defined by 

$$I(x) \circeq \ . \bigwedge_{m \in M} trans.Init(m)$$

where `trans` is a *logic.S circuit and $M$ denotes the latches whose initial values are 
either `trans.T` or `trans.F`.

### Functional representations

While a bit more involved than the relational symbolic representation above, this
functional representation is common and more compact than the simple one above
since the variables $y$ allow us to express relations over state variables more
easily.

### Why Symbolic Representation?

Symbolic methods are the only feasible methods for solving problems with very large
state spaces, like $2^n$ where $n$ can be in the thousands, or even millions.

Symbolic methods are usually slower than basic graph algorithms when the state
space is small, like $2^{16}$ or so.

Symbolic methods also allow the specification of problems by means of constraints which
is often convenient, especially if the state space is large.

## Methods

Since the symbolic reachability problem is so hard, in practice different
methods are used some which are approximate or solve simpler problems
than the full reachability problem.

The main methods are simulation (approximate), bounded model checking (BMC --
approximate) and full out proving.

### Simulation

Simulation simply runs the transition relation under some given or generated
values.  It is fast and very good at finding how a system tends to behave but
very poor at finding corner cases.

### BMC

Bounded model checking checks reachability, but only for paths up to some
finite length.  If the length is long enough, this suffices for solving
reachability, but it usually is not.

### Proofs

Several different full decision procedures for full reachability which can
prove unreachability or find traces exist.  For proofs, Reach uses a method
based on some relatively recent (circa 2010) developments that have proved
indispensible and fairly widely applicable in practice.  The slides directory
contains a presentation on these methods.

## Organisation

Reach is organised into a library and command line tool.

### Reach Library

The Reach library is intended for use when tight integration is desired, for
customization, tuning, research and experimentation.

The library is [documented](http://godoc.org/github.com/irifrance/reach) on godoc.org.


### Reach CLI

The Reach CLI is intended for easy evaluation of further use of the library,
managing sets of reachability problems, generating invariants and stimuli
reaching bugs and verifying results.

Result verification can be applied to outputs by other tools.

Please note there are some restrictions around the CLI centered around the fact
that Reach uses its own internal logic representation.  In particular, only the
`stim` command works with indices guaranteed to correspond to aiger inputs [1].
We are considering how to improve this situation.

#### CLI Usage

##### Overview

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

##### Proving Unreachability

The `iic` command provides the main mechanism for proving 
unreachability with reach.

```sh
⎣ ⇨ reach iic -h
reach iic [options] <aiger0> [<aiger1>, ...]
  -csift
    	do consecutive sifting. (default true)
  -dur duration
    	timeout. (default 30s)
  -filter
    	filter proof obligations. (default true)
  -justify
    	justify proof obligations. (default true)
  -o string
    	output directory (default ".")
  -pp
    	pre-process aig. (default true)
  -pull
    	do pulling with consecutive sifting. (default true)
  -to int
    	maximum depth. (default 1073741824)
  -v	run with verbosity.

iic runs an incremental inductive checker on the supplied aiger files to find
or disprove reachability of bad states. Iic can find and output deep
counterexample traces and output inductive invariants as a witness to
unreachable bad states.

iic counterexamples are not necessarily shortest counterexamples. Bad state
depths for traces are the trace length itself.  For unknown results, depths
represent the depth to which it is known no counterexample trace exists.
```

##### BMC

```sh
reach bmc -h
reach bmc [opts] <aiger0> <aiger1> ...
  -dur duration
    	timeout (default 30s)
  -o string
    	output directory (default ".")
  -to int
    	maximum depth (default 1073741824)

bmc does SAT based bounded model checking on aiger files.  Bounded model
checking is the most effective way to find or verify the absense of corner case
bugs which don't require very many steps of computation.  If no bugs are found,
then the depth of the result indicates that there are no reachable bad steps
within "depth" steps.
```

##### Simulation

```sh
reach sim -h
reach sim [opts] <aiger>
  -dur duration
    	timeout. (default 30s)
  -n int
    	repeat n times until stopping condition. (default 1)
  -o string
    	output directory (default ".")
  -restart int
    	restart factor for Luby series restarts.
  -seed int
    	random seed. (default 44)
  -to int
    	stop after reaching the specified depth (if -restart==0). (default 1073741824)
  -trace
    	generate traces. (default true)
  -until int
    	"-until n" will limit sim so that it runs at most
    	until all bad states have been reached n times. (default 1)
  -v	verbosity.
  -win int
    	memory for trace gen in steps. (default 1024)

sim simulates an aiger file with the specified trace.  Simulation does 64
Boolean operations in parallel with a single 64-bit word operation.

Upon completion, any bad states which were visited will cause sim to create
a trace.  The trace may be incomplete and only contains the last 'win' steps
leading to the bad state.

Reachable bad states have 'Depth' reported as the true number of steps, which
may exceed the trace memory limit.
```

##### Result checking
Reach can check traces and inductive invariants.

```sh
⎣ ⇨ reach ck -h
reach ck [opts] <output0> [<output1>, ...]
  -dur duration
    	time limit for checking each invariant. (default 5s)
  -v	verbose, provide more info.

ck verifies traces and inductive invariants in reach output directories.
ck prints out whether or each bad state is verified and any errors.  If
there are any bad states which fail verification, then check causes reach
to exit with states 1.
```

##### Output formatting and Aiger stimuli

Reach can format output/result info in various ways.  One way uses the
'info' command.  Another way uses the 'stim' command:

```sh
⎣ ⇨ reach info -h
reach info [opts] <aiger | output>
  -f string
    	format for bad state json (eg -f '{{.FormatStatus}}').
  -v	verbose, provide more info.

info provides information about an aiger or output directory of reach.

⎣ ⇨ reach stim -h
reach stim [opts] <output>
  -o string
    	suffix (after bad.) for aiger stimuli output files.

stim output saiger stimuli from an output directory.  The output
directory should have a .trace file associated with a bad state.

By default, the output is written to stdout.

⎣ ⇨
```

The `stim` command is only applicable when the input problem conforms to the aiger
requirement that all latches have initial state false, since it only ouputs values
for system input variables ($y$ above).


Here is an example which uses both.

```sh
⎣ ⇨ reach bmc ~/tmp/tip/ken.flash\^12.C.aig
/Users/scott/tmp/tip/ken.flash^12.C.aig: solved 1
	bad[2065]: status=reachable depth=3 dur=0s
⎣ ⇨ reach info -v -f '{{.FormatStatus}}' ken.flash\^12.C
ken.flash^12.C/aig reachable
⎣ ⇨ reach stim ken.flash\^12.C
c (aiger) trace for bad[2065]: status=reachable depth=3 dur=0s
100011110001110000000000000000001011111000
010011100100010000110000000000000110100100
000000000000011110101000000010100000100000
000000000000000000000000000000000000000000
.
```

Likewise, the aigsim tool can be used to cross-check results:

```sh
⎣ ⇨ reach stim ken.flash\^12.C | aigsim ken.flash\^12.C/aig
[aigsim] WARNING no properties found, using outputs instead
00000000000000000000000000000000000000000000 100011110001110000000000000000001011111000 0 10001111000111000000000000000000101111100011
10001111000111000000000000000000101111100011 010011100100010000110000000000000110100100 0 01001110010001000011000000000000011010010011
01001110010001000011000000000000011010010011 000000000000011110101000000010100000100000 0 00000000000001111010100000001010000010000011
00000000000001111010100000001010000010000011 000000000000000000000000000000000000000000 1 00000000000000000000000000000000000000000001
Trace is a witness for: { b0 }
```

## Notes and References

[1] Some aiger files still aren't supported at this time, in particular aigers
which allow and gates with constant inputs.

