// analyze expt results

package main

import (
	"fmt"

	"github.com/emer/etable/agg"
	"github.com/emer/etable/clust"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/emer/etable/simat"
	"github.com/emer/etable/split"
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

// Expt is the main data structure for all expt results and tables
// is visualized in gui so you can click on stuff..
type Expt struct {
	RawFull    etable.Table  `desc:"full raw data -- not checked in and only used initially"`
	Raw        etable.Table  `desc:"raw relevant data"`
	TrialList  etable.Table  `desc:"trial list -- what were the inputs for each image?"`
	RawStats   *etable.Table `desc:"summary stats on Raw per image"`
	RawPctsAll etable.Table  `desc:"Resp is now 0 = left, 1 = right, result is mean"`
	RawPcts    *etable.Table `desc:"Resp is now 0 = left, 1 = right, result is mean"`
	SimMat     simat.SimMat  `desc:"weights for each pair of objs (NxN) (average proportion choice)"`
	SimMatN    simat.SimMat  `desc:"n's for each pair of objs (NxN) (average proportion choice)"`
	SimMatCats simat.SimMat  `desc:"sim mat with category labels"`
	ClustPlot  *eplot.Plot2D `desc:"cluster plot"`
	SubjStats  *etable.Table `desc:"summary stats on subjects -- quality control"`
}

func (ex *Expt) Init() {
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

func (ex *Expt) OpenRawFull() {
	sch := etable.Schema{
		{"HITId", etensor.STRING, nil, nil},
		{"HITTypeId", etensor.STRING, nil, nil},
		{"Title", etensor.STRING, nil, nil},
		{"Description", etensor.STRING, nil, nil},
		{"Keywords", etensor.STRING, nil, nil},
		{"Reward", etensor.STRING, nil, nil},
		{"CreationTime", etensor.STRING, nil, nil},
		{"MaxAssignments", etensor.INT64, nil, nil},
		{"RequesterAnnotation", etensor.STRING, nil, nil},
		{"AssignmentDurationInSeconds", etensor.INT64, nil, nil},
		{"AutoApprovalDelayInSeconds", etensor.INT64, nil, nil},
		{"Expiration", etensor.STRING, nil, nil},
		{"NumberOfSimilarHITs", etensor.INT64, nil, nil},
		{"LifetimeInSeconds", etensor.STRING, nil, nil},
		{"AssignmentId", etensor.STRING, nil, nil},
		{"WorkerId", etensor.STRING, nil, nil},
		{"AssignmentStatus", etensor.STRING, nil, nil},
		{"AcceptTime", etensor.STRING, nil, nil},
		{"SubmitTime", etensor.STRING, nil, nil},
		{"AutoApprovalTime", etensor.STRING, nil, nil},
		{"ApprovalTime", etensor.STRING, nil, nil},
		{"RejectionTime", etensor.STRING, nil, nil},
		{"RequesterFeedback", etensor.STRING, nil, nil},
		{"WorkTimeInSeconds", etensor.INT64, nil, nil},
		{"LifetimeApprovalRate", etensor.STRING, nil, nil},
		{"Last30DaysApprovalRate", etensor.STRING, nil, nil},
		{"Last7DaysApprovalRate", etensor.STRING, nil, nil},
		{"Input.image_url", etensor.STRING, nil, nil},
		{"Answer.category.label", etensor.STRING, nil, nil},
		{"Approve", etensor.STRING, nil, nil},
		{"Reject", etensor.STRING, nil, nil},
	}

	ex.RawFull.SetFromSchema(sch, 40)
	// note: have to manually delete the header to get this to read.
	err := ex.RawFull.OpenCSV("raw_full_data.csv", etable.Comma)
	if err != nil {
		fmt.Println(err)
	}
}

func (ex *Expt) SaveRawFromFull() {
	sch := etable.Schema{
		{"Image", etensor.STRING, nil, nil},
		{"Subj", etensor.STRING, nil, nil},
		{"Resp", etensor.STRING, nil, nil},
	}
	ex.Raw.SetFromSchema(sch, ex.RawFull.Rows)
	smap := map[string]int{}
	subjNo := 0
	for ri := 0; ri < ex.RawFull.Rows; ri++ {
		wid := ex.RawFull.CellString("WorkerId", ri)
		img := ex.RawFull.CellString("Input.image_url", ri)
		resp := ex.RawFull.CellString("Answer.category.label", ri)
		sno, ok := smap[wid]
		if !ok {
			sno = subjNo
			subjNo++
			smap[wid] = sno
		}
		ex.Raw.SetCellString("Image", ri, img)
		ex.Raw.SetCellString("Subj", ri, fmt.Sprintf("%v", sno))
		ex.Raw.SetCellString("Resp", ri, resp)
	}
	ex.Raw.SaveCSV("raw_data.csv", etable.Comma, true)
}

// raw data extracted from raw full
func (ex *Expt) OpenRaw() {
	err := ex.Raw.OpenCSV("raw_data.csv", etable.Comma)
	if err != nil {
		fmt.Println(err)
	}
}

func (ex *Expt) OpenTrialList() {
	sch := etable.Schema{
		{"Image", etensor.STRING, nil, nil},
		{"L_A", etensor.STRING, nil, nil},
		{"L_B", etensor.STRING, nil, nil},
		{"R_A", etensor.STRING, nil, nil},
		{"R_B", etensor.STRING, nil, nil},
	}
	ex.TrialList.SetFromSchema(sch, 40)
	err := ex.TrialList.OpenCSV("trial-list.csv", etable.Comma)
	if err != nil {
		fmt.Println(err)
	}
}

func (ex *Expt) SumStats() {
	ix := etable.NewIdxView(&ex.Raw)
	byImgResp := split.GroupBy(ix, []string{"Image", "Resp"})
	split.Agg(byImgResp, "Resp", agg.AggCount)
	ex.RawStats = byImgResp.AggsToTable(false) // false = include aggs in col name

	sch := etable.Schema{
		{"Image", etensor.STRING, nil, nil},
		{"Subj", etensor.STRING, nil, nil},
		{"Resp", etensor.FLOAT64, nil, nil},
	}
	ex.RawPctsAll.SetFromSchema(sch, ex.Raw.Rows)
	for ri := 0; ri < ex.Raw.Rows; ri++ {
		resp := ex.Raw.CellString("Resp", ri)
		img := ex.Raw.CellString("Image", ri)
		subj := ex.Raw.CellString("Subj", ri)
		rval := 0.0
		if resp == "Right" {
			rval = 1
		}
		ex.RawPctsAll.SetCellString("Image", ri, img)
		ex.RawPctsAll.SetCellString("Subj", ri, subj)
		ex.RawPctsAll.SetCellFloat("Resp", ri, rval)
	}
	byImg := split.GroupBy(etable.NewIdxView(&ex.RawPctsAll), []string{"Image"})
	split.Agg(byImg, "Resp", agg.AggMean)
	ex.RawPcts = byImg.AggsToTable(true) // true = col names only
}

func (ex *Expt) DoSims() {
	no := len(Objs)
	ex.SimMat.Init()
	smat := ex.SimMat.Mat.(*etensor.Float64)
	smat.SetShape([]int{no, no}, nil, nil)
	ex.SimMat.Rows = Objs
	ex.SimMat.Cols = Objs
	smat.SetZeros()
	smat.SetMetaData("max", "1")
	smat.SetMetaData("min", "0")
	smat.SetMetaData("colormap", "Viridis")
	smat.SetMetaData("grid-fill", "1")
	smat.SetMetaData("dim-extra", "0.15")

	ex.SimMatCats.Mat = smat
	ex.SimMatCats.Rows = CatsBlanks
	ex.SimMatCats.Cols = CatsBlanks

	ex.SimMatN.Init()
	ntsr := ex.SimMatN.Mat.(*etensor.Float64)
	ntsr.SetShape([]int{no, no}, nil, nil)
	ex.SimMatN.Rows = CatsBlanks
	ex.SimMatN.Cols = CatsBlanks
	ntsr.SetZeros()

	tlix := etable.NewIdxView(&ex.TrialList)
	tlix.SortColName("Image", true) // true = ascending
	for ri := 0; ri < ex.RawPcts.Rows; ri++ {
		resp := ex.RawPcts.CellFloat("Resp", ri)
		img := ex.RawPcts.CellString("Image", ri)
		tli := tlix.Idxs[ri]
		if ex.TrialList.CellString("Image", tli) != img {
			panic("error trial list image != raw pcts image!")
		}
		pleft := 1 - resp
		pright := resp
		{
			la := ex.TrialList.CellString("L_A", tli)
			lb := ex.TrialList.CellString("L_B", tli)
			lai := ObjIdxs[la]
			lbi := ObjIdxs[lb]
			cv := smat.Value([]int{lai, lbi})
			cn := ntsr.Value([]int{lai, lbi})
			cn += 1
			cv += pleft
			smat.Set([]int{lai, lbi}, cv)
			smat.Set([]int{lbi, lai}, cv)
			ntsr.Set([]int{lai, lbi}, cn)
			ntsr.Set([]int{lbi, lai}, cn)
		}
		{
			ra := ex.TrialList.CellString("R_A", tli)
			rb := ex.TrialList.CellString("R_B", tli)
			rai := ObjIdxs[ra]
			rbi := ObjIdxs[rb]
			cv := smat.Value([]int{rai, rbi})
			cn := ntsr.Value([]int{rai, rbi})
			cn += 1
			cv += pright
			smat.Set([]int{rai, rbi}, cv)
			smat.Set([]int{rbi, rai}, cv)
			ntsr.Set([]int{rai, rbi}, cn)
			ntsr.Set([]int{rbi, rai}, cn)
		}
	}
	// normalize
	for y := 0; y < no; y++ {
		for x := 0; x < no; x++ {
			cn := ntsr.Value([]int{y, x})
			if cn > 0 {
				cv := smat.Value([]int{y, x})
				cv /= cn
				smat.Set([]int{y, x}, 1-cv) // invert
			} else {
				if y != x {
					smat.Set([]int{y, x}, 0.5) // doh, missing data!
				}
			}
		}
	}
	smat.SetMetaData("precision", "4")
	etensor.SaveCSV(smat, "simat.csv", etable.Comma)
}

func (ex *Expt) Clust() {
	smat := &ex.SimMat
	cl := clust.Glom(smat, clust.ContrastDist) // ContrastDist, MaxDist, Avg all produce similar good fits
	// then plot the results
	pt := &etable.Table{}
	clust.Plot(pt, cl, smat)
	ex.ClustPlot = &eplot.Plot2D{}
	plt := ex.ClustPlot
	plt.InitName(plt, "ClustPlot")
	plt.Params.Title = "Cluster Plot"
	plt.Params.XAxisCol = "X"
	plt.Params.Scale = 3
	plt.SetTable(pt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("X", false, true, 0, false, 0)
	plt.SetColParams("Y", true, true, 0, false, 0)
	plt.SetColParams("Label", true, false, 0, false, 0)
}

func (ex *Expt) Subjs() {
	ix := etable.NewIdxView(&ex.Raw)
	bySubjResp := split.GroupBy(ix, []string{"Subj", "Resp"})
	split.Agg(bySubjResp, "Resp", agg.AggCount)
	ex.SubjStats = bySubjResp.AggsToTable(false) // false = include aggs in col name
}

func (ex *Expt) Analyze() {
	// ex.OpenRawFull()
	// ex.SaveRawFromFull()
	ex.OpenRaw()
	ex.OpenTrialList()
	ex.SumStats()
	ex.DoSims()
	ex.Clust()
	ex.Subjs()
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

// ConfigGui configures the GoGi gui interface for this Vis
func (ex *Expt) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("expt")
	gi.SetAppAbout(`analyze experiment`)

	win := gi.NewWindow2D("expt", "analyze expt", width, height, true)
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
	sv.SetStruct(ex)

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

var TheExpt Expt

func mainrun() {
	TheExpt.Init()
	win := TheExpt.ConfigGui()
	TheExpt.Analyze()
	win.StartEventLoop()
}
