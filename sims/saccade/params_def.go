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
					"Layer.Inhib.FBAct.Tau":              "30",  // 30 > 20 >> 1 definitively
					"Layer.Act.Dt.IntTau":                "40",  // 40 > 20
					"Layer.Inhib.Layer.Gi":               "1.1", // general default
					"Layer.Inhib.Pool.Gi":                "1.1", // general default
					"Layer.Inhib.ActAvg.LoTol":           "1.1", // no low adapt
					"Layer.Inhib.ActAvg.AdaptRate":       "0.2", // 0.5 default
					"Layer.Inhib.ActAvg.Init":            "0.06",
					"Layer.Inhib.ActAvg.Targ":            "0.06",
					"Layer.Act.Gbar.L":                   "0.2", // 0.2 now best
					"Layer.Act.Decay.Act":                "0.2", // 0 best
					"Layer.Act.Decay.Glong":              "0.6", // 0.5 > 0.2
					"Layer.Act.KNa.Fast.Max":             "0.1", // fm both .2 worse
					"Layer.Act.KNa.Med.Max":              "0.2", // 0.2 > 0.1 def
					"Layer.Act.KNa.Slow.Max":             "0.2", // 0.2 > higher
					"Layer.Act.Noise.Dist":               "Gaussian",
					"Layer.Act.Noise.Mean":               "0.0",     // .05 max for blowup
					"Layer.Act.Noise.Var":                "0.01",    // .01 a bit worse
					"Layer.Act.Noise.Type":               "NoNoise", // off for now
					"Layer.Act.GTarg.GeMax":              "1.2",     // 1.2 > 1 > .8 -- rescaling not very useful.
					"Layer.Act.Dt.LongAvgTau":            "20",      // 20 > 50 > 100
					"Layer.Learn.TrgAvgAct.ErrLrate":     "0.01",    // 0.01 orig > 0.005
					"Layer.Learn.TrgAvgAct.SynScaleRate": "0.005",   // 0.005 orig > 0.01
					"Layer.Learn.TrgAvgAct.TrgRange.Min": "0.5",     // .5 > .2 overall
					"Layer.Learn.TrgAvgAct.TrgRange.Max": "2.0",     // objrec 2 > 1.8
				}},
			{Sel: ".CT", Desc: "CT gain factor is key",
				Params: params.Params{
					"Layer.CtxtGeGain":      "0.2", // .2 > .15 > .1 > .05
					"Layer.Inhib.Layer.Gi":  "1.1",
					"Layer.Act.KNa.On":      "true",
					"Layer.Act.NMDA.Gbar":   "0.03",
					"Layer.Act.GABAB.Gbar":  "0.2",
					"Layer.Act.Decay.Act":   "0.0", // 0 better
					"Layer.Act.Decay.Glong": "0.0",
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.NoTopo":      "false", //
					"Layer.TRC.AvgMix":      "0.5",   //
					"Layer.TRC.DriveScale":  "0.2",   // .2 > .15 > .1, .05
					"Layer.Act.NMDA.Gbar":   "0.03",
					"Layer.Act.GABAB.Gbar":  "0.2", //
					"Layer.Act.Decay.Act":   "0.5",
					"Layer.Act.Decay.Glong": "1", // clear long
				}},
			{Sel: "SuperLayer", Desc: "burst params don't really matter",
				Params: params.Params{
					"Layer.Burst.ThrRel": "0.1", // not big diffs
					"Layer.Burst.ThrAbs": "0.1",
				}},
			{Sel: ".V1", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.1",
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.ActAvg.Init": "0.02", // .02 for .1 sigma, .04 for .15
					"Layer.Inhib.ActAvg.Targ": "0.02",
				}},
			{Sel: ".PopIn", Desc: "pop-code input",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.02",
					"Layer.Inhib.ActAvg.Targ": "0.02",
				}},
			{Sel: "#MDe", Desc: "",
				Params: params.Params{
					"Layer.Act.Clamp.Ge":      "0.6", // 0.6 > 0.8?
					"Layer.Inhib.ActAvg.Init": "0.02",
					"Layer.Inhib.ActAvg.Targ": "0.02",
					"Layer.Inhib.Layer.Gi":    "1.3", // 1.3 > 1.2 > 1.1
				}},
			{Sel: ".LIP", Desc: "high, pool inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":               "10",
					"Layer.Inhib.Pool.Gi":                "0.8",
					"Layer.Inhib.Pool.On":                "true",
					"Layer.Inhib.ActAvg.Init":            "0.02",
					"Layer.Inhib.ActAvg.Targ":            "0.02",
					"Layer.Learn.TrgAvgAct.TrgRange.Min": "0.5",
					"Layer.Learn.TrgAvgAct.TrgRange.Max": "2.0",   // reducing does not help anything
					"Layer.Learn.TrgAvgAct.Pool":         "false", // pool sizes too small for trgs!
				}},
			{Sel: "#LIPCT", Desc: "strong inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "5",
					"Layer.Inhib.Pool.On":  "true",
				}},
			{Sel: "#LIPP", Desc: "strong inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "5",
					"Layer.Inhib.Pool.On":  "false",
				}},
			{Sel: "#LIPPS", Desc: "strong inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "5",
					"Layer.Inhib.Pool.On":  "false",
				}},
			{Sel: "#FEF", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":       "1.1",
					"Layer.Inhib.Pool.Gi":        "0.9",
					"Layer.Inhib.Pool.On":        "false", // full layer best
					"Layer.Inhib.ActAvg.Init":    "0.09",
					"Layer.Inhib.ActAvg.Targ":    "0.09",
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
					"Prjn.Learn.Lrate.Base": "0.04",
					// "Prjn.SWt.Init.Sym":          "false", // experimenting with asymmetry
					"Prjn.PrjnScale.ScaleLrate": "2",     // 2 = fast response, effective
					"Prjn.PrjnScale.LoTol":      "0.8",   // good now...
					"Prjn.PrjnScale.AvgTau":     "500",   // slower default
					"Prjn.PrjnScale.Adapt":      "false", // adapt bad maybe?  put GeMax at 1.2, adjust to avoid
					"Prjn.SWt.Adapt.On":         "true",  // true > false, esp in cosdiff
					"Prjn.SWt.Adapt.Lrate":      "0.01",  //
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
					"Prjn.SWt.Init.Var":  "0.25",
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
					"Prjn.PrjnScale.Abs": ".5",
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
					"Prjn.PrjnScale.Abs": "0.2",
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
					"Prjn.SWt.Init.Mean": "0.8", // 0.8 > 0.5 with lower S -> CT rel (2 instead of 4)
					"Prjn.PrjnScale.Rel": "1",   // def 2
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
					"Prjn.PrjnScale.Abs": "1.2",
				}},
			{Sel: "#FEFToLIP", Desc: "weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": ".5",
				}},
			{Sel: "#MDeToFEF", Desc: "weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".1",
					"Prjn.PrjnScale.Abs": "1", // 5x5 elevates rel vs. full -- need to recompensate
				}},
			{Sel: "#V1fToFEF", Desc: "stronger",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "3", // this pathway is too weak
				}},
			{Sel: "#V1fToLIP", Desc: "stronger",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1", // this pathway is too weak with full prjn, fine with 5x5s1
				}},
			{Sel: "#FEFToMDe", Desc: "sensitive",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
				}},
		},
	}},
}
