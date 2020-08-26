// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "github.com/emer/emergent/params"

// todo:
// * pool instead of one2one v4, IT CT prjns -- seems better
//    * v4 = better w/pool. maybe TE better w/out pool?  TEO with pool?
// * weaker ctxt
// * topo within layers

// ParamSets is the default set of parameters -- Base is always applied, and others can be optionally
// selected to apply on top of that
var ParamSets = params.Sets{
	{Name: "Base", Desc: "these are the best params", Sheets: params.Sheets{
		"Network": &params.Sheet{
			// layer classes, specifics
			{Sel: "Layer", Desc: "needs some special inhibition and learning params",
				Params: params.Params{
					"Layer.Learn.AvgL.Gain":   "3.0", // key param -- 3 > 2.5 > 3.5 except IT!
					"Layer.Act.Gbar.L":        "0.1", // todo: orig has 0.2 -- don't see any exploration notes..
					"Layer.Inhib.Layer.FBTau": "1.4", // smoother = faster? but worse?
					"Layer.Inhib.Pool.FBTau":  "1.4", // smoother = faster?
				}},
			{Sel: "TRCLayer", Desc: "avg mix param",
				Params: params.Params{
					"Layer.TRC.AvgMix": "0.0", // best off for non-TE layers
				}},
			{Sel: ".V1", Desc: "pool inhib (not used), initial activity",
				Params: params.Params{
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "3",
					"Layer.Inhib.ActAvg.Init": "0.03",
				}},
			{Sel: ".LIP", Desc: "high, pool inhib",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":    "2.4",
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.5",
					"Layer.Inhib.ActAvg.Init": "0.1",
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
					"Layer.Inhib.Pool.On":     "true",
					"Layer.Inhib.Pool.Gi":     "1.8",
					"Layer.Inhib.ActAvg.Init": "0.15",
					"Layer.Learn.AvgL.Gain":   "2.5", // key param -- 3 > 2.5 > 3.5 except V4/IT!
				}},
			{Sel: ".TE", Desc: "pool inhib, initial activity, less avgl.gain",
				Params: params.Params{
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
			// {Sel: "#V3P", Desc: "more AvgMix?",
			// 	Params: params.Params{
			// 		"Layer.TRC.AvgMix": "0.1",
			// 	}},
			// {Sel: "#V4P", Desc: "more AvgMix?",
			// 	Params: params.Params{
			// 		"Layer.TRC.AvgMix": "0.1",
			// 	}},
			{Sel: "#TEOP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "false", // false is definitely better!
					"Layer.TRC.AvgMix": "0.0",   // fine
				}},
			{Sel: "#TEP", Desc: "no topo",
				Params: params.Params{
					"Layer.TRC.NoTopo": "true", // definitely needs true here..
					"Layer.TRC.AvgMix": "0.2",
				}},

			// prjn classes, specifics
			{Sel: "Prjn", Desc: "yes extra learning factors",
				Params: params.Params{
					"Prjn.Learn.Norm.On":       "true",
					"Prjn.Learn.Momentum.On":   "true",
					"Prjn.Learn.Momentum.MTau": "20",   // has repeatedly been beneficial
					"Prjn.Learn.WtBal.On":      "true", // essential
					"Prjn.Learn.Lrate":         "0.04", // must set initial lrate here when using schedule!
				}},
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates -- smaller as network gets bigger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".Fixed", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.8",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "true",
				}},

			{Sel: ".StdFF", Desc: "standard feedforward",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1.0",
				}},
			{Sel: ".FwdWeak", Desc: "weak feedforward",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},

			{Sel: ".StdFB", Desc: "standard feedback",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".BackMed", Desc: "medium / default",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".BackStrong", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2",
				}},
			{Sel: ".BackMax", Desc: "strongest",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5",
				}},
			{Sel: ".BackWeak05", Desc: "weak .05",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},
			{Sel: ".BackLIPCT", Desc: "strength = 1",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1.0",
				}},

			{Sel: ".FmPulvMed", Desc: "medium",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.1",
				}},
			{Sel: ".FmPulvStrong", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.2",
				}},
			{Sel: ".FmPulvWeak05", Desc: "weaker",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},
			{Sel: ".FmPulvWeak02", Desc: "weaker",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.02",
				}},
			{Sel: ".FmPulvWeak01", Desc: "weaker",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.01",
				}},

			{Sel: "#LIPToLIPCT", Desc: "default 1",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1",
				}},
			{Sel: "#V2ToV2CT", Desc: "V2 has weaker -- todo: untested!",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5",
				}},
			{Sel: "#V3ToV3CT", Desc: "V3 default",
				Params: params.Params{
					"Prjn.WtScale.Rel": "1",
				}},
			{Sel: "#DPToDPCT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "3",
				}},
			{Sel: "#V4ToV4CT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4",
				}},
			{Sel: "#TEOToTEOCT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4",
				}},
			{Sel: "#TEOCTToTEOCT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4",
				}},
			{Sel: "#TEToTECT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4",
				}},
			{Sel: "#TECTToTECT", Desc: "stronger",
				Params: params.Params{
					"Prjn.WtScale.Rel": "4",
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

			{Sel: "#TEToTEO", Desc: "weaker top-down than std .1",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.05",
				}},

			{Sel: "#MTPosToLIP", Desc: "fixed weights",
				Params: params.Params{
					"Prjn.WtScale.Rel": "0.5",
				}},
		},
	}},
}
