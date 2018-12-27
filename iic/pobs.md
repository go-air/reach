# Proof Obligation Simplifications

This document sketches ideas for simplifying sets of proof 
obligations.

The basic idea is that we want to generate many fewer 
proof obligations.  We've already started this by means
of justification, but that's not good enough.

Instead, we simplify based on sets of proof obligations using
the following rule (dnf dual of resolution)

```
(A and m) or (B and not(m)) => (A or B)
```

We only apply the rule when A => B, 
which is equivalent to saying when B is a subset of A
in terms of a set of literals.

## tracking minLevel and distToBad
Now currently, we keep a priority queue of proof obligations,
and when that is exhausted we propagate.  So we need to keep
track of when we are exhausted.

We thus supplant the rule above to keep the minimum of
the min distance of obligations and also the minimum
of the distToBad.  Then we can know when all obligations
are unsat at the current frame.


## maintaining the set
Every time we generate a new proof obligation, we keep
the parent and the child.

Every time we find that an obligation is inductive, we
no longer need to consider 

## trace reconstruction
ok, keep track




