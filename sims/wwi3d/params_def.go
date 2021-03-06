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
					"Layer.Learn.AvgL.Gain":       "3.0",  // key param -- 3 > 2.5 > 3.5 except IT!
					"Layer.Act.Gbar.L":            "0.1",  // todo: orig has 0.2 -- don't see any exploration notes..
					"Layer.Inhib.Layer.FBTau":     "1.4",  // smoother = faster? but worse?
					"Layer.Inhib.Pool.FBTau":      "1.4",  // smoother = faster?
					"Layer.Act.Init.Decay":        "0",    // used deep default, now must set
					"Layer.Inhib.ActAvg.UseFirst": "true", // doesn't fix weird V3 effect, works better overall
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.AvgMix": "0.5", // actually best on
				}},
			{Sel: "SuperLayer", Desc: "burst params don't really matter",
				Params: params.Params{
					"Layer.Burst.ThrRel": ".1", // not big diffs
					"Layer.Burst.ThrAbs": ".1",
				}},
			{Sel: ".V1", Desc: "pool inhib (not used), initial activity",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "3.0",
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.03",
				}},
			{Sel: ".LIP", Desc: "high, pool inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "2.4",
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.5",
					"Layer.Inhib.ActAvg.Init": "0.1",
				}},
			{Sel: ".PopIn", Desc: "pop-code input",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.1",
				}},
			{Sel: "#EyePos", Desc: "eyeposition input",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init": "0.025",
				}},
			{Sel: ".V2", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.04",
				}},
			{Sel: ".V3", Desc: "pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.1",
				}},
			{Sel: ".V4", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.1",
					"Layer.Learn.AvgL.Gain":   "2.5", // key param -- 3 > 2.5 > 3.5 except V4/IT!
				}},
			{Sel: ".DP", Desc: "no pool inhib, initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.ActAvg.Init": "0.15",
				}},
			{Sel: ".TEO", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Layer.On":    "true",
					"Layer.Inhib.Layer.Gi":    "1.8",
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.1",
					"Layer.Learn.AvgL.Gain":   "2.5", // key param -- 3 > 2.5 > 3.5 except V4/IT!
				}},
			{Sel: ".TE", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Layer.On":    "true",
					"Layer.Inhib.Layer.Gi":    "1.8", // no benefits to reducing
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.15",
					"Layer.Learn.AvgL.Gain":   "2.5", // key param -- 3 > 2.5 > 3.5 except V4/IT!
				}},
			{Sel: "#LIPCT", Desc: "higher inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.6",
				}},
			{Sel: "#LIPP", Desc: "layer only",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.8",
					"Layer.Inhib.Pool.On":  "false",
				}},
			{Sel: "#MTPos", Desc: "layer only",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.8",
					"Layer.Inhib.Pool.On":  "false",
				}},
			{Sel: "#V2P", Desc: "less AvgMix?",
				Params: params.Params{
					"Layer.TRC.AvgMix": "0.0", // no real diff vs. .5
				}},
			{Sel: "#TEOP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "true", // true def
				}},
			{Sel: "#TEP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "true", // true def
				}},

			// prjn classes, specifics
			{Sel: "Prjn", Desc: "yes extra learning factors",
				Params: params.Params{
					"Prjn.Learn.Norm.On":       "true",
					"Prjn.Learn.Momentum.On":   "true",
					"Prjn.Learn.Momentum.MTau": "10",   // now 10 much better than 20!
					"Prjn.Learn.WtBal.On":      "true", // essential
					"Prjn.Learn.Lrate":         "0.04", // must set initial lrate here when using schedule!
					// "Prjn.WtInit.Sym":          "false", // experimenting with asymmetry
				}},
			{Sel: "CTCtxtPrjn", Desc: "defaults for CT Ctxt prjns",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1",
				}},
			{Sel: ".Fixed", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.8",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "false",
				}},
			// {Sel: ".Forward", Desc: "std feedforward",
			// 	Params: params.Params{
			// 		"Prjn.Learn.WtSig.PFail":      "0.5",
			// 		"Prjn.Learn.WtSig.PFailWtMax": "0.8",
			// 	}},
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates -- smaller as network gets bigger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
					// "Prjn.Learn.WtSig.PFail":      "0.5",
					// "Prjn.Learn.WtSig.PFailWtMax": "0.8",
				}},

			{Sel: ".FwdWeak", Desc: "weak feedforward",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1", // .1 orig -- had a bug tho!! also trying .05
				}},

			{Sel: ".FmLIP", Desc: "no random weights here",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0.05", // was 0 -- trying .05
					"Prjn.WtInit.Sym":  "false",
				}},
			{Sel: ".BackStrong", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2", // .1 > orig .2 > .05 -- not sep fm BackMax -- .1 = better TE_V1Sim, V2P cosdiff
				}},
			{Sel: ".BackMax", Desc: "strongest",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5", // .1 > .2, orig .5 -- see BackStrong
				}},

			{Sel: ".BackToPulv", Desc: "top-down to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".FwdToPulv", Desc: "feedforward to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},

			{Sel: ".FmPulv", Desc: "default for pulvinar",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2", // .2 > .1 > .05 still true
				}},
			{Sel: ".Lateral", Desc: "default for lateral",
				Params: params.Params{
					"Prjn.WtInit.Sym":  "false",
					"Prjn.WtScale.Rel": "0.02", // .02 > .05 == .01 > .1  -- very minor diffs on TE cat
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0",
				}},

			{Sel: ".CTFmSuper", Desc: "CT from main super -- fixed one2one",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.8", // 0.8 > 0.5 with lower S -> CT rel (2 instead of 4)
					"Prjn.WtScale.Rel": "2",
				}},
			{Sel: ".CTFmSuperLower", Desc: "CT from main super -- for lower layers",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.8", // 0.8 makes a diff for lower too, more V1 divergence at .5
					"Prjn.WtScale.Rel": "1",   // 1 maybe better
				}},
			{Sel: ".CTSelfLower", Desc: "CT to CT for lower-level layers: V2,3,4",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5", // 0.1 > 0.2
				}},
			{Sel: ".CTSelfHigher", Desc: "CT to CT for higher-level layers: TEO, TE",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1", // 1.0 > 0.5
				}},
			{Sel: ".CTBack", Desc: "CT to CT back (top-down)",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".2", // .2 > .1 in std
				}},
			{Sel: ".SToCT", Desc: "higher Super to CT back (top-down), leaks current state to prediction..",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".2", // .2 > .1 in std
				}},
			{Sel: ".CTBackMax", Desc: "CT to CT back (top-down), max",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".5",
				}},
			{Sel: ".SToCTMax", Desc: "higher Super to CT back (top-down), leaks current state to prediction..",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".5",
				}},

			{Sel: "#V2PToV2CT", Desc: "weaker pulv -> CT lower -- necessary",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".1",
				}},
			{Sel: "#V3PToV3CT", Desc: "weaker pulv -> CT lower -- necessary",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".1",
				}},
			{Sel: "#TEOCTToDPCT", Desc: "todo more consistent if .2",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".1",
				}},

			{Sel: "#V2ToV3", Desc: "otherwise V2 too strong",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.5",
					"Prjn.WtScale.Rel": "2",
				}},
			{Sel: "#V2ToV4", Desc: "otherwise V2 too strong",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.5",
					"Prjn.WtScale.Rel": "2",
				}},
			{Sel: "#V3ToDP", Desc: "too weak full from topo",
				Params: params.Params{
					"Prjn.WtScale.Abs": "2",
					"Prjn.WtScale.Rel": "0.5",
				}},
			{Sel: "#V4ToTEO", Desc: "too weak full from topo",
				Params: params.Params{
					"Prjn.WtScale.Abs": "2",
					"Prjn.WtScale.Rel": "0.5",
				}},
			{Sel: "#TEOToTE", Desc: "too weak full from topo",
				Params: params.Params{
					"Prjn.WtScale.Abs": "1.5",
					"Prjn.WtScale.Rel": "0.667",
				}},

			// {Sel: "#TEToTEO", Desc: "weaker top-down than std .1",
			// 	Params: params.Params{
			// 		"Prjn.WtScale.Rel": "0.1", // 0.1 > 0.2
			// 	}},

			{Sel: "#MTPosToLIP", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5",
				}},
		},
	}},
}
