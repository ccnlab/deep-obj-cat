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
					"Layer.Learn.AvgL.Gain":       "3.0",   // key param -- 3 > 2.5 > 3.5 except IT!
					"Layer.Act.Gbar.L":            "0.2",   // 0.2
					"Layer.Inhib.Layer.FBTau":     "1.4",   // smoother = faster? but worse?
					"Layer.Inhib.Pool.FBTau":      "1.4",   // smoother = faster?
					"Layer.Inhib.ActAvg.UseFirst": "false", // true is default
					"Layer.Act.Init.Decay":        "0",     // used deep default, now must set
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.AvgMix": "0.5", // actually best on
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
					"Layer.Inhib.ActAvg.Init": "0.15",
				}},
			{Sel: ".V4", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
					"Layer.Inhib.Layer.On":    "true",
					"Layer.Inhib.Layer.Gi":    "1.8",
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.15",
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
					"Layer.Inhib.ActAvg.Init": "0.15",
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
					"Layer.TRC.NoTopo": "false", // new = true", // now best
				}},
			{Sel: "#TEP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "false", // new = true", // now best
				}},

			// prjn classes, specifics
			{Sel: "Prjn", Desc: "yes extra learning factors",
				Params: params.Params{
					"Prjn.Learn.Norm.On":       "true",
					"Prjn.Learn.Momentum.On":   "true",
					"Prjn.Learn.Momentum.MTau": "20",   // 20 orig -- now 10 much better than 20!
					"Prjn.Learn.WtBal.On":      "true", // essential
					"Prjn.Learn.Lrate":         "0.04", // must set initial lrate here when using schedule!
				}},
			{Sel: "CTCtxtPrjn", Desc: "defaults for CT Ctxt prjns",
				Params: params.Params{
					"Prjn.WtScale.Rel":       "1",
					"Prjn.Learn.Norm.On":     "false", // critical to be off!
					"Prjn.Learn.Momentum.On": "false",
				}},
			{Sel: ".Fixed", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.8",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "false",
				}},
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates -- smaller as network gets bigger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},

			{Sel: ".StdFF", Desc: "standard feedforward",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1.0",
				}},
			{Sel: ".StdFB", Desc: "standard feedback",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".FwdAbs5Rel2", Desc: "reduced abs, compensatory 2x rel -- too strongly activated otherwise",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.5",
					"Prjn.WtScale.Rel": "2.0",
				}},
			{Sel: ".FwdWeak", Desc: "weak feedforward",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1", // .1 orig
				}},

			{Sel: ".FmLIP", Desc: "In new model: no random weights here",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0.05",  // new has 0
					"Prjn.WtInit.Sym":  "false", // some have false -- go with this
				}},
			{Sel: ".BackMed", Desc: "medium / default",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1", // orig .1 -- todo try .05 in new
				}},
			{Sel: ".BackStrong", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2", // orig .2 -- todo try .1
				}},
			{Sel: ".BackMax", Desc: "strongest",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5", // orig .5 -- reduced in new to .1 -- todo
				}},
			{Sel: ".BackWeak05", Desc: "weak .05",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},
			{Sel: ".BackWeak02", Desc: "weak .02",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.02",
				}},
			{Sel: ".BackLIPCT", Desc: "strength = 1",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1", // 1 orig; todo try .1 == .05 > .2 > .5 in V2ct hogging, no diff else
				}},

			{Sel: ".BackToPulv", Desc: "top-down to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".BackToPulv2", Desc: "top-down to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2",
				}},
			{Sel: ".BackToPulv5", Desc: "top-down to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5",
				}},
			{Sel: ".BackToPulv1", Desc: "top-down to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1.0",
				}},
			{Sel: ".FwdToPulv", Desc: "feedforward to pulvinar directly",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},

			{Sel: ".FmPulv", Desc: "default for pulvinar",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".FmPulv2", Desc: "strong",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2",
				}},
			{Sel: ".FmPulv05", Desc: "weak",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},
			{Sel: ".FmPulv02", Desc: "weak",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.02",
				}},
			{Sel: "#V2PToV2CT", Desc: "trying pulvinar prjns better",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".1", // .1 > .2 > .05 for cosdiff, not hog (.05 bad)
				}},
			{Sel: "#V3PToV3CT", Desc: "weaker pulvinar prjns better",
				Params: params.Params{
					"Prjn.WtScale.Rel": ".1", // .05 > .1 for hog but worse for cosdif; .1 > .2 for hog, minor for cosdiff
				}},

			{Sel: ".Lateral", Desc: "default for lateral",
				Params: params.Params{
					"Prjn.WtInit.Sym":  "false",
					"Prjn.WtScale.Rel": "0.02", // .02 > .05 > .1  -- similar, but .05 has less TEO hog
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0",
				}},

			{Sel: ".ToCT1to1", Desc: "1to1 has no weight var... fixed?",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0",
				}},
			{Sel: "#LIPToLIPCT", Desc: "default 1",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1",
				}},
			{Sel: "#V2ToV2CT", Desc: "standard",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5", // .5 orig
				}},
			{Sel: "#V2CTToV2CT", Desc: "standard",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1",
				}},
			{Sel: "#V3ToV3CT", Desc: "V3 default",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1",
				}},
			{Sel: "#DPToDPCT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "3", // 3 orig
				}},
			{Sel: "#V2ToV4", Desc: "abs rel flip",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.5",
					"Prjn.WtScale.Rel": "2",
				}},
			{Sel: "#V4ToV4CT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4", // 4 orig
				}},
			{Sel: "#V4CTToV4CT", Desc: "lesioned in orig",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1", // 1 = less TEO hogging; 4 orig
				}},
			{Sel: "#TEOToTEOCT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4", // 4 orig
				}},
			{Sel: "#TEOCTToTEOCT", Desc: "reg but beneficial",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4", // 4 orig
				}},
			{Sel: "#TEToTECT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4", // 4 orig
				}},
			{Sel: "#TECTToTECT", Desc: "reg but beneficial",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4", // 4 orig
				}},
			{Sel: "#TEToTEO", Desc: "weaker",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},
			{Sel: "#MTPosToLIP", Desc: "lower variance",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5",
					"Prjn.WtInit.Var":  "0.05",
				}},
		},
	}},
}
