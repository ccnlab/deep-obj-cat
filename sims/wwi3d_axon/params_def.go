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
					"Layer.Act.Gbar.L":                   "0.2", // 0.2 now best
					"Layer.Act.Decay.Act":                "0.0", // 0 best
					"Layer.Act.Decay.Glong":              "0.0", // 0.5 > 0.2
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
					"Layer.CtxtGeGain":      "0.1",
					"Layer.Inhib.Layer.Gi":  "1.1",
					"Layer.Act.KNa.On":      "true",
					"Layer.Act.NMDA.Gbar":   "0.03", // larger not better
					"Layer.Act.GABAB.Gbar":  "0.2",
					"Layer.Act.Decay.Act":   "0.0",
					"Layer.Act.Decay.Glong": "0.0",
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.NoTopo":      "false", //
					"Layer.TRC.AvgMix":      "0.5",   //
					"Layer.TRC.DriveScale":  "0.05",  // LIP .1 > .05 -- too high might = too much plus phase
					"Layer.Act.NMDA.Gbar":   "0.03",  // 0.1 > .05 / .03 > .2 -- much stronger!
					"Layer.Act.GABAB.Gbar":  "0.2",   //
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
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.1",
					"Layer.Inhib.ActAvg.Init": "0.03",
					"Layer.Inhib.ActAvg.Targ": "0.03",
				}},
			{Sel: "#V1m", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.03",
					"Layer.Inhib.ActAvg.Targ": "0.03",
				}},
			{Sel: "#V1h", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.02",
					"Layer.Inhib.ActAvg.Targ": "0.02",
				}},
			{Sel: ".LIP", Desc: "high, pool inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.1",
					"Layer.Inhib.Pool.On":     "false", // false > true
					"Layer.Inhib.ActAvg.Init": "0.05",
					"Layer.Inhib.ActAvg.Targ": "0.05",
				}},
			{Sel: ".PopIn", Desc: "pop-code input",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.06",
					"Layer.Inhib.ActAvg.Targ": "0.06",
				}},
			{Sel: "#EyePos", Desc: "eyeposition input",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.02",
					"Layer.Inhib.ActAvg.Targ": "0.02",
				}},
			{Sel: ".V2", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "1.1",
					"Layer.Inhib.ActAvg.Init":    "0.06", // .06 stable without adapting, was .04
					"Layer.Inhib.ActAvg.Targ":    "0.06",
					"Layer.Inhib.ActAvg.AdaptGi": "false", // try false?
				}},
			{Sel: ".V3", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "1.1",
					"Layer.Inhib.ActAvg.Init":    "0.08", // was .06
					"Layer.Inhib.ActAvg.Targ":    "0.08",
					"Layer.Inhib.ActAvg.AdaptGi": "false",
				}},
			{Sel: ".DP", Desc: "no pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":        "false",
					"Layer.Inhib.ActAvg.Init":    "0.1", // .3 with gi1.1, DPCT .06
					"Layer.Inhib.ActAvg.Targ":    "0.1",
					"Layer.Inhib.ActAvg.AdaptGi": "false",
				}},
			{Sel: "#DP", Desc: "no pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.2", // 1.2 > 1.3 probably; actavg .3 with 1.1
				}},
			{Sel: ".V4", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.ActAvg.Init":    "0.06", // .06 once DP, TEO fixed
					"Layer.Inhib.ActAvg.Targ":    "0.06",
					"Layer.Inhib.ActAvg.AdaptGi": "false",
				}},
			{Sel: ".TEO", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.ActAvg.Init":    "0.1", // .1 for 1.2, .08 for 1.3
					"Layer.Inhib.ActAvg.Targ":    "0.1",
					"Layer.Inhib.ActAvg.AdaptGi": "false", // false works with higher gi
					"Layer.Inhib.Pool.Gi":        "1.2",
				}},
			{Sel: "#TEO", Desc: "stronger than reg",
				Params: params.Params{
					"Layer.Inhib.Pool.Gi": "1.3", // 1.2 or 1.3 > 1.1
				}},
			{Sel: ".TE", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.ActAvg.Init":    "0.1",
					"Layer.Inhib.ActAvg.Targ":    "0.1",
					"Layer.Inhib.ActAvg.AdaptGi": "false", // false ok with higher gi
					"Layer.Inhib.Pool.Gi":        "1.1",
				}},
			{Sel: "#TE", Desc: "stronger than reg",
				Params: params.Params{
					"Layer.Inhib.Pool.Gi": "1.2", // 1.2 > 1.1
				}},
			{Sel: "#LIPCT", Desc: "special",
				Params: params.Params{
					"Layer.CtxtGeGain": "0.2", //
				}},
			{Sel: "#LIPP", Desc: "layer only",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":   "1.0", //
					"Layer.Inhib.Pool.On":    "false",
					"Layer.TRC.DriveScale":   "0.15", // .15 > .1 > .05
					"Layer.TRC.FullDriveAct": "0.4",
				}},
			{Sel: "#MTPos", Desc: "layer only",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.0", // very high to get center of mass blob
					"Layer.Inhib.Pool.On":  "false",
				}},
			{Sel: "#V2P", Desc: "less AvgMix?",
				Params: params.Params{
					"Layer.TRC.AvgMix": "0.0", // no real diff vs. .5
				}},
			{Sel: "#TEOP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "false", // true def
				}},
			{Sel: "#TEP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "false", // true def
				}},

			// prjn classes, specifics
			{Sel: "Prjn", Desc: "yes extra learning factors",
				Params: params.Params{
					"Prjn.Learn.Lrate.Base": "0.02", // .02 > .04 here & lvis
					// "Prjn.SWt.Init.Sym":          "false", // experimenting with asymmetry
					"Prjn.PrjnScale.ScaleLrate": "2",   // 2 = fast response, effective
					"Prjn.PrjnScale.LoTol":      "0.8", // good now...
					"Prjn.PrjnScale.Init":       "1",
					"Prjn.PrjnScale.AvgTau":     "500",   // slower default
					"Prjn.PrjnScale.Adapt":      "false", // adapt bad maybe?  put GeMax at 1.2, adjust to avoid
					"Prjn.SWt.Adapt.On":         "true",  // true > false, esp in cosdiff
					"Prjn.SWt.Adapt.Lrate":      "0.001", //
					"Prjn.SWt.Adapt.SigGain":    "6",
					"Prjn.SWt.Adapt.DreamVar":   "0.02", // 0.02 good in lvis
					"Prjn.SWt.Init.SPct":        "1",    // 1 > lower
					"Prjn.SWt.Init.Mean":        "0.5",  // .5 > .4 -- key, except v2?
					"Prjn.SWt.Limit.Min":        "0.2",  // .2-.8 == .1-.9; .3-.7 not better
					"Prjn.SWt.Limit.Max":        "0.8",  //
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
					// "Prjn.PrjnScale.Init": "0.8", // weaker?
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
					"Prjn.PrjnScale.Init":   "0.1", // .1 = .2, slower blowup
					"Prjn.PrjnScale.Adapt":  "false",
					"Prjn.IncGain":          "1", // .5 def
				}},

			{Sel: ".V1V2", Desc: "special SWt params",
				Params: params.Params{
					"Prjn.SWt.Init.Mean":  "0.4", // .4 here is key!
					"Prjn.SWt.Limit.Min":  "0.1", // .1-.7
					"Prjn.SWt.Limit.Max":  "0.7", //
					"Prjn.PrjnScale.Init": "0.5", // .5 = 1.5 MaxGeM
					"Prjn.PrjnScale.Rel":  "2",   //
				}},
			{Sel: ".FwdWeak", Desc: "weak feedforward",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1", // .1 orig -- had a bug tho!! also trying .05
				}},

			{Sel: ".V1SC", Desc: "v1 shortcut",
				Params: params.Params{
					"Prjn.Learn.Lrate.Base": "0.001", //
					"Prjn.PrjnScale.Rel":    "0.5",   // .5 lvis
					"Prjn.SWt.Adapt.On":     "false", // seems better
				}},
			{Sel: ".FmLIP", Desc: "no random weights here",
				Params: params.Params{
					"Prjn.SWt.Init.Var": "0.05", // was 0 -- trying .05
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
					"Prjn.PrjnScale.Init": "0.8",
					"Prjn.PrjnScale.Rel":  "1.25",
				}},
			{Sel: "#TECTToTEP", Desc: "not as weak",
				Params: params.Params{
					"Prjn.PrjnScale.Init": "0.8",
					"Prjn.PrjnScale.Rel":  "1.25",
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
			{Sel: ".CTSelfLower", Desc: "CT to CT for lower-level layers: V2,3,4",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.5", // 0.1 > 0.2
				}},
			{Sel: ".CTSelfHigher", Desc: "CT to CT for higher-level layers: TEO, TE",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1", // 1.0 > 0.5
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

			{Sel: "#V2PToV2CT", Desc: "weaker pulv -> CT lower -- necessary",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".1",
				}},
			{Sel: "#V3PToV3CT", Desc: "weaker pulv -> CT lower -- necessary",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".1",
				}},
			{Sel: "#TEOCTToDPCT", Desc: "todo more consistent if .2",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": ".1",
				}},

			{Sel: "#V2ToV3", Desc: "otherwise V2 too strong",
				Params: params.Params{
					"Prjn.PrjnScale.Init": "0.5",
					"Prjn.PrjnScale.Rel":  "2",
				}},
			{Sel: "#V2ToV4", Desc: "otherwise V2 too strong",
				Params: params.Params{
					"Prjn.PrjnScale.Init": "0.5",
					"Prjn.PrjnScale.Rel":  "2",
				}},
			{Sel: "#V3ToDP", Desc: "too weak full from topo",
				Params: params.Params{
					"Prjn.PrjnScale.Init": "1",
					"Prjn.PrjnScale.Rel":  "1",
				}},
			{Sel: "#V4ToTEO", Desc: "too weak full from topo",
				Params: params.Params{
					"Prjn.PrjnScale.Init": ".5",
					"Prjn.PrjnScale.Rel":  "2",
				}},
			{Sel: "#TEOToTE", Desc: "too weak full from topo",
				Params: params.Params{
					"Prjn.PrjnScale.Init": "1",
					"Prjn.PrjnScale.Rel":  "1",
				}},

			// {Sel: "#TEToTEO", Desc: "weaker top-down than std .1",
			// 	Params: params.Params{
			// 		"Prjn.PrjnScale.Rel": "0.1", // 0.1 > 0.2
			// 	}},

			{Sel: "#MTPosToLIP", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.5",
				}},
		},
	}},
}
