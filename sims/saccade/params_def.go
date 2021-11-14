// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "github.com/emer/emergent/params"

// ParamSets is the default set of parameters -- Base is always applied, and others can be optionally
// selected to apply on top of that
var ParamSets = params.Sets{
	{Name: "Base", Desc: "these are the best params", Sheets: params.Sheets{
		"Network": &params.Sheet{
			// layer classes, specifics
			{Sel: "Layer", Desc: "needs some special inhibition and learning params",
				Params: params.Params{
					"Layer.Inhib.Inhib.AvgTau":     "30",   // 30 > 20 >> 1 definitively
					"Layer.Act.Dt.IntTau":          "40",   // 40 > 20
					"Layer.Inhib.Layer.Gi":         "1.1",  // general default
					"Layer.Inhib.Pool.Gi":          "1.1",  // general default
					"Layer.Inhib.ActAvg.LoTol":     "1.1",  // no low adapt
					"Layer.Inhib.ActAvg.AdaptRate": "0.2",  // 0.5 default
					"Layer.Inhib.ActAvg.Init":      "0.06", // .06 = sigma .2, .04 = sigma .15, .02 = sigma .1
					"Layer.Inhib.ActAvg.Targ":      "0.06",
					"Layer.Inhib.Pool.FFEx0":       "0.18",
					"Layer.Inhib.Pool.FFEx":        "0.05", // .05 makes big diff on Top5
					"Layer.Inhib.Layer.FFEx0":      "0.18",
					"Layer.Inhib.Layer.FFEx":       "0.05", // .05 best so far
					"Layer.Act.Gbar.L":             "0.2",  // 0.2 now best
					"Layer.Act.Decay.Act":          "0.2",  // 0 best
					"Layer.Act.Decay.Glong":        "0.6",  // 0.5 > 0.2
					"Layer.Act.KNa.Fast.Max":       "0.1",  // fm both .2 worse
					"Layer.Act.KNa.Med.Max":        "0.2",  // 0.2 > 0.1 def
					"Layer.Act.KNa.Slow.Max":       "0.2",  // 0.2 > higher
					"Layer.Act.Noise.Dist":         "Gaussian",
					"Layer.Act.Noise.Mean":         "0.0",     // .05 max for blowup
					"Layer.Act.Noise.Var":          "0.01",    // .01 a bit worse
					"Layer.Act.Noise.Type":         "NoNoise", // off for now
					"Layer.Act.GTarg.GeMax":        "1.2",     // 1.2 > 1 > .8 -- rescaling not very useful.
					"Layer.Act.Dt.LongAvgTau":      "20",      // 20 > 50 > 100
					"Layer.Learn.TrgAvgAct.On":     "false",   // not relevant to topo-driven layers, makes a diff
				}},
			{Sel: ".CT", Desc: "CT gain factor is key",
				Params: params.Params{
					"Layer.CtxtGeGain":      "0.3", // .25 > .2 > .15 > .1 > .05
					"Layer.Inhib.Layer.Gi":  "1.1",
					"Layer.Act.KNa.On":      "true",
					"Layer.Act.NMDA.Gbar":   "0.03",
					"Layer.Act.GABAB.Gbar":  "0.2",
					"Layer.Act.Decay.Act":   "0.0", // 0 better v91
					"Layer.Act.Decay.Glong": "0.0", // 0 both better than std
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.DriveScale":  "0.15", // .15 >= .1 > .05
					"Layer.Act.NMDA.Gbar":   "0.03",
					"Layer.Act.GABAB.Gbar":  "0.2", //
					"Layer.Act.Decay.Act":   "0.5", // 0.5 actually better
					"Layer.Act.Decay.Glong": "1",   // 1 better
				}},
			{Sel: "SuperLayer", Desc: "burst params don't really matter",
				Params: params.Params{
					"Layer.Burst.ThrRel": "0.1", // not big diffs
					"Layer.Burst.ThrAbs": "0.1",
				}},
			{Sel: ".V1f", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.2",
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.ActAvg.Init": "0.06", // .06 = sigma .2, .04 = sigma .15, .02 = sigma .1
					"Layer.Inhib.ActAvg.Targ": "0.06",
				}},
			{Sel: ".PopIn", Desc: "pop-code input",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.2",
					"Layer.Inhib.ActAvg.Init": "0.06", // .06 = sigma .2, .04 = sigma .15, .02 = sigma .1
					"Layer.Inhib.ActAvg.Targ": "0.06",
				}},
			{Sel: "#MDe", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.4", // 1.3 > 1.4 > 1.2 > 1.1
					"Layer.Act.Clamp.Ge":      "0.4", // 0.4 > 0.6 > 0.8?
					"Layer.Inhib.ActAvg.Init": "0.04",
					"Layer.Inhib.ActAvg.Targ": "0.04",
				}},
			{Sel: ".LIP", Desc: "high, pool inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":               "1.4", // no perf diff, lower act
					"Layer.Inhib.Pool.Gi":                "0.8",
					"Layer.Inhib.Pool.On":                "false", // false > true -- no pools!
					"Layer.Inhib.ActAvg.Init":            "0.1",   // was .03, actual .1
					"Layer.Inhib.ActAvg.Targ":            "0.1",
					"Layer.Learn.TrgAvgAct.TrgRange.Min": "0.5",
					"Layer.Learn.TrgAvgAct.TrgRange.Max": "2.0",   // reducing does not help anything
					"Layer.Learn.TrgAvgAct.Pool":         "false", // pool sizes too small for trgs!
				}},
			{Sel: "#LIPCT", Desc: "strong inhib",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.Layer.Gi":    "1.2", // 1.2 > 1.4 major
					"Layer.Inhib.Pool.Gi":     "0.8", // 0.8 key for 2x2, 1.0 works for 4x4 but no diff
					"Layer.Inhib.ActAvg.Init": "0.06",
					"Layer.Inhib.ActAvg.Targ": "0.06",
				}},
			// {Sel: "#LIP", Desc: "strong inhib",
			// 	Params: params.Params{
			// 		"Layer.Inhib.ActAvg.Init": "0.06",
			// 		"Layer.Inhib.ActAvg.Targ": "0.06",
			// 	}},
			{Sel: "#S1eP", Desc: "strong inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":       "1.2",
					"Layer.Inhib.Pool.On":        "false",
					"Layer.Inhib.ActAvg.Init":    "0.06",
					"Layer.Inhib.ActAvg.Targ":    "0.06",
					"Layer.Inhib.ActAvg.AdaptGi": "true", // gets overly active
				}},
			{Sel: "#LIPPS", Desc: "strong inhib",
				Params: params.Params{
					"Layer.Inhib.Pool.On": "false",
				}},
			{Sel: "#FEF", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":       "1.2", // 1.2 > 1.3 > 1.1
					"Layer.Inhib.Pool.Gi":        "0.9",
					"Layer.Inhib.Pool.On":        "false", // full layer best
					"Layer.Inhib.ActAvg.Init":    "0.2",
					"Layer.Inhib.ActAvg.Targ":    "0.1",
					"Layer.Learn.TrgAvgAct.Pool": "false", // pool sizes too small for trgs!
				}},
			{Sel: ".SEF", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.1",
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.ActAvg.Init": "0.06",
					"Layer.Inhib.ActAvg.Targ": "0.06",
				}},

			// prjn classes, specifics
			{Sel: "Prjn", Desc: "yes extra learning factors",
				Params: params.Params{
					"Prjn.Learn.Lrate.Base":     "0.1",   // .1 with RLrate on seems good
					"Prjn.PrjnScale.ScaleLrate": "0.5",   // 2 = fast response, effective
					"Prjn.PrjnScale.LoTol":      "0.8",   // good now...
					"Prjn.PrjnScale.AvgTau":     "500",   // slower default
					"Prjn.PrjnScale.Adapt":      "false", // adapt bad maybe?  put GeMax at 1.2, adjust to avoid
					"Prjn.SWt.Adapt.On":         "true",  // true > false, esp in cosdiff
					"Prjn.SWt.Adapt.Lrate":      "0.1",   // trying faster
					"Prjn.SWt.Adapt.SigGain":    "6",
					"Prjn.SWt.Adapt.DreamVar":   "0.0", // 0.02 good in lvis
					"Prjn.SWt.Init.SPct":        "1",   // 1 > lower
					"Prjn.SWt.Init.Mean":        "0.5", // .5 > .4 -- key, except v2?
					"Prjn.SWt.Limit.Min":        "0.2", // .2-.8 == .1-.9; .3-.7 not better
					"Prjn.SWt.Limit.Max":        "0.8", //
				}},
			{Sel: "CTCtxtPrjn", Desc: "defaults for CT Ctxt prjns",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1",
				}},
			{Sel: ".Fixed", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.Learn.Learn":     "false",
					"Prjn.PrjnScale.Adapt": "false", // key to not adapt!
					"Prjn.SWt.Init.Mean":   "0.8",   // 0.8 better
					"Prjn.SWt.Init.Var":    "0",
					"Prjn.SWt.Init.Sym":    "false",
				}},
			{Sel: ".Forward", Desc: "std feedforward",
				Params: params.Params{
					// "Prjn.PrjnScale.Abs": "0.8", // weaker?
				}},
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates -- smaller as network gets bigger",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.2",
				}},
			{Sel: ".Inhib", Desc: "inhibitory projection",
				Params: params.Params{
					"Prjn.Learn.Learn":      "true",   // learned decorrel is good
					"Prjn.Learn.Lrate.Base": "0.0001", // .0001 > .001 -- slower better!
					"Prjn.SWt.Init.Var":     "0.0",
					"Prjn.SWt.Init.Mean":    "0.1",
					"Prjn.SWt.Adapt.On":     "false",
					"Prjn.PrjnScale.Abs":    "0.1", // .1 = .2, slower blowup
					"Prjn.PrjnScale.Adapt":  "false",
					"Prjn.IncGain":          "1", // .5 def
				}},

			{Sel: ".FwdWeak", Desc: "weak feedforward",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1", // .1 orig -- had a bug tho!! also trying .05
				}},

			{Sel: ".FmLIP", Desc: "no random weights here",
				Params: params.Params{
					"Prjn.SWt.Init.Var": "0.25", // was 0 -- trying .05
					"Prjn.SWt.Init.Sym": "false",
				}},
			{Sel: ".BackStrong", Desc: "stronger",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.2", // .1 > orig .2 > .05 -- not sep fm BackMax -- .1 = better TE_V1Sim, V2P cosdiff
				}},
			{Sel: ".BackMax", Desc: "strongest",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.2", // .1 > .2, orig .5 -- see BackStrong
				}},

			{Sel: ".CTToPulv", Desc: "CT to pulvinar needs to be weaker in general, like most prjns",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.PrjnScale.Rel": "1",
					"Prjn.SWt.Init.Var":  "0.25",
				}},
			{Sel: ".BackToPulv", Desc: "top-down to pulvinar directly",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1",
				}},
			{Sel: ".FwdToPulv", Desc: "feedforward to pulvinar directly",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1",
				}},

			{Sel: ".FmPulv", Desc: "default for pulvinar",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.PrjnScale.Rel": "0.2", // .2 > .1 > .05 still true
				}},
			{Sel: ".Lateral", Desc: "default for lateral",
				Params: params.Params{
					"Prjn.SWt.Init.Sym":  "false",
					"Prjn.PrjnScale.Rel": "0.02", // .02 > .05 == .01 > .1  -- very minor diffs on TE cat
					"Prjn.SWt.Init.Mean": "0.5",
					"Prjn.SWt.Init.Var":  "0",
				}},

			{Sel: ".CTFmSuper", Desc: "CT from main super -- fixed one2one",
				Params: params.Params{
					"Prjn.SWt.Init.Mean": "0.5",  // 0.5 with var at 5x5
					"Prjn.SWt.Init.Var":  "0.25", // 0.25
					"Prjn.PrjnScale.Rel": "1",    // def 2
				}},
			{Sel: ".CTFmSuperLower", Desc: "CT from main super -- for lower layers",
				Params: params.Params{
					"Prjn.SWt.Init.Mean": "0.8", // 0.8 makes a diff for lower too, more V1 divergence at .5
					"Prjn.PrjnScale.Rel": "1",   // 1 maybe better
				}},
			{Sel: ".CTSelfLIP", Desc: "CT to CT for LIP",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "2", // 2 = 3 > 1
				}},
			{Sel: ".CTBack", Desc: "CT to CT back (top-down)",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".2", // .2 > .1 in std
				}},
			{Sel: ".SToCT", Desc: "higher Super to CT back (top-down), leaks current state to prediction..",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".2", // .2 > .1 in std
				}},
			{Sel: ".CTBackMax", Desc: "CT to CT back (top-down), max",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".5",
				}},
			{Sel: ".SToCTMax", Desc: "higher Super to CT back (top-down), leaks current state to prediction..",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".5",
				}},
			{Sel: "#LIPToFEF", Desc: "stronger",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.PrjnScale.Rel": "1",
				}},
			{Sel: "#LIPCTToS1eP", Desc: "stronger",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
				}},
			{Sel: "#LIPToLIPCT", Desc: "stronger",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
				}},
			{Sel: "#FEFToLIP", Desc: "weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": ".5",
				}},
			{Sel: "#MDeToFEF", Desc: "weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".1", // .1 = worse!
					"Prjn.PrjnScale.Abs": "1",  // 5x5 elevates rel vs. full -- need to recompensate
				}},
			{Sel: "#S1eToFEF", Desc: "weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".5",
				}},
			{Sel: "#MDeToLIP", Desc: "weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1", // 1 > .5 > .1
					"Prjn.PrjnScale.Abs": "1", // 5x5 elevates rel vs. full -- need to recompensate
				}},
			{Sel: "#V1fToFEF", Desc: "adjust",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "0.8", // this pathway is too weak
				}},
			{Sel: "#V1fToLIP", Desc: "adjust",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "0.5", // this pathway is too weak with full prjn, to strong with 5x5s1
				}},
			{Sel: "#FEFToMDe", Desc: "sensitive",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
				}},
		},
	}},
}
