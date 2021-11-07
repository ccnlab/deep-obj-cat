This is the parameter search notes for wwi3d.


```


Layer: V1m

Layer: V1h

Layer: LIP
	             MTPosToLIP		Abs:	0.5	Rel:	1
	              LIPPToLIP		Abs:	1	Rel:	0.2
	                V2ToLIP		Abs:	1	Rel:	0.1
	                V3ToLIP		Abs:	1	Rel:	0.1

Layer: LIPCT
	             LIPToLIPCT		Abs:	1	Rel:	1
	            LIPPToLIPCT		Abs:	1	Rel:	0.2
	         SaccadeToLIPCT		Abs:	1	Rel:	1
	          EyePosToLIPCT		Abs:	1	Rel:	1
	          ObjVelToLIPCT		Abs:	1	Rel:	1
	            V2CTToLIPCT		Abs:	1	Rel:	0.1
	            V3CTToLIPCT		Abs:	1	Rel:	0.1

Layer: LIPP
	            LIPCTToLIPP		Abs:	0.8	Rel:	1.25

Layer: MTPos
	             V1hToMTPos		Abs:	1	Rel:	1
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
	                 V2ToV2		Abs:	0.1	Rel:	1

Layer: V2CT
	               V2ToV2CT		Abs:	1	Rel:	1
	             V1mPToV2CT		Abs:	1	Rel:	0.2
	             V1hPToV2CT		Abs:	1	Rel:	0.2
	            LIPCTToV2CT		Abs:	1	Rel:	1
	             LIPPToV2CT		Abs:	1	Rel:	0.2
	             V3CTToV2CT		Abs:	1	Rel:	0.5
	             V4CTToV2CT		Abs:	1	Rel:	0.5
	               V3ToV2CT		Abs:	1	Rel:	0.5
	              TEOToV2CT		Abs:	1	Rel:	0.5
	             V2CTToV2CT		Abs:	0.1	Rel:	1

Layer: V3
	                 V2ToV3		Abs:	1	Rel:	1
	                 DPToV3		Abs:	1	Rel:	0.2
	               V1mPToV3		Abs:	1	Rel:	0.2
	               V1hPToV3		Abs:	1	Rel:	0.2
	                 V4ToV3		Abs:	1	Rel:	0.2
	                LIPToV3		Abs:	1	Rel:	0.1
	                TEOToV3		Abs:	1	Rel:	0.1
	              TEOCTToV3		Abs:	1	Rel:	0.1
	                 V3ToV3		Abs:	0.1	Rel:	1
	                V1mToV3		Abs:	1	Rel:	0.5

Layer: V3CT
	               V3ToV3CT		Abs:	1	Rel:	1
	             V1mPToV3CT		Abs:	1	Rel:	0.2
	             V1hPToV3CT		Abs:	1	Rel:	0.2
	            LIPCTToV3CT		Abs:	1	Rel:	0.2
	             DPCTToV3CT		Abs:	1	Rel:	0.2
	             V4CTToV3CT		Abs:	1	Rel:	0.2
	              TEOToV3CT		Abs:	1	Rel:	0.5
	               DPToV3CT		Abs:	1	Rel:	0.2
	               V4ToV3CT		Abs:	1	Rel:	0.2
	             V3CTToV3CT		Abs:	0.1	Rel:	1
	              V1mToV3CT		Abs:	1	Rel:	0.5

Layer: DP
	                 V3ToDP		Abs:	1	Rel:	1
	               V1mPToDP		Abs:	1	Rel:	0.2
	               V1hPToDP		Abs:	1	Rel:	0.2
	                TEOToDP		Abs:	1	Rel:	0.1
	                 DPToDP		Abs:	0.1	Rel:	1
	                V1mToDP		Abs:	1	Rel:	0.5

Layer: DPCT
	               DPToDPCT		Abs:	1	Rel:	1
	             V1mPToDPCT		Abs:	1	Rel:	0.2
	             V1hPToDPCT		Abs:	1	Rel:	0.2
	             DPCTToDPCT		Abs:	1	Rel:	0.5
	            TEOCTToDPCT		Abs:	1	Rel:	0.1
	             DPCTToDPCT		Abs:	0.1	Rel:	1
	              V1mToDPCT		Abs:	1	Rel:	0.5

Layer: V4
	                 V2ToV4		Abs:	1	Rel:	1
	                TEOToV4		Abs:	1	Rel:	0.1
	               V1mPToV4		Abs:	1	Rel:	0.2
	               V1hPToV4		Abs:	1	Rel:	0.2
	                 V4ToV4		Abs:	0.1	Rel:	1
	                V1mToV4		Abs:	1	Rel:	0.5

Layer: V4CT
	               V4ToV4CT		Abs:	1	Rel:	1
	             V1mPToV4CT		Abs:	1	Rel:	0.2
	             V1hPToV4CT		Abs:	1	Rel:	0.2
	              V4PToV4CT		Abs:	1	Rel:	0.05
	            TEOCTToV4CT		Abs:	1	Rel:	0.2
	              TEOToV4CT		Abs:	1	Rel:	0.2
	             TECTToV4CT		Abs:	1	Rel:	0.2
	             V4CTToV4CT		Abs:	0.1	Rel:	1
	              V1mToV4CT		Abs:	1	Rel:	0.5

Layer: V4P
	              V2CTToV4P		Abs:	1	Rel:	0.5
	              V3CTToV4P		Abs:	1	Rel:	0.5
	              V4CTToV4P		Abs:	1	Rel:	0.2
	             TEOCTToV4P		Abs:	1	Rel:	0.2

Layer: TEO
	                V4ToTEO		Abs:	1	Rel:	1
	                TEToTEO		Abs:	1	Rel:	0.05
	              V1mPToTEO		Abs:	1	Rel:	0.1
	              V1hPToTEO		Abs:	1	Rel:	0.1
	               TEOToTEO		Abs:	0.1	Rel:	1
	               V1mToTEO		Abs:	1	Rel:	0.5

Layer: TEOCT
	             TEOToTEOCT		Abs:	1	Rel:	1
	            V1mPToTEOCT		Abs:	1	Rel:	0.1
	            V1hPToTEOCT		Abs:	1	Rel:	0.1
	             V4PToTEOCT		Abs:	1	Rel:	0.2
	            TEOPToTEOCT		Abs:	1	Rel:	0.05
	           TEOCTToTEOCT		Abs:	1	Rel:	1
	            TECTToTEOCT		Abs:	1	Rel:	0.2
	           TEOCTToTEOCT		Abs:	0.1	Rel:	1
	             V1mToTEOCT		Abs:	1	Rel:	0.5

Layer: TEOP
	             V3CTToTEOP		Abs:	1	Rel:	0.2
	             V4CTToTEOP		Abs:	1	Rel:	0.5
	             TECTToTEOP		Abs:	1	Rel:	0.5
	            TEOCTToTEOP		Abs:	1	Rel:	0.2

Layer: TE
	                TEOToTE		Abs:	1	Rel:	1
	               V1mPToTE		Abs:	1	Rel:	0.1
	               V1hPToTE		Abs:	1	Rel:	0.1
	                 TEToTE		Abs:	0.1	Rel:	1
	                V1mToTE		Abs:	1	Rel:	0.5

Layer: TECT
	               TEToTECT		Abs:	1	Rel:	1
	             TECTToTECT		Abs:	1	Rel:	1
	             V1mPToTECT		Abs:	1	Rel:	0.1
	             V1hPToTECT		Abs:	1	Rel:	0.1
	              V4PToTECT		Abs:	1	Rel:	0.2
	             TEOPToTECT		Abs:	1	Rel:	0.2
	            TEOCTToTECT		Abs:	1	Rel:	0.2
	             TECTToTECT		Abs:	0.1	Rel:	1
	              V1mToTECT		Abs:	1	Rel:	0.5

```

