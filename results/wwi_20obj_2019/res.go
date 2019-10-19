// analyze overall results

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/emer/etable/clust"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/emer/etable/metric"
	"github.com/emer/etable/simat"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
)

// this is the stub main for gogi that calls our actual
// mainrun function, at end of file
func main() {
	gimain.Main(func() {
		mainrun()
	})
}

/* output:
V1_V1Cat  avg contrast dist: 0.2448
V1_BpCat  avg contrast dist: 0.1078
V1_LbaCat  avg contrast dist: 0.1796
Lba_LbaCat  avg contrast dist: 0.5071
Lba_V1Cat  avg contrast dist: 0.3070
Lba_BpCat  avg contrast dist: 0.2645
BpPred_BpCat  avg contrast dist: 0.0838
BpPred_V1Cat  avg contrast dist: 0.0513
BpPred_LbaCat  avg contrast dist: 0.0585
BpEnc_BpCat  avg contrast dist: 0.0050
BpEnc_V1Cat  avg contrast dist: 0.0078
PredNet_BpCat  avg contrast dist: 0.0095
PredNet_V1Cat  avg contrast dist: 0.0250
PredNet_LbaCat  avg contrast dist: 0.0196
PredNet_PNCat  avg contrast dist: 0.0253
Expt1_LbaCat  avg contrast dist: 0.3083
Expt1_BpCat  avg contrast dist: 0.0643
Expt1_V1Cat  avg contrast dist: 0.2583
Expt1_Ex5Cat  avg contrast dist: 0.3225
*/

// 20 Object categs: IMPORTANT: do not change the order of this list as it is used
// in various places as the cannonical ordering for e.g., Expt1 data
var Objs = []string{
	"banana",
	"layercake",
	"trafficcone",
	"sailboat",
	"trex",
	"person",
	"guitar",
	"tablelamp",
	"doorknob",
	"handgun",
	"donut",
	"chair",
	"slrcamera",
	"elephant",
	"piano",
	"fish",
	"car",
	"heavycannon",
	"stapler",
	"motorcycle",
}

var ObjIdxs map[string]int

// LbaCats3 is best-fitting 3-category leabra: -0.5399
var LbaCats3 = map[string]string{
	"banana":      "1-pyramid",
	"layercake":   "1-pyramid",
	"trafficcone": "1-pyramid",
	"sailboat":    "1-pyramid",
	"trex":        "1-pyramid",
	"guitar":      "1-pyramid",
	"person":      "1-pyramid",
	"tablelamp":   "1-pyramid",
	"chair":       "2-box",
	"slrcamera":   "2-box",
	"elephant":    "2-box",
	"piano":       "2-box",
	"fish":        "2-box",
	"donut":       "2-box",
	"handgun":     "2-box",
	"doorknob":    "2-box",
	"car":         "3-horiz",
	"heavycannon": "3-horiz",
	"stapler":     "3-horiz",
	"motorcycle":  "3-horiz",
}

var LbaCats = LbaCats5

// LbaCats5 is best-fitting 5-category leabra = -0.5071 -- can get to -0.5526 by
// 2-vertical = tablelamp only, 3-round = chair only
// this is best compromize with good dist score and shape similarity (via Expt)
// called "Centroid" in paper
var LbaCats5 = map[string]string{
	"banana":      "1-pyramid",
	"layercake":   "1-pyramid",
	"trafficcone": "1-pyramid",
	"sailboat":    "1-pyramid",
	"trex":        "1-pyramid",
	"person":      "2-vertical",
	"guitar":      "2-vertical",
	"tablelamp":   "2-vertical",
	"doorknob":    "3-round",
	"donut":       "3-round",
	"handgun":     "3-round",
	"chair":       "3-round",
	"slrcamera":   "4-box",
	"elephant":    "4-box",
	"piano":       "4-box",
	"fish":        "4-box",
	"car":         "5-horiz",
	"heavycannon": "5-horiz",
	"stapler":     "5-horiz",
	"motorcycle":  "5-horiz",
}

// Expt1Cats5 is best-fitting 5-category expt1 = -0.3225
var Expt1Cats5 = map[string]string{
	"banana":      "1-pyramid",
	"layercake":   "1-pyramid",
	"trafficcone": "1-pyramid",
	"sailboat":    "1-pyramid",
	"chair":       "1-pyramid",
	"person":      "2-vertical",
	"guitar":      "2-vertical",
	"tablelamp":   "2-vertical",
	"doorknob":    "3-round",
	"donut":       "3-round",
	"slrcamera":   "4-box",
	"elephant":    "4-box",
	"piano":       "4-box",
	"fish":        "4-box",
	"car":         "5-horiz",
	"heavycannon": "5-horiz",
	"stapler":     "5-horiz",
	"motorcycle":  "5-horiz",
	"trex":        "5-horiz",
	"handgun":     "5-horiz",
}

// Expt1Cats3 is best-fitting 3-categ expt1, worse than 5: -0.2739
var Expt1Cats3 = map[string]string{
	"trafficcone": "1-pyramid",
	"sailboat":    "1-pyramid",
	"guitar":      "1-pyramid",
	"person":      "1-pyramid",
	"tablelamp":   "1-pyramid",
	"chair":       "2-box",
	"slrcamera":   "2-box",
	"elephant":    "2-box",
	"piano":       "2-box",
	"fish":        "2-box",
	"donut":       "2-box",
	"doorknob":    "2-box",
	"handgun":     "2-box",
	"motorcycle":  "2-box",
	"car":         "3-horiz",
	"heavycannon": "3-horiz",
	"stapler":     "3-horiz",
	"trex":        "3-horiz",
	"layercake":   "3-horiz",
	"banana":      "3-horiz",
}

var JustLbaCats []string
var LbaCatsBlanks []string // with blanks

var BpCats = map[string]string{
	"banana":      "cat1",
	"layercake":   "cat1",
	"trafficcone": "cat1",
	"sailboat":    "cat1",
	"trex":        "cat1",
	"person":      "cat1",
	"guitar":      "cat1",
	"tablelamp":   "cat1",
	"doorknob":    "cat1",
	"donut":       "cat1",
	"handgun":     "cat1",
	"chair":       "cat1",
	"slrcamera":   "cat1",
	"elephant":    "cat1",
	"piano":       "cat1",
	"fish":        "cat2",
	"car":         "cat2",
	"heavycannon": "cat2",
	"stapler":     "cat2",
	"motorcycle":  "cat2",
}

var V1Cats = map[string]string{
	"trafficcone": "cat1",
	"sailboat":    "cat1",
	"person":      "cat1",
	"guitar":      "cat1",
	"tablelamp":   "cat1",
	"chair":       "cat1",
	"layercake":   "cat2",
	"trex":        "cat2",
	"doorknob":    "cat2",
	"donut":       "cat2",
	"handgun":     "cat2",
	"slrcamera":   "cat2",
	"elephant":    "cat2",
	"piano":       "cat2",
	"fish":        "cat2",
	"car":         "cat2",
	"heavycannon": "cat2",
	"stapler":     "cat2",
	"motorcycle":  "cat2",
	"banana":      "cat3",
}

// 0.2820 = best 3 categ
var PredNetCats3 = map[string]string{
	"tablelamp":   "cat1",
	"person":      "cat1",
	"guitar":      "cat1",
	"trafficcone": "cat1",
	"sailboat":    "cat1",
	"layercake":   "cat1",
	"elephant":    "cat2",
	"donut":       "cat2",
	"banana":      "cat2",
	"handgun":     "cat2",
	"slrcamera":   "cat2",
	"trex":        "cat2",
	"car":         "cat2",
	"heavycannon": "cat2",
	"motorcycle":  "cat2",
	"stapler":     "cat2",
	"fish":        "cat2",
	"doorknob":    "cat3",
	"chair":       "cat3",
	"piano":       "cat3",
}

// 0.2546 = best 2 categ
var PredNetCats2 = map[string]string{
	"tablelamp":   "cat1",
	"person":      "cat1",
	"guitar":      "cat1",
	"trafficcone": "cat1",
	"chair":       "cat1",
	"sailboat":    "cat1",
	"layercake":   "cat1",
	"elephant":    "cat2",
	"piano":       "cat2",
	"donut":       "cat2",
	"doorknob":    "cat2",
	"banana":      "cat2",
	"handgun":     "cat2",
	"slrcamera":   "cat2",
	"trex":        "cat2",
	"car":         "cat2",
	"heavycannon": "cat2",
	"motorcycle":  "cat2",
	"stapler":     "cat2",
	"fish":        "cat2",
}

// Res is the main data structure for all expt results and tables
// is visualized in gui so you can click on stuff..
type Res struct {
	LbaFullSimMat          simat.SimMat  `desc:"Leabra TEs full similarity matrix"`
	LbaFullNames           []string      `view:"-" desc:"object names in order for FullSimMat"`
	LbaLbaCatSimMat        simat.SimMat  `desc:"Leabra TEs full similarity matrix sorted fresh in Lba cat order"`
	LbaV1CatSimMat         simat.SimMat  `desc:"Leabra TEs full similarity matrix, in V1 cat order"`
	LbaBpCatSimMat         simat.SimMat  `desc:"Leabra TEs full similarity matrix, in Bp cat order"`
	LbaV4FullSimMat        simat.SimMat  `desc:"Leabra V4s full similarity matrix"`
	V1FullSimMat           simat.SimMat  `desc:"V1 full similarity matrix"`
	V1FullNames            []string      `view:"-" desc:"object names in order for FullSimMat"`
	V1V1CatSimMat          simat.SimMat  `desc:"V1 in V1 Cat order"`
	V1BpCatSimMat          simat.SimMat  `desc:"V1 in Bp Cat order"`
	V1LbaCatSimMat         simat.SimMat  `desc:"V1 in Lba Cat order"`
	Lba200SimMat           simat.SimMat  `desc:"Leabra TEs full similarity matrix, 200 epcs"`
	Lba200Names            []string      `view:"-" desc:"object names in order"`
	Lba600SimMat           simat.SimMat  `desc:"Leabra TEs full similarity matrix, 600 epcs"`
	Lba600Names            []string      `view:"-" desc:"object names in order"`
	BpPredFullSimMat       simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix"`
	BpPredFullNames        []string      `view:"-" desc:"object names in order for FullSimMat"`
	BpPredBpCatSimMat      simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix, in Bp Cat order"`
	BpPredV1CatSimMat      simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix, in V1 Cat order"`
	BpPredLbaCatSimMat     simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix, in Lba Cat order"`
	BpEncFullSimMat        simat.SimMat  `desc:"WWI Bp Encoder full similarity matrix"`
	BpEncFullNames         []string      `view:"-" desc:"object names in order for FullSimMat"`
	BpEncBpCatSimMat       simat.SimMat  `desc:"WWI Bp Encoder full similarity matrix, in Bp Cat order"`
	BpEncV1CatSimMat       simat.SimMat  `desc:"WWI Bp Encoder full similarity matrix, in Bp Cat order"`
	PredNetFullSimMat      simat.SimMat  `desc:"PredNet predictor full similarity matrix"`
	PredNetFullNames       []string      `view:"-" desc:"object names in order for FullSimMat"`
	PredNetBpCatSimMat     simat.SimMat  `desc:"PredNet predictor in Bp Cat order"`
	PredNetV1CatSimMat     simat.SimMat  `desc:"PredNet predictor in V1 Cat order"`
	PredNetLbaCatSimMat    simat.SimMat  `desc:"PredNet predictor in Lba Cat order"`
	PredNetPNCatSimMat     simat.SimMat  `desc:"PredNet predictor in PN Cat order"`
	PredNetPixelSimMat     simat.SimMat  `desc:"PredNet predictor full similarity matrix, pixel layer"`
	PredNetLay0SimMat      simat.SimMat  `desc:"PredNet predictor full similarity matrix, layer 0"`
	PredNetPixV1CatSimMat  simat.SimMat  `desc:"PredNet predictor full similarity matrix, pixel layer, v1 cats"`
	PredNetLay0V1CatSimMat simat.SimMat  `desc:"PredNet predictor full similarity matrix, layer 0, v1 cats"`
	Expt1SimMat            simat.SimMat  `desc:"Expt1 similarity matrix"`
	Expt1LbaSimMat         simat.SimMat  `desc:"Expt1 similarity matrix, leabra sorted"`
	Expt1Ex5SimMat         simat.SimMat  `desc:"Expt1 similarity matrix, v1 sorted"`
	Expt1BpSimMat          simat.SimMat  `desc:"Expt1 similarity matrix, bp sorted"`
	Expt1V1SimMat          simat.SimMat  `desc:"Expt1 similarity matrix, v1 sorted"`
	LbaObjSimMat           simat.SimMat  `desc:"Leabra TEs obj-cat reduced similarity matrix"`
	V1ObjSimMat            simat.SimMat  `desc:"V1 obj-cat reduced similarity matrix"`
	BpPredObjSimMat        simat.SimMat  `desc:"WWI Bp Predictive obj-cat reduced similarity matrix"`
	BpEncObjSimMat         simat.SimMat  `desc:"WWI Bp Encoder obj-cat reduced similarity matrix"`
	LbaTickSimMat          simat.SimMat  `desc:"Leabra TEs full similarity matrix, by tick"`
	LbaTickNames           []string      `view:"-" desc:"object names in order"`
	ExptCorrel             etable.Table  `desc:"correlations with expt data for each sim data"`
	Expt1ClustPlot         *eplot.Plot2D `desc:"cluster plot"`
	LbaObjClustPlot        *eplot.Plot2D `desc:"cluster plot"`
	LbaFullClustPlot       *eplot.Plot2D `desc:"cluster plot"`
}

func (rs *Res) Init() {
	if ObjIdxs == nil {
		no := len(Objs)
		ObjIdxs = make(map[string]int, no)
		JustLbaCats = make([]string, no)
		LbaCatsBlanks = make([]string, no)
		lstcat := ""
		for i, o := range Objs {
			ObjIdxs[o] = i
			cat := LbaCats[o]
			JustLbaCats[i] = cat
			if cat != lstcat {
				LbaCatsBlanks[i] = cat
				lstcat = cat
			}
		}
	}
}

func (rs *Res) OpenFullSimMat(sm *simat.SimMat, nms *[]string, fname string, lab string, maxv string) {
	ltab := &etable.Table{}
	err := ltab.OpenCSV(gi.FileName(lab), etable.Tab)
	if err != nil {
		log.Println(err)
		return
	}
	no := ltab.Rows
	// fmt.Printf("rows: %v\n", no)
	*nms = make([]string, no)
	sm.Init()
	smat := sm.Mat.(*etensor.Float64)
	smat.SetShape([]int{no, no}, nil, nil)
	smat.SetMetaData("max", maxv)
	smat.SetMetaData("min", "0")
	smat.SetMetaData("colormap", "Viridis")
	smat.SetMetaData("grid-fill", "1")
	smat.SetMetaData("dim-extra", "0.5")
	err = etensor.OpenCSV(smat, gi.FileName(fname), etable.Tab)
	if err != nil {
		log.Println(err)
		return
	}
	cl, err := ltab.ColByNameTry("categ")
	if err != nil {
		log.Println(err)
		return
	}
	svals := cl.(*etensor.String).Values
	sm.Rows = simat.BlankRepeat(svals)
	sm.Cols = sm.Rows

	cl, err = ltab.ColByNameTry("group_name")
	if err != nil {
		log.Println(err)
		return
	}
	svals = cl.(*etensor.String).Values
	for ri, nm := range svals {
		ui := strings.Index(nm, "_")
		if ui > 0 {
			nm = nm[0:ui]
		}
		_, ok := ObjIdxs[nm]
		if !ok {
			fmt.Printf("%v not found\n", nm)
		}
		(*nms)[ri] = nm
	}
}

func (rs *Res) OpenFullSimMatPredNet(sm *simat.SimMat, nms *[]string, fname string, lab string, maxv string) {
	no := 156 // known
	ltab := &etensor.String{}
	ltab.SetShape([]int{no}, nil, nil)
	err := etensor.OpenCSV(ltab, gi.FileName(lab), etable.Comma)
	if err != nil {
		log.Println(err)
		return
	}
	*nms = make([]string, no)
	sm.Init()
	smat := sm.Mat.(*etensor.Float64)
	smat.SetShape([]int{no, no}, nil, nil)
	smat.SetMetaData("max", maxv)
	smat.SetMetaData("min", "0")
	smat.SetMetaData("colormap", "Viridis")
	smat.SetMetaData("grid-fill", "1")
	smat.SetMetaData("dim-extra", "0.5")
	err = etensor.OpenCSV(smat, gi.FileName(fname), etable.Tab)
	if err != nil {
		log.Println(err)
		return
	}
	// for i, v := range smat.Values { // getting correlations here, not 1-correls
	// 	smat.Values[i] = 1 - v
	// }
	svals := ltab.Values
	sm.Rows = simat.BlankRepeat(svals)
	sm.Cols = sm.Rows

	for ri, nm := range svals {
		ui := strings.Index(nm, "_")
		if ui > 0 {
			nm = nm[0:ui]
		}
		_, ok := ObjIdxs[nm]
		if !ok {
			fmt.Printf("%v not found\n", nm)
		}
		(*nms)[ri] = nm
	}
}

// CatSortSimMat takes an input sim matrix and categorizes the items according to given cats
// and then sorts items within that according to their average within - between cat similarity
func (rs *Res) CatSortSimMat(insm *simat.SimMat, osm *simat.SimMat, nms []string, catmap map[string]string, contrast bool, name string) {
	no := len(insm.Rows)
	sch := etable.Schema{
		{"Cat", etensor.STRING, nil, nil},
		{"Dist", etensor.FLOAT64, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, no)
	cats := dt.Cols[0].(*etensor.String).Values
	dists := dt.Cols[1].(*etensor.Float64).Values
	for i, nm := range nms {
		cats[i] = catmap[nm]
	}
	smatv := insm.Mat.(*etensor.Float64).Values
	avgCtrstDist := 0.0
	for ri := 0; ri < no; ri++ {
		roff := ri * no
		aid := 0.0
		ain := 0
		abd := 0.0
		abn := 0
		rc := cats[ri]
		for ci := 0; ci < no; ci++ {
			if ri == ci {
				continue
			}
			cc := cats[ci]
			d := smatv[roff+ci]
			if cc == rc {
				aid += d
				ain++
			} else {
				abd += d
				abn++
			}
		}
		if ain > 0 {
			aid /= float64(ain)
		}
		if abn > 0 {
			abd /= float64(abn)
		}
		dval := aid
		if contrast {
			dval -= abd
		}
		dists[ri] = dval
		avgCtrstDist += (1 - aid) - (1 - abd)
	}
	avgCtrstDist /= float64(no)
	ix := etable.NewIdxView(dt)
	ix.SortColNames([]string{"Cat", "Dist"}, true) // ascending
	osm.Init()
	osm.Mat.CopyShapeFrom(insm.Mat)
	osm.Mat.CopyMetaData(insm.Mat)
	omatv := osm.Mat.(*etensor.Float64).Values
	bcols := make([]string, no)
	last := ""
	for sri := 0; sri < no; sri++ {
		sroff := sri * no
		ri := ix.Idxs[sri]
		roff := ri * no
		cat := cats[ri]
		if cat != last {
			bcols[sri] = cat
			last = cat
		}
		// bcols[sri] = nms[ri] // uncomment this to see all the names
		for sci := 0; sci < no; sci++ {
			ci := ix.Idxs[sci]
			d := smatv[roff+ci]
			omatv[sroff+sci] = d
		}
	}
	osm.Rows = bcols
	osm.Cols = bcols
	fmt.Printf("%v  avg contrast dist: %.4f\n", name, avgCtrstDist)
}

func (rs *Res) OpenSimMats() {
	rs.OpenFullSimMat(&rs.LbaFullSimMat, &rs.LbaFullNames, "sim_leabra_simat.tsv", "sim_leabra_simat_lbl.tsv", "1.5")
	rs.OpenFullSimMat(&rs.LbaV4FullSimMat, &rs.LbaFullNames, "sim_leabra_simat_v4.tsv", "sim_leabra_simat_lbl.tsv", "1.0")
	rs.OpenFullSimMat(&rs.V1FullSimMat, &rs.V1FullNames, "sim_v1_simat.tsv", "sim_v1_simat_lbl.tsv", "1.0")
	rs.OpenFullSimMat(&rs.BpPredFullSimMat, &rs.BpPredFullNames, "sim_bp_pred_simat.tsv", "sim_bp_pred_simat_lbl.tsv", "0.3")
	rs.OpenFullSimMat(&rs.BpEncFullSimMat, &rs.BpEncFullNames, "sim_bp_enc_simat.tsv", "sim_bp_enc_simat_lbl.tsv", "0.04")

	rs.OpenFullSimMat(&rs.Lba200SimMat, &rs.Lba200Names, "sim_leabra_simat_200epc.tsv", "sim_leabra_simat_200epc_lbl.tsv", "1.5")
	rs.OpenFullSimMat(&rs.Lba600SimMat, &rs.Lba600Names, "sim_leabra_simat_600epc.tsv", "sim_leabra_simat_600epc_lbl.tsv", "1.5")

	// rs.OpenFullSimMatPredNet(&rs.PredNetFullSimMat, &rs.PredNetFullNames, "prednet_layer3.csv", "prednet_labels.csv", "0.15")
	// rs.OpenFullSimMatPredNet(&rs.PredNetPixelSimMat, &rs.PredNetFullNames, "prednet_pixels.csv", "prednet_labels.csv", "0.06")
	// rs.OpenFullSimMatPredNet(&rs.PredNetLay0SimMat, &rs.PredNetFullNames, "prednet_layer0.csv", "prednet_labels.csv", "0.04")

	rs.OpenFullSimMatPredNet(&rs.PredNetFullSimMat, &rs.PredNetFullNames, "prednet_64x64_6l_dropout0p1_layer6.csv", "prednet_64x64_6l_dropout0p1_labels.csv", "0.75")
	rs.OpenFullSimMatPredNet(&rs.PredNetPixelSimMat, &rs.PredNetFullNames, "prednet_64x64_6l_dropout0p1_pixels.csv", "prednet_64x64_6l_dropout0p1_labels.csv", "0.06")
	rs.OpenFullSimMatPredNet(&rs.PredNetLay0SimMat, &rs.PredNetFullNames, "prednet_64x64_6l_dropout0p1_layer1.csv", "prednet_64x64_6l_dropout0p1_labels.csv", "0.04")

	// bool arg = use within - between (else just within)
	rs.CatSortSimMat(&rs.V1FullSimMat, &rs.V1V1CatSimMat, rs.V1FullNames, V1Cats, true, "V1_V1Cat")
	rs.CatSortSimMat(&rs.V1FullSimMat, &rs.V1BpCatSimMat, rs.V1FullNames, BpCats, true, "V1_BpCat")
	rs.CatSortSimMat(&rs.V1FullSimMat, &rs.V1LbaCatSimMat, rs.V1FullNames, LbaCats, true, "V1_LbaCat")
	rs.CatSortSimMat(&rs.LbaFullSimMat, &rs.LbaLbaCatSimMat, rs.LbaFullNames, LbaCats, true, "Lba_LbaCat")
	rs.CatSortSimMat(&rs.LbaFullSimMat, &rs.LbaV1CatSimMat, rs.LbaFullNames, V1Cats, true, "Lba_V1Cat")
	rs.CatSortSimMat(&rs.LbaFullSimMat, &rs.LbaBpCatSimMat, rs.LbaFullNames, BpCats, true, "Lba_BpCat")

	sm200 := &simat.SimMat{}
	rs.CatSortSimMat(&rs.Lba200SimMat, sm200, rs.Lba200Names, LbaCats, true, "Lba200_LbaCat")
	rs.Lba200SimMat = *sm200
	sm600 := &simat.SimMat{}
	rs.CatSortSimMat(&rs.Lba600SimMat, sm600, rs.Lba600Names, LbaCats, true, "Lba600_LbaCat")
	rs.Lba600SimMat = *sm600

	rs.CatSortSimMat(&rs.BpPredFullSimMat, &rs.BpPredBpCatSimMat, rs.BpPredFullNames, BpCats, true, "BpPred_BpCat")
	rs.CatSortSimMat(&rs.BpPredFullSimMat, &rs.BpPredV1CatSimMat, rs.BpPredFullNames, V1Cats, true, "BpPred_V1Cat")
	rs.CatSortSimMat(&rs.BpPredFullSimMat, &rs.BpPredLbaCatSimMat, rs.BpPredFullNames, LbaCats, true, "BpPred_LbaCat")
	rs.CatSortSimMat(&rs.BpEncFullSimMat, &rs.BpEncBpCatSimMat, rs.BpEncFullNames, BpCats, true, "BpEnc_BpCat")
	rs.CatSortSimMat(&rs.BpEncFullSimMat, &rs.BpEncV1CatSimMat, rs.BpEncFullNames, V1Cats, true, "BpEnc_V1Cat")
	rs.CatSortSimMat(&rs.PredNetFullSimMat, &rs.PredNetBpCatSimMat, rs.PredNetFullNames, BpCats, false, "PredNet_BpCat")       // doesn't work with contrast as is too noisy
	rs.CatSortSimMat(&rs.PredNetFullSimMat, &rs.PredNetV1CatSimMat, rs.PredNetFullNames, V1Cats, false, "PredNet_V1Cat")       // doesn't work with contrast as is too noisy
	rs.CatSortSimMat(&rs.PredNetFullSimMat, &rs.PredNetLbaCatSimMat, rs.PredNetFullNames, LbaCats, false, "PredNet_LbaCat")    // doesn't work with contrast as is too noisy
	rs.CatSortSimMat(&rs.PredNetFullSimMat, &rs.PredNetPNCatSimMat, rs.PredNetFullNames, PredNetCats3, false, "PredNet_PNCat") // doesn't work with contrast as is too noisy

	rs.CatSortSimMat(&rs.PredNetPixelSimMat, &rs.PredNetPixV1CatSimMat, rs.PredNetFullNames, V1Cats, false, "PredNetPixels_V1Cat")
	rs.CatSortSimMat(&rs.PredNetLay0SimMat, &rs.PredNetLay0V1CatSimMat, rs.PredNetFullNames, V1Cats, false, "PredNetLayer0_V1Cat")

	// rs.OpenFullSimMat(&rs.LbaTickSimMat, &rs.LbaTickNames, "sim_leabra_simat_bytick.tsv", "sim_leabra_simat_bytick_lbl.tsv", "1.5")

}

// ObjSimMat compresses full simat into a much smaller per-object sim mat
func (rs *Res) ObjSimMat(fsm *simat.SimMat, nms []string, osm *simat.SimMat, maxv string) {
	fsmat := fsm.Mat.(*etensor.Float64)

	ono := len(Objs)
	osm.Init()
	osmat := osm.Mat.(*etensor.Float64)
	osmat.SetShape([]int{ono, ono}, nil, nil)
	osm.Rows = LbaCatsBlanks
	osm.Cols = LbaCatsBlanks
	osmat.SetMetaData("max", maxv)
	osmat.SetMetaData("min", "0")
	osmat.SetMetaData("colormap", "Viridis")
	osmat.SetMetaData("grid-fill", "1")
	osmat.SetMetaData("dim-extra", "0.15")

	nmat := &etensor.Float64{}
	nmat.SetShape([]int{ono, ono}, nil, nil)

	nf := len(nms)
	for ri := 0; ri < nf; ri++ {
		roi := ObjIdxs[nms[ri]]
		for ci := 0; ci < nf; ci++ {
			sidx := ri*nf + ci
			sval := fsmat.Values[sidx]
			coi := ObjIdxs[nms[ci]]
			oidx := roi*ono + coi
			osmat.Values[oidx] += sval
			nmat.Values[oidx] += 1
		}
	}
	for ri := 0; ri < ono; ri++ {
		for ci := 0; ci < ono; ci++ {
			oidx := ri*ono + ci
			osmat.Values[oidx] /= nmat.Values[oidx]
		}
	}
}

func (rs *Res) ObjSimMats() {
	rs.ObjSimMat(&rs.LbaFullSimMat, rs.LbaFullNames, &rs.LbaObjSimMat, "1.5")
	rs.ObjSimMat(&rs.V1FullSimMat, rs.V1FullNames, &rs.V1ObjSimMat, "1.0")
	rs.ObjSimMat(&rs.BpPredFullSimMat, rs.BpPredFullNames, &rs.BpPredObjSimMat, "0.23")
	rs.ObjSimMat(&rs.BpEncFullSimMat, rs.BpEncFullNames, &rs.BpEncObjSimMat, "0.032")
}

func (rs *Res) OpenExptMat() {
	no := len(Objs)
	sm := &rs.Expt1SimMat
	sm.Init()
	smat := sm.Mat.(*etensor.Float64)
	smat.SetShape([]int{no, no}, nil, nil)
	err := etensor.OpenCSV(smat, gi.FileName("expt1_simat.csv"), etable.Comma)
	if err != nil {
		log.Println(err)
		return
	}
	sm.Rows = LbaCatsBlanks
	sm.Cols = LbaCatsBlanks
	smat.SetMetaData("max", "1")
	smat.SetMetaData("min", "0")
	smat.SetMetaData("colormap", "Viridis")
	smat.SetMetaData("grid-fill", "1")
	smat.SetMetaData("dim-extra", "0.15")
}

func (rs *Res) TestExptMats() {
	rs.CatSortSimMat(&rs.Expt1SimMat, &rs.Expt1LbaSimMat, Objs, LbaCats, true, "Expt1_LbaCat")
	rs.CatSortSimMat(&rs.Expt1SimMat, &rs.Expt1BpSimMat, Objs, BpCats, true, "Expt1_BpCat")
	rs.CatSortSimMat(&rs.Expt1SimMat, &rs.Expt1V1SimMat, Objs, V1Cats, true, "Expt1_V1Cat")
	rs.CatSortSimMat(&rs.Expt1SimMat, &rs.Expt1Ex5SimMat, Objs, Expt1Cats5, true, "Expt1_Ex5Cat")
}

func (rs *Res) SetCorrel(dt *etable.Table, row int, nm string, smat *simat.SimMat) {
	svals := smat.Mat.(*etensor.Float64).Values
	evals := rs.Expt1SimMat.Mat.(*etensor.Float64).Values
	cosine := metric.Cosine64(svals, evals)
	dt.SetCellFloat("Num", row, float64(row))
	dt.SetCellString("Sim", row, nm)
	dt.SetCellFloat("Cosine", row, cosine)
}

func (rs *Res) Correls() {
	dt := &rs.ExptCorrel
	sch := etable.Schema{
		{"Num", etensor.FLOAT64, nil, nil},
		{"Sim", etensor.STRING, nil, nil},
		{"Cosine", etensor.FLOAT64, nil, nil},
	}
	nsim := 4
	dt.SetFromSchema(sch, nsim)
	rs.SetCorrel(dt, 0, "Leabra", &rs.LbaObjSimMat)
	rs.SetCorrel(dt, 1, "V1", &rs.V1ObjSimMat)
	rs.SetCorrel(dt, 2, "Bp Pred", &rs.BpPredObjSimMat)
	rs.SetCorrel(dt, 3, "Bp Enc", &rs.BpEncObjSimMat)
}

func (rs *Res) ClustObj(smat *simat.SimMat, title string) *eplot.Plot2D {
	prv := smat.Rows
	// smat.Rows = JustLbaCats
	// smat.Cols = JustLbaCats
	smat.Rows = Objs
	smat.Cols = Objs
	cl := clust.Glom(smat, clust.MaxDist) // ContrastDist, MaxDist, Avg all produce similar good fits
	// then plot the results
	pt := &etable.Table{}
	clust.Plot(pt, cl, smat)
	plt := &eplot.Plot2D{}
	plt.InitName(plt, "ClustPlot")
	plt.Params.Title = title
	plt.Params.XAxisCol = "X"
	plt.Params.Scale = 3
	plt.SetTable(pt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("X", false, true, 0, false, 0)
	plt.SetColParams("Y", true, true, 0, false, 0)
	plt.SetColParams("Label", true, false, 0, false, 0)
	smat.Rows = prv
	smat.Cols = prv
	return plt
}

// GlomInit returns a standard root node initialized with all of the leaves
func (rs *Res) ClustFull(smat *simat.SimMat, nms []string, title string) *eplot.Plot2D {
	prv := smat.Rows
	smat.Rows = nms
	smat.Cols = nms

	// pre-allocate all objects into clusters
	no := len(Objs)
	root := &clust.Node{}
	root.Kids = make([]*clust.Node, no)
	for i := 0; i < no; i++ {
		ond := &clust.Node{Dist: 0.1}
		kidx := []int{}
		onm := Objs[i]
		for ni, nm := range nms {
			if nm == onm {
				kidx = append(kidx, ni)
			}
		}
		ond.Kids = make([]*clust.Node, len(kidx))
		for ki, kix := range kidx {
			ond.Kids[ki] = &clust.Node{Idx: kix}
		}
		root.Kids[i] = ond
	}

	cl := clust.GlomClust(root, smat, clust.ContrastDist) // ContrastDist, MaxDist, Avg all produce similar good fits
	// then plot the results
	pt := &etable.Table{}
	clust.Plot(pt, cl, smat)
	plt := &eplot.Plot2D{}
	plt.InitName(plt, "ClustPlot")
	plt.Params.Title = title
	plt.Params.XAxisCol = "X"
	plt.Params.Scale = 3
	plt.SetTable(pt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("X", false, true, 0, false, 0)
	plt.SetColParams("Y", true, true, 0, false, 0)
	plt.SetColParams("Label", true, false, 0, false, 0)

	smat.Rows = prv
	smat.Cols = prv
	return plt
}

func (rs *Res) ClustPlots() {
	rs.Expt1ClustPlot = rs.ClustObj(&rs.Expt1SimMat, "Experiment")
	rs.LbaObjClustPlot = rs.ClustObj(&rs.LbaObjSimMat, "Leabra Obj Sum")
	rs.LbaFullClustPlot = rs.ClustFull(&rs.LbaFullSimMat, rs.LbaFullNames, "Leabra Full")
}

// AvgContrastDist computes average contrast dist over given cat map
func (rs *Res) AvgContrastDist(insm *simat.SimMat, nms []string, catmap map[string]string) float64 {
	no := len(insm.Rows)
	smatv := insm.Mat.(*etensor.Float64).Values
	avgd := 0.0
	for ri := 0; ri < no; ri++ {
		roff := ri * no
		aid := 0.0
		ain := 0
		abd := 0.0
		abn := 0
		rnm := nms[ri]
		rc := catmap[rnm]
		for ci := 0; ci < no; ci++ {
			if ri == ci {
				continue
			}
			cnm := nms[ci]
			cc := catmap[cnm]
			d := smatv[roff+ci]
			if cc == rc {
				aid += d
				ain++
			} else {
				abd += d
				abn++
			}
		}
		if ain > 0 {
			aid /= float64(ain)
		}
		if abn > 0 {
			abd /= float64(abn)
		}
		avgd += aid - abd
	}
	avgd /= float64(no)
	return avgd
}

// PermuteCatTest takes an input sim matrix and tries all one-off permutations relative to given
// initial set of categories, and computes overall average constrast distance for each
// selects categs with lowest dist and iterates until no better permutation can be found
func (rs *Res) PermuteCatTest(insm *simat.SimMat, nms []string, catmap map[string]string, desc string) {
	fmt.Printf("\n#########\n%v\n", desc)
	catm := map[string]int{} // list of categories and index into catnms
	catnms := []string{}
	for _, nm := range nms {
		cat := catmap[nm]
		if _, has := catm[cat]; !has {
			catm[cat] = len(catnms)
			catnms = append(catnms, cat)
		}
	}
	ncats := len(catnms)

	itrmap := make(map[string]string)
	for k, v := range catmap {
		itrmap[k] = v
	}

	std := rs.AvgContrastDist(insm, nms, catmap)
	fmt.Printf("std: %.4f  starting\n", std)

	for itr := 0; itr < 100; itr++ {
		std = rs.AvgContrastDist(insm, nms, itrmap)

		effmap := make(map[string]string)
		mind := 100.0
		mindnm := ""
		mindcat := ""
		for _, nm := range nms { // go over each item
			cat := itrmap[nm]
			for oc := 0; oc < ncats; oc++ { // go over alternative categories
				ocat := catnms[oc]
				if ocat == cat {
					continue
				}
				for k, v := range itrmap {
					if k == nm {
						effmap[k] = ocat // switch
					} else {
						effmap[k] = v
					}
				}
				avgd := rs.AvgContrastDist(insm, nms, effmap)
				if avgd < mind {
					mind = avgd
					mindnm = nm
					mindcat = ocat
				}
				// if avgd < std {
				// 	fmt.Printf("Permute test better than std dist: %v  min dist: %v  for name: %v  in cat: %v\n", std, avgd, nm, ocat)
				// }
			}
		}
		if mind >= std {
			break
		}
		fmt.Printf("itr %v std: %.4f  min: %.4f  name: %v  cat: %v\n", itr, std, mind, mindnm, mindcat)
		itrmap[mindnm] = mindcat // make the switch
	}
	fmt.Printf("std: %.4f  final\n", std)

	for oc := 0; oc < ncats; oc++ {
		cat := catnms[oc]
		fmt.Printf("%v\n", cat)
		for _, nm := range Objs {
			ct := itrmap[nm]
			if ct == cat {
				fmt.Printf("\t%v\n", nm)
			}
		}
	}
}

func (rs *Res) PermuteFitCats() {
	old := false
	nw := true
	_ = nw
	if old {
		rs.PermuteCatTest(&rs.Expt1SimMat, Objs, LbaCats5, "Expt1 LbaCats5")
		rs.PermuteCatTest(&rs.Expt1SimMat, Objs, Expt1Cats5, "Expt1 Expt1Cats5")
		rs.PermuteCatTest(&rs.Expt1SimMat, Objs, LbaCats3, "Expt1 LbaCats3")
		rs.PermuteCatTest(&rs.Expt1SimMat, Objs, Expt1Cats3, "Expt1 Expt1Cats3")
		rs.PermuteCatTest(&rs.Expt1SimMat, Objs, V1Cats, "Expt1 V1Cats") // = -.2928 final
		// cat1 = pyramid, cat2 = everything else, cat3 = banana
	}

	if old {
		rs.PermuteCatTest(&rs.LbaFullSimMat, rs.LbaFullNames, LbaCats5, "Lba LbaCats5")
		// starts -.5071, gets to -.5526
		rs.PermuteCatTest(&rs.LbaFullSimMat, rs.LbaFullNames, LbaCats3, "Lba LbaCats3")
		// is -0.5399 -- looks decent
		rs.PermuteCatTest(&rs.LbaFullSimMat, rs.LbaFullNames, Expt1Cats5, "Lba Expt1Cats5")
		// gets to same -.5526 as LbaCats5
		rs.PermuteCatTest(&rs.LbaFullSimMat, rs.LbaFullNames, Expt1Cats3, "Lba Expt1Cats3")
		rs.PermuteCatTest(&rs.LbaFullSimMat, rs.LbaFullNames, V1Cats, "Lba V1Cats")
		// both of these get to a -.5455 with 1 cat = tablelamp, other 2 = horiz, box
	}

	if old {
		rs.PermuteCatTest(&rs.BpPredFullSimMat, rs.BpPredFullNames, LbaCats5, "BpPred LbaCats5")
		// wow, gets all the way to -0.0838 BpCat!
		rs.PermuteCatTest(&rs.BpPredFullSimMat, rs.BpPredFullNames, BpCats, "BpPred BpCats")
		rs.PermuteCatTest(&rs.BpPredFullSimMat, rs.BpPredFullNames, V1Cats, "BpPred V1Cats")
		// all get to BpCats, nmw
	}
	if old {
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, PredNetCats2, "PredNet PredNetCats2")
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, PredNetCats3, "PredNet PredNetCats3")
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, BpCats, "PredNet BpCats")
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, V1Cats, "PredNet V1Cats")
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, LbaCats5, "PredNet LbaCats5")
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, Expt1Cats5, "PredNet Expt1Cats5")
		rs.PermuteCatTest(&rs.PredNetFullSimMat, rs.PredNetFullNames, Expt1Cats3, "PredNet Expt1Cats3")
	}
	if old {
		rs.PermuteCatTest(&rs.V1FullSimMat, rs.V1FullNames, V1Cats, "V1 V1Cats")
		rs.PermuteCatTest(&rs.V1FullSimMat, rs.V1FullNames, BpCats, "V1 BpCats")
		rs.PermuteCatTest(&rs.V1FullSimMat, rs.V1FullNames, LbaCats5, "V1 LbaCats5")
	}
	if old {
		rs.PermuteCatTest(&rs.PredNetPixelSimMat, rs.PredNetFullNames, BpCats, "PredNet Pixel BpCats")
		rs.PermuteCatTest(&rs.PredNetPixelSimMat, rs.PredNetFullNames, V1Cats, "PredNet Pixel V1Cats")
	}
}

// AvgTickDist computes average within-tick distance (7 ticks in a row)
func (rs *Res) AvgTickDist(insm *simat.SimMat) float64 {
	no := len(insm.Rows)
	smatv := insm.Mat.(*etensor.Float64).Values
	ntick := 7
	avgd := 0.0
	navg := 0
	nobj := no / ntick
	for ri := 0; ri < nobj; ri++ {
		for tri := 0; tri < ntick; tri++ {
			roff := (ri*ntick + tri) * no
			for tci := 0; tci < ntick; tci++ {
				if tri == tci {
					continue
				}
				ci := ri*ntick + tci
				d := smatv[roff+ci]
				avgd += d
				navg++
			}
		}
	}
	avgd /= float64(navg)
	return avgd
}

func (rs *Res) Analyze() {
	rs.OpenSimMats()
	rs.ObjSimMats()
	rs.OpenExptMat()
	rs.TestExptMats()
	rs.Correls()
	rs.ClustPlots()
	rs.PermuteFitCats()
	// atd := rs.AvgTickDist(&rs.LbaTickSimMat)
	// fmt.Printf("avg within-tick distance: %v\n", atd)
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

// ConfigGui configures the GoGi gui interface for this Vis
func (rs *Res) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("results")
	gi.SetAppAbout(`analyze results`)

	win := gi.NewWindow2D("results", "analyze results", width, height, true)
	// vi.Win = win

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tbar := gi.AddNewToolBar(mfr, "tbar")
	tbar.SetStretchMaxWidth()
	// vi.ToolBar = tbar

	split := gi.AddNewSplitView(mfr, "split")
	split.Dim = gi.X
	split.SetStretchMaxWidth()
	split.SetStretchMaxHeight()

	sv := giv.AddNewStructView(split, "sv")
	sv.Viewport = vp
	sv.SetStruct(rs)

	split.SetSplits(1)

	// main menu
	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(win)

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(win)

	gi.SetQuitReqFunc(func() {
		gi.Quit()
	})
	win.SetCloseReqFunc(func(w *gi.Window) {
		gi.Quit()
	})
	win.SetCloseCleanFunc(func(w *gi.Window) {
		go gi.Quit() // once main window is closed, quit
	})

	vp.UpdateEndNoSig(updt)

	win.MainMenuUpdated()
	return win
}

var TheRes Res

func mainrun() {
	TheRes.Init()
	win := TheRes.ConfigGui()
	TheRes.Analyze()
	win.StartEventLoop()
}
