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

Below we show an example session with some annotations.

#### CLI Usage


## Notes and References

[1] Some aiger files still aren't supported at this time, in particular aigers
which allow and gates with constant inputs.

