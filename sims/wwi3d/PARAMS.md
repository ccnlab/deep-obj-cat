This is the parameter search notes for wwi3d.

# TODO

* figuring out: ctxt .1 or .2; te/o self?  FwdWeak .1 vs. .05; Lat .1 vs .05; 

# Cur best

* ore000065: ctxt=1noself, FmPulv=.2, back .05 all, lipct .2 -- minimal hogging in supers, CT still hoggy, but some getting better.  TEO starts out hoggy but gets better.  TE_V1sim very strong.

* but actually 62 maybe better with ctxt2.

# Notes

* overall, much more hogging across all layers than in cemer version -- dead units are comparable but hard to interpret due to undersampling of edges of input, so hog units are the most important factor to look at. most likely it is lack of constant pinging by changing input in v1 that causes hogging to set in -- that is probably what the v1p inputs did most of all.. 

* weaker ctxt is better in V4/IT -- CT hogging is very bad and context is a major contributor.  new drivers are much more static so need less context.

* pool instead of one2one (p1to1) on V4, TE, TEO CT prjns: in general better for V4, TEO (topo), but not much diff for TE -- maybe slightly better localist?  actually, only better at start then same over time..

* fm pulv: .2 > .1 > 0.05 (major diffs) -- for hogging

* back topdown = .05 > .1 > stronger -- for hogging 

* TEO using topo drivers (i.e., NoTopo = false) is much better for V1sim, dist

* no dwt on tick 0 surprisingly not much effect except LIPP cosdif.
