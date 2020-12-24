# WWI3D Orig

This is the Go replication of the original cemer (C++ emergent) version of the model.

`wwi3d` does deep predictive learning of 3D objects tumbling through space, with periodic saccadic eye movements, providing plenty of opportunity for prediction errors.  **wwi** = *what, where integration*: both pathways combine to predict object -- *where* (dorsal) pathway is trained first and residual prediction error trains *what* pathway.

This is (an updated version of) the model described in:

* O’Reilly, R. C., Russin, J. L., Zolfaghar, M., & Rohrlich, J. (2020). Deep Predictive Learning in Neocortex and Pulvinar. ArXiv:2006.14800 [q-Bio]. http://arxiv.org/abs/2006.14800

# Install

See [Emergent Wiki Install](https://github.com/emer/emergent/wiki/Install) page for installation instructions -- basically you need install Go (e.g., `brew install go` on mac), then do `go build` in this directory.

Then, you need to get `CU3D100_20obj_8tick4sac.tar` from this [google drive folder](https://drive.google.com/drive/folders/13Mi9aUlF1A3sx3JaofX-qzKlxGoViT86?usp=sharing), which has the 3D rendered movies that the network is trained on.  Install it as `images` in the directory where this code is.  For example:

```bash
$ tar -xf CU3D100_20obj_8tick4sac.tar
$ mv CU3D100_20obj_8tick4sac images
```

(we usually have it in a centralized place and create a symbolic link, which works on the cluster too..)

# Running

Just run the wwi3d executable that is built with the `go build` command.  You can see how it processes processes input patterns, etc.  It takes about 1 day to train across 32 processors on our older cluster (use `go build -tags mpi` to build with mpi support), so it would take about 16 days without MPI.  Threading has decreasing benefits but is quite efficient for 2 threads, which is what it is configured for.

# Wt Scales

## cemer model

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

## Go model

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

Layer: V1hP
	             V2CTToV1hP		Abs:	1	Rel:	1
	             V3CTToV1hP		Abs:	1	Rel:	0.2
	             V4CTToV1hP		Abs:	1	Rel:	0.2
	            TEOCTToV1hP		Abs:	1	Rel:	0.1

Layer: V1mP
	             V2CTToV1mP		Abs:	1	Rel:	1
	             V3CTToV1mP		Abs:	1	Rel:	0.2
	             V4CTToV1mP		Abs:	1	Rel:	0.2
	            TEOCTToV1mP		Abs:	1	Rel:	0.1

Layer: V2
	                V1mToV2		Abs:	1	Rel:	1
	                V1hToV2		Abs:	1	Rel:	1
	                 V4ToV2		Abs:	1	Rel:	0.1
	                 V3ToV2		Abs:	1	Rel:	0.5
	               V1mPToV2		Abs:	1	Rel:	0.02
	               V1hPToV2		Abs:	1	Rel:	0.02
	                LIPToV2		Abs:	1	Rel:	0.5
	              TEOCTToV2		Abs:	1	Rel:	0.1

Layer: V2CT
	               V2ToV2CT		Abs:	1	Rel:	0.5
	             V1mPToV2CT		Abs:	1	Rel:	0.2
	             V1hPToV2CT		Abs:	1	Rel:	0.2
	            LIPCTToV2CT		Abs:	1	Rel:	1
	             LIPPToV2CT		Abs:	1	Rel:	0.2
	             V3CTToV2CT		Abs:	1	Rel:	0.5
	             V4CTToV2CT		Abs:	1	Rel:	0.5
	               V3ToV2CT		Abs:	1	Rel:	0.5
	              TEOToV2CT		Abs:	1	Rel:	0.5

Layer: V3
	                 V2ToV3		Abs:	0.5	Rel:	2
	                 DPToV3		Abs:	1	Rel:	0.2
	                 V4ToV3		Abs:	1	Rel:	0.2
	                LIPToV3		Abs:	1	Rel:	0.1
	                TEOToV3		Abs:	1	Rel:	0.1
	              TEOCTToV3		Abs:	1	Rel:	0.1
	               V1mPToV3		Abs:	1	Rel:	0.2
	               V1hPToV3		Abs:	1	Rel:	0.2
	                DPPToV3		Abs:	1	Rel:	0.05

Layer: V3CT
	               V3ToV3CT		Abs:	1	Rel:	1
	            LIPCTToV3CT		Abs:	1	Rel:	0.2
	             DPCTToV3CT		Abs:	1	Rel:	0.2
	             V4CTToV3CT		Abs:	1	Rel:	0.2
	               DPToV3CT		Abs:	1	Rel:	0.2
	               V4ToV3CT		Abs:	1	Rel:	0.2
	              TEOToV3CT		Abs:	1	Rel:	0.5
	             V1mPToV3CT		Abs:	1	Rel:	0.2
	             V1hPToV3CT		Abs:	1	Rel:	0.2
	              DPPToV3CT		Abs:	1	Rel:	0.2
	            LIPCTToV3CT		Abs:	1	Rel:	0.2

Layer: V3P
	              V2CTToV3P		Abs:	1	Rel:	0.5
	              DPCTToV3P		Abs:	1	Rel:	0.2
	             TEOCTToV3P		Abs:	1	Rel:	0.1

Layer: DP
	                 V3ToDP		Abs:	1	Rel:	1
	                 V2ToDP		Abs:	1	Rel:	1
	                TEOToDP		Abs:	1	Rel:	0.1
	               V1mPToDP		Abs:	1	Rel:	0.2
	               V1hPToDP		Abs:	1	Rel:	0.2
	                V3PToDP		Abs:	1	Rel:	0.1
	               TEOPToDP		Abs:	1	Rel:	0.1

Layer: DPCT
	               DPToDPCT		Abs:	1	Rel:	3
	              DPPToDPCT		Abs:	1	Rel:	0.05
	            TEOCTToDPCT		Abs:	1	Rel:	0.2
	             V1mPToDPCT		Abs:	1	Rel:	0.2
	             V1hPToDPCT		Abs:	1	Rel:	0.2

Layer: DPP
	              DPCTToDPP		Abs:	1	Rel:	0.2
	              V2CTToDPP		Abs:	1	Rel:	0.2
	              V3CTToDPP		Abs:	1	Rel:	0.5
	             TEOCTToDPP		Abs:	1	Rel:	0.2

Layer: V4
	                 V2ToV4		Abs:	0.5	Rel:	2
	                TEOToV4		Abs:	1	Rel:	0.1
	               V1mPToV4		Abs:	1	Rel:	0.2
	               V1hPToV4		Abs:	1	Rel:	0.2

Layer: V4CT
	               V4ToV4CT		Abs:	1	Rel:	4
	              V4PToV4CT		Abs:	1	Rel:	0.05
	            TEOCTToV4CT		Abs:	1	Rel:	0.2
	             TECTToV4CT		Abs:	1	Rel:	0.2
	              TEOToV4CT		Abs:	1	Rel:	0.2
	             V1mPToV4CT		Abs:	1	Rel:	0.2
	             V1hPToV4CT		Abs:	1	Rel:	0.2

Layer: V4P
	              V4CTToV4P		Abs:	1	Rel:	0.2
	              V2CTToV4P		Abs:	1	Rel:	0.5
	              V3CTToV4P		Abs:	1	Rel:	0.5
	             TEOCTToV4P		Abs:	1	Rel:	0.2

Layer: TEO
	                V4ToTEO		Abs:	1	Rel:	1
	                TEToTEO		Abs:	1	Rel:	0.05
	              V1mPToTEO		Abs:	1	Rel:	0.1
	              V1hPToTEO		Abs:	1	Rel:	0.1

Layer: TEOCT
	             TEOToTEOCT		Abs:	1	Rel:	4
	            TEOPToTEOCT		Abs:	1	Rel:	0.05
	           TEOCTToTEOCT		Abs:	1	Rel:	4
	            TECTToTEOCT		Abs:	1	Rel:	0.1
	             V4PToTEOCT		Abs:	1	Rel:	0.2
	             TEPToTEOCT		Abs:	1	Rel:	0.05
	            V1mPToTEOCT		Abs:	1	Rel:	0.1
	            V1hPToTEOCT		Abs:	1	Rel:	0.1

Layer: TEOP
	            TEOCTToTEOP		Abs:	1	Rel:	0.2
	             V3CTToTEOP		Abs:	1	Rel:	0.2
	             V4CTToTEOP		Abs:	1	Rel:	0.5
	             TECTToTEOP		Abs:	1	Rel:	0.5

Layer: TE
	                TEOToTE		Abs:	1	Rel:	1
	               V1mPToTE		Abs:	1	Rel:	0.1
	               V1hPToTE		Abs:	1	Rel:	0.1

Layer: TECT
	               TEToTECT		Abs:	1	Rel:	4
	              TEPToTECT		Abs:	1	Rel:	0.05
	             TECTToTECT		Abs:	1	Rel:	4
	            TEOCTToTECT		Abs:	1	Rel:	0.1
	              V4PToTECT		Abs:	1	Rel:	0.2
	             TEOPToTECT		Abs:	1	Rel:	0.2
	             V1mPToTECT		Abs:	1	Rel:	0.1
	             V1hPToTECT		Abs:	1	Rel:	0.1

Layer: TEP
	              TECTToTEP		Abs:	1	Rel:	0.2
	              V4CTToTEP		Abs:	1	Rel:	0.5
	             TEOCTToTEP		Abs:	1	Rel:	0.2
```

