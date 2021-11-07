Standard cemer params:

```
Layer: V1s

Layer: V1p
	Deep_Raw_Fm_V1s		Abs:	1	Rel:	1
	Fm_V2d		Abs:	1	Rel:	1
	Fm_V3d		Abs:	1	Rel:	0.2
	Fm_V4d		Abs:	1	Rel:	0.2
	Fm_TEOd		Abs:	1	Rel:	0.1

Layer: V1hs

Layer: V1hp
	Deep_Raw_Fm_V1hs		Abs:	1	Rel:	1
	Fm_V2d		Abs:	1	Rel:	1
	Fm_V3d		Abs:	1	Rel:	0.2
	Fm_V4d		Abs:	1	Rel:	0.2
	Fm_TEOd		Abs:	1	Rel:	0.1

Layer: V2s
	Fm_V1hs		Abs:	1	Rel:	1
	Fm_V1s		Abs:	1	Rel:	1
	Fm_LIPs		Abs:	1	Rel:	0.5
	Fm_V3s		Abs:	1	Rel:	0.5
	Fm_V4s		Abs:	1	Rel:	0.1
	Fm_TEOd		Abs:	1	Rel:	0.1
	Fm_V1hp		Abs:	1	Rel:	0.02
	Fm_V1p		Abs:	1	Rel:	0.02

Layer: V2d
	Ctxt_Fm_V2s		Abs:	1	Rel:	0.5
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2
	Fm_LIPd		Abs:	1	Rel:	1
	Fm_LIPp		Abs:	1	Rel:	0.2
	Fm_V3d		Abs:	1	Rel:	0.5
	Fm_V4d		Abs:	1	Rel:	0.5
	Fm_V3s		Abs:	1	Rel:	0.5
	Fm_TEOs		Abs:	1	Rel:	0.5

Layer: MtPos
	Fm_V1s		Abs:	1	Rel:	1

Layer: LIPs
	Fm_MtPos		Abs:	1	Rel:	0.5
	Fm_ObjVel		Abs:	1	Rel:	1
	Fm_SaccadePlan		Abs:	1	Rel:	1
	Fm_EyePos		Abs:	1	Rel:	1
	Fm_LIPp		Abs:	1	Rel:	0.2
	Fm_V2s		Abs:	1	Rel:	0.1
	Fm_V3s		Abs:	1	Rel:	0.1

Layer: LIPd
	Ctxt_Fm_LIPs		Abs:	1	Rel:	1
	Fm_LIPp		Abs:	1	Rel:	0.2
	Fm_ObjVel		Abs:	1	Rel:	1
	Fm_Saccade		Abs:	1	Rel:	1
	Fm_EyePos		Abs:	1	Rel:	1
	Fm_V2d		Abs:	1	Rel:	0.1
	Fm_V3d		Abs:	1	Rel:	0.1

Layer: LIPp
	Deep_Raw_Fm_MtPos		Abs:	1	Rel:	1
	Fm_LIPd		Abs:	1	Rel:	1

Layer: EyePos

Layer: SaccadePlan

Layer: Saccade

Layer: ObjVel

Layer: V3s
	Fm_V2s		Abs:	0.5	Rel:	2
	Fm_V4s		Abs:	1	Rel:	0.2
	Fm_TEOs		Abs:	1	Rel:	0.1
	Fm_DPs		Abs:	1	Rel:	0.2
	Fm_LIPs		Abs:	1	Rel:	0.1
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2
	Fm_DPp		Abs:	1	Rel:	0.05
	Fm_TEOd		Abs:	1	Rel:	0.1

Layer: V3d
	Ctxt_Fm_V3s		Abs:	1	Rel:	1
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2
	Fm_DPp		Abs:	1	Rel:	0.2
	Fm_LIPd		Abs:	1	Rel:	0.2
	Fm_DPd		Abs:	1	Rel:	0.2
	Fm_V4d		Abs:	1	Rel:	0.2
	Fm_V4s		Abs:	1	Rel:	0.2
	Fm_DPs		Abs:	1	Rel:	0.2
	Fm_TEOs		Abs:	1	Rel:	0.5

Layer: V3p
	Deep_Raw_Fm_V3s		Abs:	1	Rel:	1
	Fm_V2d		Abs:	1	Rel:	0.5
	Fm_DPd		Abs:	1	Rel:	0.2
	Fm_TEOd		Abs:	1	Rel:	0.1

Layer: DPs
	Fm_V2s		Abs:	1	Rel:	1
	Fm_V3s		Abs:	1	Rel:	1
	Fm_TEOs		Abs:	1	Rel:	0.1
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2
	Fm_V3p		Abs:	1	Rel:	0.1
	Fm_TEOp		Abs:	1	Rel:	0.1

Layer: DPd
	Ctxt_Fm_DPs		Abs:	1	Rel:	3
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2
	Fm_DPp		Abs:	1	Rel:	0.05
	Fm_TEOd		Abs:	1	Rel:	0.2

Layer: DPp
	Deep_Raw_Fm_DPs		Abs:	1	Rel:	1
	Fm_V2d		Abs:	1	Rel:	0.2
	Fm_V3d		Abs:	1	Rel:	0.5
	Fm_DPd		Abs:	1	Rel:	0.2
	Fm_TEOd		Abs:	1	Rel:	0.2

Layer: V4s
	Fm_V2s		Abs:	0.5	Rel:	2
	Fm_TEOs		Abs:	1	Rel:	0.1
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2

Layer: V4d
	Ctxt_Fm_V4s		Abs:	1	Rel:	4
	Fm_V1hp		Abs:	1	Rel:	0.2
	Fm_V1p		Abs:	1	Rel:	0.2
	Fm_V4p		Abs:	1	Rel:	0.05
	Fm_TEOd		Abs:	1	Rel:	0.2
	Fm_TEd		Abs:	1	Rel:	0.2
	Fm_TEOs		Abs:	1	Rel:	0.2

Layer: V4p
	Deep_Raw_Fm_V4s		Abs:	1	Rel:	1
	Fm_V2d		Abs:	1	Rel:	0.5
	Fm_V3d		Abs:	1	Rel:	0.5
	Fm_V4d		Abs:	1	Rel:	0.2
	Fm_TEOd		Abs:	1	Rel:	0.2

Layer: TEOs
	Fm_V4s		Abs:	1	Rel:	1
	Fm_V1hp		Abs:	1	Rel:	0.1
	Fm_V1p		Abs:	1	Rel:	0.1
	Fm_TEs		Abs:	1	Rel:	0.05

Layer: TEOd
	Ctxt_Fm_TEOs		Abs:	1	Rel:	4
	Ctxt_Fm_TEOd		Abs:	1	Rel:	4
	Fm_V1hp		Abs:	1	Rel:	0.1
	Fm_V1p		Abs:	1	Rel:	0.1
	Fm_V4p		Abs:	1	Rel:	0.2
	Fm_TEOp		Abs:	1	Rel:	0.05
	Fm_TEp		Abs:	1	Rel:	0.05
	Fm_TEd		Abs:	1	Rel:	0.1

Layer: TEOp
	Deep_Raw_Fm_TEOs		Abs:	1	Rel:	1
	Fm_V3d		Abs:	1	Rel:	0.2
	Fm_V4d		Abs:	1	Rel:	0.5
	Fm_TEOd		Abs:	1	Rel:	0.2
	Fm_TEd		Abs:	1	Rel:	0.5

Layer: TEs
	Fm_TEOs		Abs:	1	Rel:	1
	Fm_V1hp		Abs:	1	Rel:	0.1
	Fm_V1p		Abs:	1	Rel:	0.1

Layer: TEd
	Ctxt_Fm_TEs		Abs:	1	Rel:	4
	Ctxt_Fm_TEd		Abs:	1	Rel:	4
	Fm_V1hp		Abs:	1	Rel:	0.1
	Fm_V1p		Abs:	1	Rel:	0.1
	Fm_V4p		Abs:	1	Rel:	0.2
	Fm_TEOp		Abs:	1	Rel:	0.2
	Fm_TEp		Abs:	1	Rel:	0.05
	Fm_TEOd		Abs:	1	Rel:	0.1

Layer: TEp
	Deep_Raw_Fm_TEs		Abs:	1	Rel:	1
	Fm_V4d		Abs:	1	Rel:	0.5
	Fm_TEOd		Abs:	1	Rel:	0.2
	Fm_TEd		Abs:	1	Rel:	0.2

Layer: ObjIdTE
	Fm_TEs		Abs:	1	Rel:	1
	Fm_TEd		Abs:	1	Rel:	1

Layer: ObjIdTEO
	Fm_TEOs		Abs:	1	Rel:	1
	Fm_TEOd		Abs:	1	Rel:	1

Layer: ObjIdV4
	Fm_V4s		Abs:	1	Rel:	1
	Fm_V4d		Abs:	1	Rel:	1

Layer: ObjIdDP
	Fm_DPs		Abs:	1	Rel:	1
	Fm_DPd		Abs:	1	Rel:	1

Layer: ObjIdV3
	Fm_V3s		Abs:	1	Rel:	1
	Fm_V3d		Abs:	1	Rel:	1

Layer: ObjIdV2
	Fm_V2s		Abs:	1	Rel:	1

Layer: ObjIdV1
	Fm_V1s		Abs:	1	Rel:	1

Layer: ObjPosTEO
	Fm_TEOs		Abs:	1	Rel:	1
	Fm_TEOd		Abs:	1	Rel:	1

Layer: ObjPosV4
	Fm_V4s		Abs:	1	Rel:	1
	Fm_V4d		Abs:	1	Rel:	1

Layer: ObjPosDP
	Fm_DPs		Abs:	1	Rel:	1
	Fm_DPd		Abs:	1	Rel:	1
```

