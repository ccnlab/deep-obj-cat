This is the parameter search notes for wwi3d.

# TODO

# 300+: 4x4 TEO, 2x2 TE, removing pulv cons

* projections from other pulv layers are generally not useful -- generally CT <-> P all in same layer

* *some* other CT -> P is definitely useful for cosdif performance:

    + V3CT -> V2P, V4CT -> V2P (todo: test each separately)
    + V2CT -> V4P (Fwd)
    + V4CT -> TEOP (Fwd), TECT -> TEOP (todo: test each separately)
    + TEOCT -> TEP (Fwd)
    
    + .2 had no cosdif improvement vs. .1 -- try lower

* general lessons: everything shows basic categorization effects, just differs in extent to which the TE remains more or less "tethered" to the bottom-up V1 structure -- when less connected, it produces lower-dimensional (typically 2) category structure, which tends to be more binary / discrete.

* not too much actually affects the Pulv cosdiff performance -- definitely try to optimize that -- some changes change the geometry of drivers, so that has an effect, but otherwise, not much.

* Lateral unit-to-unit cons do seem to make a diff!

* synaptic noise in leabrax actually seems to have a very gradual effect..


# 229 regularized

* restarted with "rationalized" parameters consistent across all layers -- didn't work very well..

* 230, 231: V2,3 S->CT 2 vs 1 -- very high V1Sim with 2, but also higher TE CatDist with 2 -- not good tradeoff.  

* 231, 233: got rid of full TEOCT->V2 @ .1, TEO->V2CT @.1, improved TE CatDist, no change in V1Sim.  inconsistent with 229 removing these prjns results.

* overall 233 is pretty good for TE CatDist, but not as clean overall as 229

* 234 v 233: CTback, S To CT both .2 -- both CatDst and V1Sim lower (tradeoff)
* 235 v 234: BackMax (.5) -> BackStrong (.2) = higher CatDst & V1Sim (tradeoff)
* 236 v 235: S to CT = .1 vs .2 -- worse CatDst, same later V1Sim (init higher) -- **S to CT important!**

* 240 v 236: BackMax -> .1, CT self V2,3,4 .1 v. 5 -- no real diffs (transient V1Sim higher)
* 241 v 240: CT Self V2,3,4 .2 v .1 -- higher V1Sim, same CatDst -- **low CT stronger = more V1Sim**
* 242 v 240: V4 CT Self = 1 -- better CatDst and V1Sim -- **V4 CT self better!**
* 243 v 242: again CT low self worse V1Sim (same 241 v. 240)
* 244 v 242: incr SToCT .2 vs. .1 made V1Sim sig worse, small incr in CatDst
* 244 v 233: 233 is sig better on V1Sim, similar on CatDst -- go back to 233
* 245 v 244: V4 -> V3 .1 v .2, TE -> V4 .2 v .1 -- reduces V1Sim -- prob from TE -> V4
* 246 v 245: CTback  .1 v .2 -- worse CatDst same V1Sim, **CT back important**
* 245 v 233: 233 better CatDst, same V1Sim

# 229 tweaks

* 250: Weaker LIP -> V2 (.2 vs .5) sig lowered V1Sim, increased CatDst -- **weaker LIP -> V2 better**
* 249: Weaker V3->V2 (.2 vs .5) sig *increased* V1Sim -- **keep V3->V2 strong**
* 251: Stronger V2,3P -> .2 v .1 CatDst drops off, V1Sim MUCH lower.  **not clear about V2p .1**

* 253: full TEO->V2 sig worse than partial (prelim)
* 254: no TEOCT->V2 -- sig worse CatDist -- CT more important than TEO
* 255: no TEO->V2CT -- not that much worse

# 229 = 223 + Low CT Self .5

* best so far -- after full training, dist matrix is very similar to original -- TE_CatDst high and TE_V1Sim low (both around .5 -- want CatDst higher and V1Sim lower)
 

# 223 best case

* TE, TEO 2x2, V1V4 driving TE, TEO with NoTopo=true

* TE, TEO CT <-> CT = 1 (self) > 2, 3

* TE/TEO -> V4 = .2 > .1

* StdFB = .1 > .05


# Summary up to 220 after CTFmSuper one-to-one, no learn throughout

* WtInit.Mean = 0.8 > 0.5 with lower Rel = 2 instead of 4

* TE, TEO = 2x2 instead of 4x4 == same perf basically and much faster

* TE, TEO CT <-> CT = 1 instead of 4

* V4 CT <-> CT = 0.5 is fine

* then try lower (V2, V3) selfs -- standard should be self with pool connectivity


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

# Current Best WtScale Listing

## 229

```
Layer: V1m

Layer: V1h

Layer: LIP
	              LIPPToLIP		Abs:	1	Rel:	0.2
	             MTPosToLIP		Abs:	1	Rel:	0.5
	            EyePosToLIP		Abs:	1	Rel:	1
	           SacPlanToLIP		Abs:	1	Rel:	1
	            ObjVelToLIP		Abs:	1	Rel:	1
	                V2ToLIP		Abs:	1	Rel:	0.1
	                V3ToLIP		Abs:	1	Rel:	0.1

Layer: LIPCT
	             LIPToLIPCT		Abs:	1	Rel:	1
	            LIPPToLIPCT		Abs:	1	Rel:	0.2
	          EyePosToLIPCT		Abs:	1	Rel:	1
	         SaccadeToLIPCT		Abs:	1	Rel:	1
	          ObjVelToLIPCT		Abs:	1	Rel:	1
	            V2CTToLIPCT		Abs:	1	Rel:	0.1
	            V3CTToLIPCT		Abs:	1	Rel:	0.1

Layer: LIPP
	            LIPCTToLIPP		Abs:	1	Rel:	1

Layer: MTPos
	             V1mToMTPos		Abs:	1	Rel:	1

Layer: EyePos

Layer: SacPlan

Layer: Saccade

Layer: ObjVel

Layer: V2
	                V2PToV2		Abs:	1	Rel:	0.2
	                V1mToV2		Abs:	1	Rel:	1
	                V1hToV2		Abs:	1	Rel:	1
	                 V4ToV2		Abs:	1	Rel:	0.1
	                 V3ToV2		Abs:	1	Rel:	0.5
	                LIPToV2		Abs:	1	Rel:	0.5
	              TEOCTToV2		Abs:	1	Rel:	0.1
	                 V2ToV2		Abs:	1	Rel:	0.02

Layer: V2CT
	               V2ToV2CT		Abs:	1	Rel:	1
	              V2PToV2CT		Abs:	1	Rel:	0.1
	             V2CTToV2CT		Abs:	1	Rel:	0.5
	            LIPCTToV2CT		Abs:	1	Rel:	0.5
	             V3CTToV2CT		Abs:	1	Rel:	0.5
	             V4CTToV2CT		Abs:	1	Rel:	0.5
	               V3ToV2CT		Abs:	1	Rel:	0.5
	              TEOToV2CT		Abs:	1	Rel:	0.5

Layer: V2P
	              V2CTToV2P		Abs:	1	Rel:	1
	              V3CTToV2P		Abs:	1	Rel:	0.1
	              V4CTToV2P		Abs:	1	Rel:	0.1

Layer: V3
	                V3PToV3		Abs:	1	Rel:	0.2
	                 V2ToV3		Abs:	0.5	Rel:	2
	                 DPToV3		Abs:	1	Rel:	0.2
	                 V4ToV3		Abs:	1	Rel:	0.2
	                LIPToV3		Abs:	1	Rel:	0.1
	                TEOToV3		Abs:	1	Rel:	0.1
	              TEOCTToV3		Abs:	1	Rel:	0.1
	                 V3ToV3		Abs:	1	Rel:	0.02

Layer: V3CT
	               V3ToV3CT		Abs:	1	Rel:	1
	              V3PToV3CT		Abs:	1	Rel:	0.1
	             V3CTToV3CT		Abs:	1	Rel:	0.5
	            LIPCTToV3CT		Abs:	1	Rel:	0.2
	             DPCTToV3CT		Abs:	1	Rel:	0.2
	             V4CTToV3CT		Abs:	1	Rel:	0.2
	               DPToV3CT		Abs:	1	Rel:	0.2
	               V4ToV3CT		Abs:	1	Rel:	0.2

Layer: V3P
	              V3CTToV3P		Abs:	1	Rel:	1
	              DPCTToV3P		Abs:	1	Rel:	0.1

Layer: DP
	                DPPToDP		Abs:	1	Rel:	0.2
	                 V3ToDP		Abs:	2	Rel:	0.5
	                 V2ToDP		Abs:	1	Rel:	1
	                V3PToDP		Abs:	1	Rel:	0.2
	                TEOToDP		Abs:	1	Rel:	0.1

Layer: DPCT
	               DPToDPCT		Abs:	1	Rel:	2
	              DPPToDPCT		Abs:	1	Rel:	0.2
	            TEOCTToDPCT		Abs:	1	Rel:	0.1
	              V3PToDPCT		Abs:	1	Rel:	0.2

Layer: DPP
	              DPCTToDPP		Abs:	1	Rel:	1

Layer: V4
	                V4PToV4		Abs:	1	Rel:	0.2
	                 V2ToV4		Abs:	0.5	Rel:	2
	                TEOToV4		Abs:	1	Rel:	0.2
	                 TEToV4		Abs:	1	Rel:	0.2
	                 V4ToV4		Abs:	1	Rel:	0.02

Layer: V4CT
	               V4ToV4CT		Abs:	1	Rel:	2
	              V4PToV4CT		Abs:	1	Rel:	0.2
	             V4CTToV4CT		Abs:	1	Rel:	0.5
	            TEOCTToV4CT		Abs:	1	Rel:	0.2
	             TECTToV4CT		Abs:	1	Rel:	0.2
	              TEOToV4CT		Abs:	1	Rel:	0.2

Layer: V4P
	              V4CTToV4P		Abs:	1	Rel:	1
	             TEOCTToV4P		Abs:	1	Rel:	0.1
	              V2CTToV4P		Abs:	1	Rel:	0.1

Layer: TEO
	              TEOPToTEO		Abs:	1	Rel:	0.2
	                V4ToTEO		Abs:	2	Rel:	0.5
	                TEToTEO		Abs:	1	Rel:	0.1

Layer: TEOCT
	             TEOToTEOCT		Abs:	1	Rel:	2
	            TEOPToTEOCT		Abs:	1	Rel:	0.2
	           TEOCTToTEOCT		Abs:	1	Rel:	1
	            TECTToTEOCT		Abs:	1	Rel:	0.1
	             V4PToTEOCT		Abs:	1	Rel:	0.2
	             TEPToTEOCT		Abs:	1	Rel:	0.2

Layer: TEOP
	            TEOCTToTEOP		Abs:	1	Rel:	1
	             V4CTToTEOP		Abs:	1	Rel:	0.1
	             TECTToTEOP		Abs:	1	Rel:	0.1

Layer: TE
	                TEPToTE		Abs:	1	Rel:	0.2
	                TEOToTE		Abs:	1.5	Rel:	0.667

Layer: TECT
	               TEToTECT		Abs:	1	Rel:	2
	              TEPToTECT		Abs:	1	Rel:	0.2
	             TECTToTECT		Abs:	1	Rel:	1
	            TEOCTToTECT		Abs:	1	Rel:	0.1
	              V4PToTECT		Abs:	1	Rel:	0.2
	             TEOPToTECT		Abs:	1	Rel:	0.2

Layer: TEP
	              TECTToTEP		Abs:	1	Rel:	1
	             TEOCTToTEP		Abs:	1	Rel:	0.1
```

## 233

```
Layer: V1m

Layer: V1h

Layer: LIP
	              LIPPToLIP		Abs:	1	Rel:	0.2
	             MTPosToLIP		Abs:	1	Rel:	0.5
	            EyePosToLIP		Abs:	1	Rel:	1
	           SacPlanToLIP		Abs:	1	Rel:	1
	            ObjVelToLIP		Abs:	1	Rel:	1
	                V2ToLIP		Abs:	1	Rel:	0.1
	                V3ToLIP		Abs:	1	Rel:	0.1

Layer: LIPCT
	             LIPToLIPCT		Abs:	1	Rel:	1
	            LIPPToLIPCT		Abs:	1	Rel:	0.2
	          EyePosToLIPCT		Abs:	1	Rel:	1
	         SaccadeToLIPCT		Abs:	1	Rel:	1
	          ObjVelToLIPCT		Abs:	1	Rel:	1
	            V2CTToLIPCT		Abs:	1	Rel:	0.1
	            V3CTToLIPCT		Abs:	1	Rel:	0.1

Layer: LIPP
	            LIPCTToLIPP		Abs:	1	Rel:	1

Layer: MTPos
	             V1mToMTPos		Abs:	1	Rel:	1

Layer: EyePos

Layer: SacPlan

Layer: Saccade

Layer: ObjVel

Layer: V2
	                V2PToV2		Abs:	1	Rel:	0.2
	                V1mToV2		Abs:	1	Rel:	1
	                V1hToV2		Abs:	1	Rel:	1
	                 V4ToV2		Abs:	1	Rel:	0.1
	                 V3ToV2		Abs:	1	Rel:	0.5
	                 V2ToV2		Abs:	1	Rel:	0.02
	                LIPToV2		Abs:	1	Rel:	0.5
    vs. 229: TEOCTToV2 Rel .1

Layer: V2CT
	               V2ToV2CT		Abs:	1	Rel:	1
	              V2PToV2CT		Abs:	1	Rel:	0.2
	             V2CTToV2CT		Abs:	1	Rel:	0.5
	            LIPCTToV2CT		Abs:	1	Rel:	0.1 <- all .5 in 229
	             V3CTToV2CT		Abs:	1	Rel:	0.1
	             V4CTToV2CT		Abs:	1	Rel:	0.1
	               V3ToV2CT		Abs:	1	Rel:	0.1

Layer: V2P
	              V2CTToV2P		Abs:	1	Rel:	1
	              V3CTToV2P		Abs:	1	Rel:	0.1
	              V4CTToV2P		Abs:	1	Rel:	0.1

Layer: V3
	                V3PToV3		Abs:	1	Rel:	0.2
	                 V2ToV3		Abs:	0.5	Rel:	2
	                 DPToV3		Abs:	1	Rel:	0.2
	                 V3ToV3		Abs:	1	Rel:	0.02
	                 V4ToV3		Abs:	1	Rel:	0.2
	                LIPToV3		Abs:	1	Rel:	0.1
	                TEOToV3		Abs:	1	Rel:	0.1
	              TEOCTToV3		Abs:	1	Rel:	0.1

Layer: V3CT
	               V3ToV3CT		Abs:	1	Rel:	1
	              V3PToV3CT		Abs:	1	Rel:	0.2
	             V3CTToV3CT		Abs:	1	Rel:	0.5
	            LIPCTToV3CT		Abs:	1	Rel:	0.1 <- these all .2 in 229
	             DPCTToV3CT		Abs:	1	Rel:	0.1
	             V4CTToV3CT		Abs:	1	Rel:	0.1
	               DPToV3CT		Abs:	1	Rel:	0.1
	               V4ToV3CT		Abs:	1	Rel:	0.1

Layer: V3P
	              V3CTToV3P		Abs:	1	Rel:	1
	              DPCTToV3P		Abs:	1	Rel:	0.1

Layer: DP
	                DPPToDP		Abs:	1	Rel:	0.2
	                 V3ToDP		Abs:	2	Rel:	0.5
	                 V2ToDP		Abs:	1	Rel:	1
	                TEOToDP		Abs:	1	Rel:	0.1
	                V3PToDP		Abs:	1	Rel:	0.2

Layer: DPCT
	               DPToDPCT		Abs:	1	Rel:	2
	              DPPToDPCT		Abs:	1	Rel:	0.2
	            TEOCTToDPCT		Abs:	1	Rel:	0.1
	              V3PToDPCT		Abs:	1	Rel:	0.2

Layer: DPP
	              DPCTToDPP		Abs:	1	Rel:	1

Layer: V4
	                V4PToV4		Abs:	1	Rel:	0.2
	                 V2ToV4		Abs:	0.5	Rel:	2
	                TEOToV4		Abs:	1	Rel:	0.2
	                 V4ToV4		Abs:	1	Rel:	0.02
	                 TEToV4		Abs:	1	Rel:	0.1

Layer: V4CT
	               V4ToV4CT		Abs:	1	Rel:	2
	              V4PToV4CT		Abs:	1	Rel:	0.2
	             V4CTToV4CT		Abs:	1	Rel:	0.5
	            TEOCTToV4CT		Abs:	1	Rel:	0.1 <- these all .2 in 229
	             TECTToV4CT		Abs:	1	Rel:	0.1
	              TEOToV4CT		Abs:	1	Rel:	0.1

Layer: V4P
	              V4CTToV4P		Abs:	1	Rel:	1
	             TEOCTToV4P		Abs:	1	Rel:	0.1
	              V2CTToV4P		Abs:	1	Rel:	0.1

Layer: TEO
	              TEOPToTEO		Abs:	1	Rel:	0.2
	                V4ToTEO		Abs:	2	Rel:	0.5
	                TEToTEO		Abs:	1	Rel:	0.1

Layer: TEOCT
	             TEOToTEOCT		Abs:	1	Rel:	2
	            TEOPToTEOCT		Abs:	1	Rel:	0.2
	           TEOCTToTEOCT		Abs:	1	Rel:	1
	            TECTToTEOCT		Abs:	1	Rel:	0.1
	             V4PToTEOCT		Abs:	1	Rel:	0.2
	             TEPToTEOCT		Abs:	1	Rel:	0.2

Layer: TEOP
	            TEOCTToTEOP		Abs:	1	Rel:	1
	             V4CTToTEOP		Abs:	1	Rel:	0.1
	             TECTToTEOP		Abs:	1	Rel:	0.1

Layer: TE
	                TEPToTE		Abs:	1	Rel:	0.2
	                TEOToTE		Abs:	1.5	Rel:	0.667

Layer: TECT
	               TEToTECT		Abs:	1	Rel:	2
	              TEPToTECT		Abs:	1	Rel:	0.2
	             TECTToTECT		Abs:	1	Rel:	1
	            TEOCTToTECT		Abs:	1	Rel:	0.1
	              V4PToTECT		Abs:	1	Rel:	0.2
	             TEOPToTECT		Abs:	1	Rel:	0.2

Layer: TEP
	              TECTToTEP		Abs:	1	Rel:	1
	             TEOCTToTEP		Abs:	1	Rel:	0.1
```

