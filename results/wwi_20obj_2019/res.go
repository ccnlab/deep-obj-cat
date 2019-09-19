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

var LbaCats = map[string]string{
	"banana":      "1-pyramid",
	"layercake":   "1-pyramid",
	"trafficcone": "1-pyramid",
	"sailboat":    "1-pyramid",
	"trex":        "1-pyramid",
	"person":      "2-vertical",
	"guitar":      "2-vertical",
	"tablelamp":   "2-vertical",
	"doorknob":    "3-round",
	"handgun":     "3-round",
	"donut":       "3-round",
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

var JustLbaCats []string
var LbaCatsBlanks []string // with blanks

var BpCats = map[string]string{
	"tablelamp":   "cat1",
	"person":      "cat1",
	"guitar":      "cat1",
	"trafficcone": "cat1",
	"chair":       "cat1",
	"sailboat":    "cat1",
	"layercake":   "cat1",
	"elephant":    "cat1",
	"piano":       "cat1",
	"donut":       "cat1",
	"doorknob":    "cat1",
	"banana":      "cat1",
	"handgun":     "cat1",
	"slrcamera":   "cat1",
	"trex":        "cat1",
	"car":         "cat2",
	"heavycannon": "cat2",
	"motorcycle":  "cat2",
	"stapler":     "cat2",
	"fish":        "cat2",
}

var PNCats = map[string]string{
	"tablelamp":   "cat1",
	"trafficcone": "cat1",
	"guitar":      "cat1",
	"chair":       "cat1",
	"doorknob":    "cat1",
	"person":      "cat1",
	"sailboat":    "cat1",
	"piano":       "cat2",
	"layercake":   "cat2",
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
}

// Res is the main data structure for all expt results and tables
// is visualized in gui so you can click on stuff..
type Res struct {
	LbaFullSimMat      simat.SimMat  `desc:"Leabra TEs full similarity matrix"`
	LbaFullNames       []string      `view:"-" desc:"object names in order for FullSimMat"`
	LbaLbaCatSimMat    simat.SimMat  `desc:"Leabra TEs full similarity matrix sorted fresh in Lba cat order"`
	LbaV4FullSimMat    simat.SimMat  `desc:"Leabra V4s full similarity matrix"`
	V1FullSimMat       simat.SimMat  `desc:"V1 full similarity matrix"`
	V1FullNames        []string      `view:"-" desc:"object names in order for FullSimMat"`
	BpPredFullSimMat   simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix"`
	BpPredFullNames    []string      `view:"-" desc:"object names in order for FullSimMat"`
	BpPredBpCatSimMat  simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix, in Bp Cat order"`
	BpEncFullSimMat    simat.SimMat  `desc:"WWI Bp Encoder full similarity matrix"`
	BpEncFullNames     []string      `view:"-" desc:"object names in order for FullSimMat"`
	PredNetFullSimMat  simat.SimMat  `desc:"PredNet predictor full similarity matrix"`
	PredNetFullNames   []string      `view:"+" desc:"object names in order for FullSimMat"`
	PredNetPNCatSimMat simat.SimMat  `desc:"PredNet predictor in PN Cat order"`
	Expt1SimMat        simat.SimMat  `desc:"Expt1 similarity matrix"`
	LbaObjSimMat       simat.SimMat  `desc:"Leabra TEs obj-cat reduced similarity matrix"`
	V1ObjSimMat        simat.SimMat  `desc:"V1 obj-cat reduced similarity matrix"`
	BpPredObjSimMat    simat.SimMat  `desc:"WWI Bp Predictive obj-cat reduced similarity matrix"`
	BpEncObjSimMat     simat.SimMat  `desc:"WWI Bp Encoder obj-cat reduced similarity matrix"`
	ExptCorrel         etable.Table  `desc:"correlations with expt data for each sim data"`
	Expt1ClustPlot     *eplot.Plot2D `desc:"cluster plot"`
	LbaObjClustPlot    *eplot.Plot2D `desc:"cluster plot"`
	LbaFullClustPlot   *eplot.Plot2D `desc:"cluster plot"`
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
	for i, v := range smat.Values { // getting correlations here, not 1-correls
		smat.Values[i] = 1 - v
	}
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
func (rs *Res) CatSortSimMat(insm *simat.SimMat, osm *simat.SimMat, nms []string, catmap map[string]string) {
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
		dists[ri] = aid - abd // within - between
	}
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
}

func (rs *Res) OpenSimMats() {
	rs.OpenFullSimMat(&rs.LbaFullSimMat, &rs.LbaFullNames, "sim_leabra_simat.tsv", "sim_leabra_simat_lbl.tsv", "1.5")
	rs.OpenFullSimMat(&rs.LbaV4FullSimMat, &rs.LbaFullNames, "sim_leabra_simat_v4.tsv", "sim_leabra_simat_lbl.tsv", "1.0")
	rs.OpenFullSimMat(&rs.V1FullSimMat, &rs.V1FullNames, "sim_v1_simat.tsv", "sim_v1_simat_lbl.tsv", "1.0")
	rs.OpenFullSimMat(&rs.BpPredFullSimMat, &rs.BpPredFullNames, "sim_bp_pred_simat.tsv", "sim_bp_pred_simat_lbl.tsv", "0.3")
	rs.OpenFullSimMat(&rs.BpEncFullSimMat, &rs.BpEncFullNames, "sim_bp_enc_simat.tsv", "sim_bp_enc_simat_lbl.tsv", "0.04")

	rs.OpenFullSimMatPredNet(&rs.PredNetFullSimMat, &rs.PredNetFullNames, "prednet_layer3.csv", "prednet_labels.csv", "0.15")

	rs.CatSortSimMat(&rs.LbaFullSimMat, &rs.LbaLbaCatSimMat, rs.LbaFullNames, LbaCats)
	rs.CatSortSimMat(&rs.BpPredFullSimMat, &rs.BpPredBpCatSimMat, rs.BpPredFullNames, BpCats)
	rs.CatSortSimMat(&rs.PredNetFullSimMat, &rs.PredNetPNCatSimMat, rs.PredNetFullNames, PNCats)
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

func (rs *Res) Analyze() {
	rs.OpenSimMats()
	rs.ObjSimMats()
	rs.OpenExptMat()
	rs.Correls()
	rs.ClustPlots()
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
