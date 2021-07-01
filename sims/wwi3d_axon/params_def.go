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
					"Layer.Act.Gbar.L":                   "0.2", // 0.2 now best
					"Layer.Act.Decay.Act":                "0.0",
					"Layer.Act.Decay.Glong":              "0.0",
					"Layer.Act.KNa.Fast.Max":             "0.1", // fm both .2 worse
					"Layer.Act.KNa.Med.Max":              "0.2", // 0.2 > 0.1 def
					"Layer.Act.KNa.Slow.Max":             "0.2", // 0.2 > higher
					"Layer.Act.Noise.Dist":               "Gaussian",
					"Layer.Act.Noise.Mean":               "0.0",     // .05 max for blowup
					"Layer.Act.Noise.Var":                "0.01",    // .01 a bit worse
					"Layer.Act.Noise.Type":               "NoNoise", // off for now
					"Layer.Act.GTarg.GeMax":              "1.0",     // 1 > .8 -- rescaling not very useful.
					"Layer.Act.Dt.LongAvgTau":            "20",      // 20 > 50 > 100
					"Layer.Learn.TrgAvgAct.ErrLrate":     "0.01",    // 0.01 orig > 0.005
					"Layer.Learn.TrgAvgAct.SynScaleRate": "0.005",   // 0.005 orig > 0.01
					"Layer.Learn.TrgAvgAct.TrgRange.Min": "0.5",     // .5 > .2 overall
					"Layer.Learn.TrgAvgAct.TrgRange.Max": "2.0",     // objrec 2 > 1.8
				}},
			{Sel: ".CT", Desc: "CT gain factor is key",
				Params: params.Params{
					"Layer.CtxtGeGain":      "0.2", // 0.2 > 0.3 > 0.1
					"Layer.Inhib.Layer.Gi":  "1.1",
					"Layer.Act.KNa.On":      "true",
					"Layer.Act.NMDA.Gbar":   "0.03", // larger not better
					"Layer.Act.GABAB.Gbar":  "0.2",
					"Layer.Act.Decay.Act":   "0.0",
					"Layer.Act.Decay.Glong": "0.0",
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.NoTopo":      "false", // actually best on
					"Layer.TRC.AvgMix":      "0.5",   // actually best on
					"Layer.TRC.DriveScale":  "0.05",  // 0.05 > 0.1 > 0.15
					"Layer.Act.GABAB.Gbar":  "0.005", //
					"Layer.Act.NMDA.Gbar":   "0.1",   // 0.1 > .05 / .03 > .2 -- much stronger!
					"Layer.Act.Decay.Act":   "0.5",
					"Layer.Act.Decay.Glong": "1", // clear long
				}},
			{Sel: "SuperLayer", Desc: "burst params don't really matter",
				Params: params.Params{
					"Layer.Burst.ThrRel": ".1", // not big diffs
					"Layer.Burst.ThrAbs": ".1",
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
					"Layer.Inhib.ActAvg.Init": "0.035",
					"Layer.Inhib.ActAvg.Targ": "0.035",
				}},
			{Sel: "#V1h", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.02",
					"Layer.Inhib.ActAvg.Targ": "0.02",
				}},
			{Sel: ".LIP", Desc: "high, pool inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "1.1", // 1.2 > 1.1
					"Layer.Inhib.Pool.On":     "false",
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
					"Layer.Inhib.Pool.Gi":        "1.0",
					"Layer.Inhib.ActAvg.Init":    "0.04",
					"Layer.Inhib.ActAvg.Targ":    "0.04",
					"Layer.Inhib.ActAvg.AdaptGi": "true",
					"Layer.Act.GTarg.GeMax":      "1.2", // these need to get stronger?
				}},
			{Sel: ".V3", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.0",
					"Layer.Inhib.ActAvg.Init": "0.1",
					"Layer.Inhib.ActAvg.Targ": "0.1",
				}},
			{Sel: ".V4", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.ActAvg.Init":    "0.04",
					"Layer.Inhib.ActAvg.Targ":    "0.04",
					"Layer.Inhib.ActAvg.AdaptGi": "true", // adapt > not still better v34
				}},
			{Sel: ".DP", Desc: "no pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":        "false",
					"Layer.Inhib.ActAvg.Init":    "0.06",
					"Layer.Inhib.ActAvg.Targ":    "0.06",
					"Layer.Inhib.ActAvg.AdaptGi": "true", // adapt > not still better v34
				}},
			{Sel: ".TEO", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.ActAvg.Init":    "0.06",
					"Layer.Inhib.ActAvg.Targ":    "0.06",
					"Layer.Inhib.ActAvg.AdaptGi": "true", // adapt > not still better v34
				}},
			{Sel: ".TE", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.ActAvg.Init":    "0.06",
					"Layer.Inhib.ActAvg.Targ":    "0.06",
					"Layer.Inhib.ActAvg.AdaptGi": "true", // adapt > not still better v34
				}},
			// {Sel: "#LIPCT", Desc: "higher inhib",
			// 	Params: params.Params{
			// 		"Layer.Inhib.Layer.Gi": "1.2",
			// 	}},
			{Sel: "#LIPP", Desc: "layer only",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "0.9", // 0.9 > 0.8 > 1.0
					"Layer.Inhib.Pool.On":  "false",
					"Layer.TRC.DriveScale": "0.05", // 0.05 > 0.1 > 0.15
				}},
			{Sel: "#MTPos", Desc: "layer only",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.0",
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
					"Prjn.Learn.Lrate.Base": "0.04", // must set initial lrate here when using schedule!
					// "Prjn.SWt.Init.Sym":          "false", // experimenting with asymmetry
					"Prjn.PrjnScale.ScaleLrate": "2",   // 2 = fast response, effective
					"Prjn.PrjnScale.LoTol":      "0.8", // good now...
					"Prjn.PrjnScale.Init":       "1",
					"Prjn.PrjnScale.AvgTau":     "500",  // slower default
					"Prjn.SWt.Adapt.On":         "true", // true > false, esp in cosdiff
					"Prjn.SWt.Adapt.Lrate":      "1",    // .01 == .001 == .1 459
					"Prjn.SWt.Adapt.SigGain":    "6",
					"Prjn.SWt.Adapt.DreamVar":   "0.0", // < 0.005 no effect 0.01 max that works in small nets
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
					"Prjn.Learn.Learn":   "false",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0",
					"Prjn.SWt.Init.Sym":  "false",
				}},
			// {Sel: ".Forward", Desc: "std feedforward",
			// 	Params: params.Params{
			// 		"Prjn.Learn.WtSig.PFail":      "0.5",
			// 		"Prjn.Learn.WtSig.PFailWtMax": "0.8",
			// 	}},
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates -- smaller as network gets bigger",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.2",
				}},

			{Sel: ".FwdWeak", Desc: "weak feedforward",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1", // .1 orig -- had a bug tho!! also trying .05
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
					"Prjn.PrjnScale.Rel": "0.5", // .1 > .2, orig .5 -- see BackStrong
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
					"Prjn.PrjnScale.Rel": "2",   // def 2
				}},
			{Sel: ".CTFmSuperLower", Desc: "CT from main super -- for lower layers",
				Params: params.Params{
					"Prjn.SWt.Init.Mean": "0.8", // 0.8 makes a diff for lower too, more V1 divergence at .5
					"Prjn.PrjnScale.Rel": "1",   // 1 maybe better
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
					// "Prjn.PrjnScale.Abs": "0.5",
					"Prjn.PrjnScale.Rel": "2",
				}},
			{Sel: "#V2ToV4", Desc: "otherwise V2 too strong",
				Params: params.Params{
					// "Prjn.PrjnScale.Abs": "0.5",
					"Prjn.PrjnScale.Rel": "2",
				}},
			{Sel: "#V3ToDP", Desc: "too weak full from topo",
				Params: params.Params{
					// "Prjn.PrjnScale.Abs": "2",
					"Prjn.PrjnScale.Rel": "0.5",
				}},
			{Sel: "#V4ToTEO", Desc: "too weak full from topo",
				Params: params.Params{
					// "Prjn.PrjnScale.Abs": "2",
					"Prjn.PrjnScale.Rel": "0.5",
				}},
			{Sel: "#TEOToTE", Desc: "too weak full from topo",
				Params: params.Params{
					// "Prjn.PrjnScale.Abs": "1.5",
					"Prjn.PrjnScale.Rel": "0.667",
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
