// analyze overall results

package main

import (
	"fmt"
	"log"
	"strings"

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

var Cats = map[string]string{
	"banana":      "pyramid",
	"layercake":   "pyramid",
	"trafficcone": "pyramid",
	"sailboat":    "pyramid",
	"trex":        "pyramid",
	"person":      "vertical",
	"guitar":      "vertical",
	"tablelamp":   "vertical",
	"doorknob":    "round",
	"handgun":     "round",
	"donut":       "round",
	"chair":       "round",
	"slrcamera":   "box",
	"elephant":    "box",
	"piano":       "box",
	"fish":        "box",
	"car":         "horiz",
	"heavycannon": "horiz",
	"stapler":     "horiz",
	"motorcycle":  "horiz",
}

var JustCats []string
var CatsBlanks []string // with blanks

// Res is the main data structure for all expt results and tables
// is visualized in gui so you can click on stuff..
type Res struct {
	LbaFullSimMat    simat.SimMat  `desc:"Leabra TEs full similarity matrix"`
	LbaFullNames     []string      `view:"-" desc:"object names in order for FullSimMat"`
	V1FullSimMat     simat.SimMat  `desc:"V1 full similarity matrix"`
	V1FullNames      []string      `view:"-" desc:"object names in order for FullSimMat"`
	BpPredFullSimMat simat.SimMat  `desc:"WWI Bp Predictive full similarity matrix"`
	BpPredFullNames  []string      `view:"-" desc:"object names in order for FullSimMat"`
	BpEncFullSimMat  simat.SimMat  `desc:"WWI Bp Encoder full similarity matrix"`
	BpEncFullNames   []string      `view:"-" desc:"object names in order for FullSimMat"`
	Expt1SimMat      simat.SimMat  `desc:"Expt1 similarity matrix"`
	LbaObjSimMat     simat.SimMat  `desc:"Leabra TEs obj-cat reduced similarity matrix"`
	V1ObjSimMat      simat.SimMat  `desc:"V1 obj-cat reduced similarity matrix"`
	BpPredObjSimMat  simat.SimMat  `desc:"WWI Bp Predictive obj-cat reduced similarity matrix"`
	BpEncObjSimMat   simat.SimMat  `desc:"WWI Bp Encoder obj-cat reduced similarity matrix"`
	ExptCorrel       etable.Table  `desc:"correlations with expt data for each sim data"`
	ClustPlot        *eplot.Plot2D `desc:"cluster plot"`
}

func (rs *Res) Init() {
	if ObjIdxs == nil {
		no := len(Objs)
		ObjIdxs = make(map[string]int, no)
		JustCats = make([]string, no)
		CatsBlanks = make([]string, no)
		lstcat := ""
		for i, o := range Objs {
			ObjIdxs[o] = i
			cat := Cats[o]
			JustCats[i] = cat
			if cat != lstcat {
				CatsBlanks[i] = cat
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

func (rs *Res) OpenSimMats() {
	rs.OpenFullSimMat(&rs.LbaFullSimMat, &rs.LbaFullNames, "sim_leabra_simat.tsv", "sim_leabra_simat_lbl.tsv", "1.5")
	rs.OpenFullSimMat(&rs.V1FullSimMat, &rs.V1FullNames, "sim_v1_simat.tsv", "sim_v1_simat_lbl.tsv", "1.0")
	rs.OpenFullSimMat(&rs.BpPredFullSimMat, &rs.BpPredFullNames, "sim_bp_pred_simat.tsv", "sim_bp_pred_simat_lbl.tsv", "0.3")
	rs.OpenFullSimMat(&rs.BpEncFullSimMat, &rs.BpEncFullNames, "sim_bp_enc_simat.tsv", "sim_bp_enc_simat_lbl.tsv", "0.04")
}

// ObjSimMat compresses full simat into a much smaller per-object sim mat
func (rs *Res) ObjSimMat(fsm *simat.SimMat, nms []string, osm *simat.SimMat, maxv string) {
	fsmat := fsm.Mat.(*etensor.Float64)

	ono := len(Objs)
	osm.Init()
	osmat := osm.Mat.(*etensor.Float64)
	osmat.SetShape([]int{ono, ono}, nil, nil)
	osm.Rows = CatsBlanks
	osm.Cols = CatsBlanks
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
	sm.Rows = CatsBlanks
	sm.Cols = CatsBlanks
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

func (rs *Res) Analyze() {
	rs.OpenSimMats()
	rs.ObjSimMats()
	rs.OpenExptMat()
	rs.Correls()
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
