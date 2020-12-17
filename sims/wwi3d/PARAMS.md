This is the parameter search notes for wwi3d.

# TODO

* Add weaker surround projection from same CT -> P

* Add back higher drivers at some point!


# Summary up to 180 pre major TRC, CT bug fixes

* load pretrained LIP -- beneficial

* non CTCtxt for self prjn in self -- not good

# Summary as of Job 159

Key changes:

* CTCtxt prjns all 1to1 -- sig helps with hogging, no worse at cosdiff pred err.
    
    + CT self prjns pool1to1 scale = 1 seem beneficial at higher layers -- trying elsewhere

* weaker top-down prjns to CT layers reduces hogging

* TRC AvgMix = .5 for higher, more pooled layers (V3+); V2P fine with 0 or .2 -- no real diff

# Jobs 140..: original ~8instance non-plus 20obj, more V2CT, V3CT de-hogging

* Now only V2CT, V3CT remain majorly hoggy -- all others good(ish)..

* LIPCT -> V2CT = .1 reduces V2CT hogging significantly.

* trying reduced top-down to V2CT, V3CT -- beneficial at .05

* V3 FmPulv = .05 reduces hogging, but also sig impairs cosiff -- stay at .1; V2 needs .1

* try ctxt1to1init .2, .3 -- weaker drivers at start -- not better, some things worse.


# Jobs 108..139: fixing CT hogging, adding lots of missing DP connections

* All super layers now mostly hog-free, but all CT layers exhibit significant hogging..

* all Pulv layers ONLY using V1 drivers for time being -- these are key de-hoggers and allow some degree of comparison across layers for predictive accuracy.

* added other inputs directly into pulv to help with prediction, as in cemer version

* projections into DP were missing.  as of 128 DPP also just using V1 drivers -- now doing something -- V3 alone was NOT good, and comment in cemer version said it was bad there too..

    + but getting DPP learning has no effect on anything else!

* TEO/TE both benefit from NoTopo

* one-to-one ctxt prjns and 4 rel strength (as in orig model) -- key for TEO / V4 hog reduction and prediction, but kinda weird!

* AvgMax .5 vs .2 in V3+: drives sig better CosDiff pred in TEO, TE, better (lower) TE_V1sim but *worse* CatDist -- fits less well with lba5 cats..

* AvgMax .5 in V2 impairs CosDif (128,129), but slightly improves V2CT hogging

* TEO -> V2CT improves V2 CosDif (128)

## 1to1 ctxt prjns

* 1to1 super -> CT renders CT as just a delayed copy of Super: main learning is then in the pooled CT -> pulv prjn, and in whatever super can do to help, but CT doesn't provide any extra mojo.

    + maybe giving CT too much rope just hangs it?  anatomically, it is more deep <-> deep and super -> CT more microcolumn?

    + test whether pooled CT <-> CT are doing something in TE/O -- and add to V4 / V2

* V3, DP self ctxt p1to1 1sc = worse V3CT hogging, worse cosdiff, 

* TE/O CT hogging takes much longer to onset than in V3 etc

* TE/O self off or 1 = much less hogging.


# Jobs 77..83: back further

* BackLIPCT = .5 (78) vs. .2 (77) -- main diff is on V2CT hogging -- .5 reduces.., and ALSO has decent TE_V1Sim benefit -- go back to .5..

* BackMax = .1 (80) vs. .05 (77) -- minor inc on V2CT hogging, but better TE_V1Sim, and much better V2P cosdiff

* BackMax = .2 (83) vs. .1 (82): .2 = worse TE Dead, worse TE_V1Sim, slightly better V2P cosdiff

* ctxt te self 1 (82) vs noself (80): TE self = lower TE hog, and lower dead, but *worse* TE_V1Sim (smallish)

# Jobs 60..75: ctxt, back, lat, fwdweak

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
