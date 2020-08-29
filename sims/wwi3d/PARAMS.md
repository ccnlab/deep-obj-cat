This is the parameter search notes for wwi3d.

# TODO

* figuring out: ctxt .1 or .2; te/o self?  FwdWeak .1 vs. .05; Lat .1 vs .05; 

# Cur best


# back further

* BackLIPCT = .5 (78) vs. .2 (77) -- main diff is on V2CT hogging -- .5 reduces.., and ALSO has decent TE_V1Sim benefit -- go back to .5..

* BackMax = .1 (80) vs. .05 (77) -- minor inc on V2CT hogging, but better TE_V1Sim, and much better V2P cosdiff

* BackMax = .2 (83) vs. .1 (82): .2 = worse TE Dead, worse TE_V1Sim, slightly better V2P cosdiff

* ctxt te self 1 (82) vs noself (80): TE self = lower TE hog, and lower dead, but *worse* TE_V1Sim (smallish)

# ctxt, back, lat, fwdweak (60-75)

In TE/O, hog++ == dead-- -- hog is more important.

Best overall: 76 = sulat.05 fwdweak.05, ctxt=2te/oSelf, back .05 all, lipct .2

But, ctxt=1 is better!

* 69: sulat.1 init.2, fwdweak.1 fix, ctxt=1noself, FmPulv=.2, back .05 all, lipct .2

Summary: 

* back .1 vs. .05: 74 v 76; 71 vs 70: V2CT hog/dead better with back = .05, V1Sim better, V2p cosdiff worse tho
* lat.1 vs. lat.05: 73 v 74 (fwdweak .05): lat.05 = less TE hog, otherwise very similar
* fwdweak .05 vs. .1: 74 v 75: not much diff -- 74 slightly better for V4 dead

* 71 = ctxt1, 72 = ctxt2, 73 = ctxt2+self: 71 = best V1sim, least TE, TEO dead = best; TEOP cosdif = slightly worse

## hog / dead

TE Hog:

* good: 62, 65, 74, 75, 66
* bad:  70, 69, 67, 71, 72 -- but after 100 or so does converge..

TE Dead:

* good: 62, 65, 69 (resolves), ..
* bad (in the end): 75, 74, 73, 66

TECT Hog:

* best: 73, 74  (but then slightly reversed on dead)
* worst: 68

V2CT Dead:

* best: 66, 68, 67, 65
* med: 69, 70, 62
* worst: 71, 72, 73, 74, 75

V4 Dead:

* best: 65, rest all similar, 75, 73, 74 worst

TEO Dead:

* best: 66, 65 (converges with most in end)
* worst: 72, 73, 74

TEOCT Dead:

* best: 66, 65
* worst: 74, 73, 75 -- but converge 

TEOCT Hog:

* best: 74, 75, 73

## TE V1Sim

* best to worst:  69, 65 .... 75, 74, 73 -- not huge diffs.

* CatDist -- all similar, and similar to others, except 44 is best

## CosDiff

LIPP: all based on fwd pre-fix (up to 68?) / post fix

V2P: best: 72+, next: 69, 70, worst: 68-

TEP: best: 72, 73

TEOP: best: 73+  (in later ticks)

# Notes

* overall, much more hogging across all layers than in cemer version -- dead units are comparable but hard to interpret due to undersampling of edges of input, so hog units are the most important factor to look at. most likely it is lack of constant pinging by changing input in v1 that causes hogging to set in -- that is probably what the v1p inputs did most of all.. 

* weaker ctxt is better in V4/IT -- CT hogging is very bad and context is a major contributor.  new drivers are much more static so need less context.

* pool instead of one2one (p1to1) on V4, TE, TEO CT prjns: in general better for V4, TEO (topo), but not much diff for TE -- maybe slightly better localist?  actually, only better at start then same over time..

* fm pulv: .2 > .1 > 0.05 (major diffs) -- for hogging

* back topdown = .05 > .1 > stronger -- for hogging 

* TEO using topo drivers (i.e., NoTopo = false) is much better for V1sim, dist

* no dwt on tick 0 surprisingly not much effect except LIPP cosdif.
