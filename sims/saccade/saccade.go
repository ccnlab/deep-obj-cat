// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
saccade does deep predictive learning on saccadic eye movements
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emer/emergent/actrf"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/env"
	"github.com/emer/emergent/netview"
	"github.com/emer/emergent/params"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/relpos"
	"github.com/emer/empi/empi"
	"github.com/emer/empi/mpi"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/etview" // include to get gui views
	"github.com/emer/etable/metric"
	"github.com/emer/etable/norm"
	"github.com/emer/etable/pca"
	"github.com/emer/etable/split"

	"github.com/emer/axon/axon"
	"github.com/emer/axon/deep"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

func main() {
	TheSim.New() // note: not running Config here -- done in CmdArgs for mpi / nogui
	if len(os.Args) > 1 {
		TheSim.CmdArgs() // simple assumption is that any args = no gui -- could add explicit arg if you want
	} else {
		TheSim.Config()      // for GUI case, config then run..
		gimain.Main(func() { // this starts gui -- requires valid OpenGL display connection (e.g., X11)
			guirun()
		})
	}
}

func guirun() {
	TheSim.Init()
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// see params_def.go for default params..

// Sim encapsulates the entire simulation model, and we define all the
// functionality as methods on this struct.  This structure keeps all relevant
// state information organized and available without having to pass everything around
// as arguments to methods, and provides the core GUI interface (note the view tags
// for the fields which provide hints to how things should be displayed).
type Sim struct {
	Net              *deep.Network   `view:"no-inline" desc:"the network -- click to view / edit parameters for layers, prjns, etc"`
	BinarizeV1       bool            `desc:"if true, V1 inputs are binarized -- todo: test continued need for this"`
	TrnTrlLog        *etable.Table   `view:"no-inline" desc:"training trial-level log data"`
	TrnTrlLogAll     *etable.Table   `view:"no-inline" desc:"all training trial-level log data (aggregated from MPI)"`
	TrnTrlRepLog     *etable.Table   `view:"no-inline" desc:"training trial-level reps log data"`
	TrnTrlRepLogAll  *etable.Table   `view:"no-inline" desc:"training trial-level reps log data"`
	CatLayActs       *etable.Table   `view:"no-inline" desc:"super layer activations per category / object"`
	CatLayActsDest   *etable.Table   `view:"no-inline" desc:"MPI dest super layer activations per category / object"`
	TrnEpcLog        *etable.Table   `view:"no-inline" desc:"training epoch-level log data"`
	TstEpcLog        *etable.Table   `view:"no-inline" desc:"testing epoch-level log data"`
	TstTrlLog        *etable.Table   `view:"no-inline" desc:"testing trial-level log data"`
	TstTrlLogAll     *etable.Table   `view:"no-inline" desc:"all testing trial-level log data (aggregated from MPI)"`
	ActRFs           actrf.RFs       `view:"no-inline" desc:"activation-based receptive fields"`
	RunLog           *etable.Table   `view:"no-inline" desc:"summary log of each run"`
	RunStats         *etable.Table   `view:"no-inline" desc:"aggregate stats on all runs"`
	MinusCycles      int             `desc:"number of minus-phase cycles"`
	PlusCycles       int             `desc:"number of plus-phase cycles"`
	SubPools         bool            `desc:"if true, organize layers and connectivity with 2x2 sub-pools within each topological pool"`
	ErrLrMod         axon.LrateMod   `view:"inline" desc:"learning rate modulation as function of error"`
	Params           params.Sets     `view:"no-inline" desc:"full collection of param sets"`
	ParamSet         string          `desc:"which set of *additional* parameters to use -- always applies Base and optionaly this next if set"`
	Tag              string          `desc:"extra tag string to add to any file names output from sim (e.g., weights files, log files, params for run)"`
	Prjn4x4Skp2      *prjn.PoolTile  `view:"Standard feedforward topographic projection, recv = 1/2 send size"`
	Prjn4x4Skp2Recip *prjn.PoolTile  `view:"Reciprocal"`
	Prjn2x2Skp2      *prjn.PoolTile  `view:"sparser skip 2 -- no overlap"`
	Prjn2x2Skp2Recip *prjn.PoolTile  `view:"Reciprocal"`
	Prjn3x3Skp1      *prjn.PoolTile  `view:"Standard same-to-same size topographic projection"`
	PrjnSigTopo      *prjn.PoolTile  `view:"sigmoidal topographic projection used in LIP saccade remapping layers"`
	PrjnGaussTopo    *prjn.PoolTile  `view:"gaussian topographic projection used in LIP saccade remapping layers"`
	StartRun         int             `desc:"starting run number -- typically 0 but can be set in command args for parallel runs on a cluster"`
	MaxRuns          int             `desc:"maximum number of model runs to perform (starting from StartRun)"`
	MaxEpcs          int             `desc:"maximum number of epochs to run per model run"`
	MaxTrls          int             `desc:"maximum number of training trials per epoch (each trial is MaxTicks ticks)"`
	MaxTicks         int             `desc:"max number of ticks, for logs, stats"`
	NZeroStop        int             `desc:"if a positive number, training will stop after this many epochs with zero SSE"`
	RepsInterval     int             `desc:"how often to analyze the representations"`
	TrainEnv         SacEnv          `desc:"Training environment -- 3D Object training"`
	TestEnv          SacEnv          `desc:"Testing environment -- testing 3D Objects"`
	Time             axon.Time       `desc:"axon timing parameters and state"`
	ViewOn           bool            `desc:"whether to update the network view while running"`
	TrainUpdt        axon.TimeScales `desc:"at what time scale to update the display during training?  Anything longer than Epoch updates at Epoch in this model"`
	TestUpdt         axon.TimeScales `desc:"at what time scale to update the display during testing?  Anything longer than Epoch updates at Epoch in this model"`
	LayStatNms       []string        `desc:"names of layers to collect more detailed stats on (avg act, etc)"`
	ActRFNms         []string        `desc:"names of layers to compute activation rfields on"`
	HidTrlCosDiff    []float64       `view:"-" desc:"trial-level cosine differnces"`

	// statistics: note use float64 as that is best for etable.Table
	PulvLays       []string  `view:"-" desc:"pulvinar layers -- for stats"`
	HidLays        []string  `view:"-" desc:"hidden layers: super and CT -- for hogging stats"`
	SuperLays      []string  `view:"-" desc:"superficial layers"`
	InLays         []string  `view:"-" desc:"input layers -- for stats"`
	PulvCosDiff    []float64 `inactive:"+" desc:"trial stats cos diff for pulvs"`
	PulvUnitErr    []float64 `inactive:"+" desc:"trial stats UnitErr for pulvs"`
	PulvTrlCosDiff []float64 `inactive:"+" desc:"trial stats trial cos diff for pulvs"`
	TrlCosDiff     float64   `inactive:"+" desc:"cos diff used for driving ErrLrMod"`
	EpcPerTrlMSec  float64   `inactive:"+" desc:"how long did the epoch take per trial in wall-clock milliseconds"`
	LastTrlMSec    float64   `inactive:"+" desc:"how long did the epoch take to run last trial in wall-clock milliseconds"`
	PCA            pca.PCA   `view:"-" desc:"pca obj"`

	// internal state - view:"-"
	Win          *gi.Window                    `view:"-" desc:"main GUI window"`
	NetView      *netview.NetView              `view:"-" desc:"the network viewer"`
	ToolBar      *gi.ToolBar                   `view:"-" desc:"the master toolbar"`
	CurImgGrid   *etview.TensorGrid            `view:"-" desc:"the current image grid view"`
	ActRFGrids   map[string]*etview.TensorGrid `view:"-" desc:"the act rf grid views"`
	TrnTrlPlot   *eplot.Plot2D                 `view:"-" desc:"the training trial plot"`
	TrnEpcPlot   *eplot.Plot2D                 `view:"-" desc:"the training epoch plot"`
	TstEpcPlot   *eplot.Plot2D                 `view:"-" desc:"the testing epoch plot"`
	TstTrlPlot   *eplot.Plot2D                 `view:"-" desc:"the test-trial plot"`
	RunPlot      *eplot.Plot2D                 `view:"-" desc:"the run plot"`
	SacTableView *etview.TableView             `view:"-" desc:"the saccade table view"`
	TrnEpcFile   *os.File                      `view:"-" desc:"log file"`
	TrnTrlFile   *os.File                      `view:"-" desc:"log file"`
	RunFile      *os.File                      `view:"-" desc:"log file"`
	ValsTsrs     map[string]*etensor.Float32   `view:"-" desc:"for holding layer values"`
	SaveWts      bool                          `view:"-" desc:"for command-line run only, auto-save final weights after each run"`
	NoGui        bool                          `view:"-" desc:"if true, runing in no GUI mode"`
	LogSetParams bool                          `view:"-" desc:"if true, print message for all params that are set"`
	IsRunning    bool                          `view:"-" desc:"true if sim is running"`
	StopNow      bool                          `view:"-" desc:"flag to stop running"`
	NeedsNewRun  bool                          `view:"-" desc:"flag to initialize NewRun if last one finished"`
	RndSeeds     []int64                       `view:"-" desc:"the current random seeds to use for each run"`
	LastEpcTime  time.Time                     `view:"-" desc:"timer for last epoch"`
	LastTrlTime  time.Time                     `view:"-" desc:"timer for last trial"`

	UseMPI      bool      `view:"-" desc:"if true, use MPI to distribute computation across nodes"`
	SaveProcLog bool      `view:"-" desc:"if true, save logs per processor"`
	Comm        *mpi.Comm `view:"-" desc:"mpi communicator"`
	AllDWts     []float32 `view:"-" desc:"buffer of all dwt weight changes -- for mpi sharing"`
	SumDWts     []float32 `view:"-" desc:"buffer of MPI summed dwt weight changes"`
}

// this registers this Sim Type and gives it properties that e.g.,
// prompt for filename for save methods.
var KiT_Sim = kit.Types.AddType(&Sim{}, SimProps)

// TheSim is the overall state for this simulation
var TheSim Sim

// New creates new blank elements and initializes defaults
func (ss *Sim) New() {
	ss.Net = &deep.Network{}
	ss.TrnTrlLog = &etable.Table{}
	ss.TrnTrlLogAll = &etable.Table{}
	ss.TrnTrlRepLog = &etable.Table{}
	ss.TrnTrlRepLogAll = &etable.Table{}
	ss.CatLayActs = &etable.Table{}
	ss.CatLayActsDest = &etable.Table{}
	ss.TrnEpcLog = &etable.Table{}
	ss.TstEpcLog = &etable.Table{}
	ss.TstTrlLog = &etable.Table{}
	ss.TstTrlLogAll = &etable.Table{}
	ss.RunLog = &etable.Table{}
	ss.RunStats = &etable.Table{}

	ss.Time.Defaults()
	ss.MinusCycles = 150
	ss.PlusCycles = 50
	ss.SubPools = false
	ss.RepsInterval = 10
	ss.ErrLrMod.Defaults()
	ss.ErrLrMod.Base = 0.2 // 0.2 > 0.1
	ss.ErrLrMod.Range.Set(0.2, 0.8)

	ss.Params = ParamSets
	ss.RndSeeds = make([]int64, 100) // make enough for plenty of runs
	for i := 0; i < 100; i++ {
		ss.RndSeeds[i] = int64(i) + 1 // exclude 0
	}
	ss.ViewOn = true
	ss.TrainUpdt = axon.Phase
	ss.TestUpdt = axon.Phase
	ss.LayStatNms = []string{"LIPP"}
	ss.ActRFNms = []string{"V4:Image", "V4:Output", "IT:Image", "IT:Output"}
	ss.Defaults()
}

// Defaults sets default values for params / prjns
func (ss *Sim) Defaults() {
	ss.Prjn4x4Skp2 = prjn.NewPoolTile()
	ss.Prjn4x4Skp2.Size.Set(4, 4)
	ss.Prjn4x4Skp2.Skip.Set(2, 2)
	ss.Prjn4x4Skp2.Start.Set(-1, -1)
	ss.Prjn4x4Skp2.TopoRange.Min = 0.8
	// but using a symmetric scale range .8 - 1.2 seems like it might be good -- otherwise
	// weights are systematicaly smaller.
	// note: gauss defaults on
	// ss.Prjn4x4Skp2.GaussFull.DefNoWrap()
	// ss.Prjn4x4Skp2.GaussInPool.DefNoWrap()

	ss.Prjn4x4Skp2Recip = prjn.NewPoolTile()
	ss.Prjn4x4Skp2Recip.Size.Set(4, 4)
	ss.Prjn4x4Skp2Recip.Skip.Set(2, 2)
	ss.Prjn4x4Skp2Recip.Start.Set(-1, -1)
	ss.Prjn4x4Skp2Recip.TopoRange.Min = 0.8 // note: none of these make a very big diff
	ss.Prjn4x4Skp2Recip.Recip = true

	ss.Prjn2x2Skp2 = prjn.NewPoolTile()
	ss.Prjn2x2Skp2.Size.Set(2, 2)
	ss.Prjn2x2Skp2.Skip.Set(2, 2)
	ss.Prjn2x2Skp2.Start.Set(0, 0)
	ss.Prjn2x2Skp2.TopoRange.Min = 0.8

	ss.Prjn2x2Skp2Recip = prjn.NewPoolTile()
	ss.Prjn2x2Skp2Recip.Size.Set(2, 2)
	ss.Prjn2x2Skp2Recip.Skip.Set(2, 2)
	ss.Prjn2x2Skp2Recip.Start.Set(0, 0)
	ss.Prjn2x2Skp2Recip.TopoRange.Min = 0.8
	ss.Prjn2x2Skp2Recip.Recip = true

	ss.Prjn3x3Skp1 = prjn.NewPoolTile()
	ss.Prjn3x3Skp1.Size.Set(3, 3)
	ss.Prjn3x3Skp1.Skip.Set(1, 1)
	ss.Prjn3x3Skp1.Start.Set(-1, -1)
	ss.Prjn3x3Skp1.TopoRange.Min = 0.8 // note: none of these make a very big diff

	ss.PrjnSigTopo = prjn.NewPoolTile()
	ss.PrjnSigTopo.GaussOff()
	ss.PrjnSigTopo.Size.Set(1, 1)
	ss.PrjnSigTopo.Skip.Set(0, 0)
	ss.PrjnSigTopo.Start.Set(0, 0)
	ss.PrjnSigTopo.TopoRange.Set(0.25, 0.75)
	ss.PrjnSigTopo.SigFull.On = true
	ss.PrjnSigTopo.SigFull.Gain = 0.05
	ss.PrjnSigTopo.SigFull.CtrMove = 0.5

	ss.PrjnGaussTopo = prjn.NewPoolTile()
	ss.PrjnGaussTopo.Size.Set(1, 1)
	ss.PrjnGaussTopo.Skip.Set(0, 0)
	ss.PrjnGaussTopo.Start.Set(0, 0)
	ss.PrjnGaussTopo.TopoRange.Set(0.25, 0.75)
	ss.PrjnGaussTopo.GaussInPool.On = false // Full only
	ss.PrjnGaussTopo.GaussFull.Sigma = 0.6
	ss.PrjnGaussTopo.GaussFull.Wrap = true
	ss.PrjnGaussTopo.GaussFull.CtrMove = 1
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Configs

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.ConfigEnv()
	ss.ConfigNet(ss.Net)
	ss.InitStats()
	ss.ConfigTrnTrlLog(ss.TrnTrlLog)
	ss.ConfigTrnTrlLog(ss.TrnTrlLogAll)
	ss.ConfigTrnTrlRepLog(ss.TrnTrlRepLog)
	ss.ConfigTrnTrlRepLog(ss.TrnTrlRepLogAll)
	ss.ConfigTrnEpcLog(ss.TrnEpcLog)
	ss.ConfigTstEpcLog(ss.TstEpcLog)
	ss.ConfigTstTrlLog(ss.TstTrlLog)
	ss.ConfigTstTrlLog(ss.TstTrlLogAll)
	ss.ConfigRunLog(ss.RunLog)
}

func (ss *Sim) ConfigEnv() {
	if ss.MaxRuns == 0 { // allow user override
		ss.MaxRuns = 1
	}
	if ss.MaxEpcs == 0 { // allow user override
		ss.MaxEpcs = 50
		ss.NZeroStop = -1
	}
	if ss.MaxTrls == 0 { // allow user override
		ss.MaxTrls = 64
		ss.MaxTicks = 8
	}

	ss.TrainEnv.Nm = "TrainEnv"
	ss.TrainEnv.Dsc = "training params and state"
	ss.TrainEnv.Defaults()
	ss.TrainEnv.Run.Max = ss.MaxRuns // note: we are not setting epoch max -- do that manually
	ss.TrainEnv.Trial.Max = ss.MaxTrls

	ss.TestEnv.Nm = "TestEnv"
	ss.TestEnv.Dsc = "testing params and state"
	ss.TestEnv.Defaults()
	ss.TestEnv.Trial.Max = 500

	ss.TrainEnv.Init(0)
	ss.TestEnv.Init(0)
	ss.TrainEnv.Validate()
	ss.TestEnv.Validate()
}

func (ss *Sim) ConfigNet(net *deep.Network) {
	net.InitName(net, "SacNet")

	vsz := ss.TrainEnv.VisSize
	asz := ss.TrainEnv.AngSize
	dsz := ss.TrainEnv.DistSize

	v1 := net.AddLayer4D("V1", vsz, vsz, 1, 1, emer.Input)
	s1e := net.AddLayer4D("S1e", dsz, asz, 1, 1, emer.Input)
	scs := net.AddLayer4D("SCs", dsz, asz, 1, 1, emer.Input)
	scd := net.AddLayer4D("SCd", dsz, asz, 1, 1, emer.Input)
	md := net.AddLayer4D("MDe", dsz, asz, 1, 1, emer.Target)

	lip, lipct, lipp := net.AddDeep4D("LIP", vsz, vsz, 2, 2)
	lipp.Shape().SetShape([]int{vsz, vsz, 2, 1}, nil, nil)
	lipp.(*deep.TRCLayer).Drivers.Add("V1", "S1e")

	fef := net.AddLayer4D("FEF", dsz, asz, 2, 2, emer.Hidden)

	sef, sefct, sefp := net.AddDeep4D("SEF", dsz, asz, 2, 2)
	sefp.Shape().SetShape([]int{dsz, asz, 2, 2}, nil, nil)
	sefp.(*deep.TRCLayer).Drivers.Add("FEF")

	v1.SetClass("V1")
	s1e.SetClass("V1")
	scs.SetClass("PopIn")
	scd.SetClass("PopIn")
	md.SetClass("PopIn")
	fef.SetClass("PopIn")

	lip.SetClass("LIP")
	lipct.SetClass("LIP")
	lipp.SetClass("LIP")

	sef.SetClass("SEF")
	sefct.SetClass("SEF")
	sefp.SetClass("SEF")

	full := prjn.NewFull()
	pone2one := prjn.NewPoolOneToOne()

	net.ConnectLayers(v1, lip, ss.Prjn3x3Skp1, emer.Forward) // .SetClass("Fixed") // has .5 wtscale in Params
	net.ConnectLayers(s1e, lip, full, emer.Forward)
	net.ConnectCtxtToCT(lipct, lipct, full).SetClass("CTSelfLIP")

	net.ConnectLayers(s1e, lipct, full, emer.Forward)
	net.BidirConnectLayers(md, lip, full)
	net.BidirConnectLayers(sef, lip, full)

	lipp.RecvPrjns().SendName("LIPCT").SetPattern(full) // full > pone2one
	lip.RecvPrjns().SendName("LIPP").SetClass("FmPulv FmLIP")
	lipct.RecvPrjns().SendName("LIPP").SetClass("FmPulv FmLIP")
	lipct.RecvPrjns().SendName("LIP").SetClass("CTCtxtStd")

	// InitWts optionally sets ss.PrjnSigTopo
	// net.ConnectLayers(sc, lip, full, emer.Forward)

	// lipct.RecvPrjns().SendName("LIP").SetPattern(ss.Prjn3x3Skp1)

	net.BidirConnectLayers(fef, md, ss.Prjn3x3Skp1)
	net.BidirConnectLayers(lip, fef, full)

	net.BidirConnectLayers(fef, sef, ss.Prjn3x3Skp1)

	net.LateralConnectLayerPrjn(lip, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(lipct, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)

	//	Position

	s1e.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v1.Name(), YAlign: relpos.Front, Space: 2})
	scd.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: s1e.Name(), XAlign: relpos.Left, Space: 10})
	scs.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: scd.Name(), XAlign: relpos.Left, Space: 10})
	md.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: s1e.Name(), YAlign: relpos.Front, Space: 2})
	fef.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: md.Name(), XAlign: relpos.Left, Space: 10})

	lip.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: v1.Name(), XAlign: relpos.Left, YAlign: relpos.Front})
	lipct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: lip.Name(), XAlign: relpos.Left, Space: 10})
	lipp.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: lipct.Name(), XAlign: relpos.Left, Space: 10})

	sef.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: lip.Name(), YAlign: relpos.Front, Space: 2})
	sefct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: sef.Name(), XAlign: relpos.Left, Space: 10})
	sefp.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: sefct.Name(), XAlign: relpos.Left, Space: 10})

	net.Defaults()
	ss.SetParams("Network", false) // only set Network params
	err := net.Build()
	if err != nil {
		log.Println(err)
		return
	}

	if !ss.NoGui {
		sr := net.SizeReport()
		mpi.Printf("%s", sr)
	}

	//	ar := net.ThreadAlloc(4) // must be done after build
	ar := net.ThreadReport() // hand tuning now..
	mpi.Printf("%s", ar)

	// ss.InitWts(net) // too slow
}

func (ss *Sim) SetTopoSWts(net *deep.Network, send, recv string, pooltile *prjn.PoolTile) {
	slay := net.LayerByName(send)
	rlay := net.LayerByName(recv)
	pj := rlay.RecvPrjns().SendName(send).(axon.AxonPrjn).AsAxon()
	scales := &etensor.Float32{}
	pooltile.TopoWts(slay.Shape(), rlay.Shape(), scales)
	pj.SetSWtsRPool(scales)
}

func (ss *Sim) InitWts(net *deep.Network) {
	if len(net.Layers) == 0 {
		return
	}
	net.InitWts()
	return
	net.InitTopoSWts() //  sets all wt scales

	// these are not set automatically b/c prjn is Full, not PoolTile
	ss.SetTopoSWts(net, "EyePos", "LIP", ss.PrjnGaussTopo)
	ss.SetTopoSWts(net, "SC", "LIP", ss.PrjnSigTopo)
	ss.SetTopoSWts(net, "ObjVel", "LIP", ss.PrjnSigTopo)

	ss.SetTopoSWts(net, "EyePos", "LIPCT", ss.PrjnGaussTopo)
	ss.SetTopoSWts(net, "Saccade", "LIPCT", ss.PrjnSigTopo)
	// ss.SetTopoSWts(net, "SC", "LIPCT", ss.PrjnSigTopo)
	ss.SetTopoSWts(net, "ObjVel", "LIPCT", ss.PrjnSigTopo)
}

////////////////////////////////////////////////////////////////////////////////
// 	    Init, utils

// Init restarts the run, and initializes everything, including network weights
// and resets the epoch log table
func (ss *Sim) Init() {
	ss.InitRndSeed()
	ss.StopNow = false
	ss.SetParams("", false) // all sheets
	ss.NewRun()
	ss.UpdateView(true)
}

// NewRndSeed gets a new random seed based on current time -- otherwise uses
// the same random seed for every run
// InitRndSeed initializes the random seed based on current training run number
func (ss *Sim) InitRndSeed() {
	run := ss.TrainEnv.Run.Cur
	rand.Seed(ss.RndSeeds[run])
}

// NewRndSeed gets a new set of random seeds based on current time -- otherwise uses
// the same random seeds for every run
func (ss *Sim) NewRndSeed() {
	rs := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		ss.RndSeeds[i] = rs + int64(i)
	}
}

// Counters returns a string of the current counter state
// use tabs to achieve a reasonable formatting overall
// and add a few tabs at the end to allow for expansion..
func (ss *Sim) Counters(train bool) string {
	if train {
		return fmt.Sprintf("Run:\t%d\tEpoch:\t%d\tTrial:\t%d\tTick:\t%d\tCycle:\t%d\tName:\t%v\t\t\t", ss.TrainEnv.Run.Cur, ss.TrainEnv.Epoch.Cur, ss.TrainEnv.Trial.Cur, ss.TrainEnv.Tick.Cur, ss.Time.Cycle, ss.TrainEnv.String())
	} else {
		return fmt.Sprintf("Run:\t%d\tEpoch:\t%d\tTrial:\t%d\tCycle:\t%d\tName:\t%v\t\t\t", ss.TrainEnv.Run.Cur, ss.TrainEnv.Epoch.Cur, ss.TestEnv.Trial.Cur, ss.Time.Cycle, ss.TestEnv.String())
	}
}

func (ss *Sim) UpdateView(train bool) {
	if ss.NetView != nil && ss.NetView.IsVisible() {
		ss.NetView.Record(ss.Counters(train))
		// note: essential to use Go version of update when called from another goroutine
		ss.NetView.GoUpdate() // note: using counters is significantly slower..
	}
}

func (ss *Sim) UpdateViewTime(train bool, viewUpdt axon.TimeScales) {
	switch viewUpdt {
	case axon.Cycle:
		ss.UpdateView(train)
	case axon.FastSpike:
		if ss.Time.Cycle%10 == 0 {
			ss.UpdateView(train)
		}
	case axon.GammaCycle:
		if ss.Time.Cycle%25 == 0 {
			ss.UpdateView(train)
		}
	case axon.AlphaCycle:
		if ss.Time.Cycle%100 == 0 {
			ss.UpdateView(train)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// 	    Running the Network, starting bottom-up..

// ThetaCyc runs one theta cycle (200 msec) of processing.
// External inputs must have already been applied prior to calling,
// using ApplyExt method on relevant layers (see TrainTrial, TestTrial).
// If train is true, then learning DWt or WtFmDWt calls are made.
// Handles netview updating within scope, and calls TrainStats()
func (ss *Sim) ThetaCyc(train bool) {
	// ss.Win.PollEvents() // this can be used instead of running in a separate goroutine
	viewUpdt := ss.TrainUpdt
	if !train {
		viewUpdt = ss.TestUpdt
	}

	// update prior weight changes at start, so any DWt values remain visible at end
	// you might want to do this less frequently to achieve a mini-batch update
	// in which case, move it out to the TrainTrial method where the relevant
	// counters are being dealt with.
	if train {
		ss.MPIWtFmDWt()
	}

	minusCyc := ss.MinusCycles
	plusCyc := ss.PlusCycles

	ss.Net.NewState()
	ss.Time.NewState()
	for cyc := 0; cyc < minusCyc; cyc++ { // do the minus phase
		ss.Net.Cycle(&ss.Time)
		// ss.LogTrnCyc(ss.TrnCycLog, ss.Time.Cycle)
		// if !ss.NoGui {
		// 	ss.RecSpikes(ss.Time.Cycle)
		// }
		ss.Time.CycleInc()
		switch ss.Time.Cycle { // save states at beta-frequency -- not used computationally
		case 75:
			ss.Net.ActSt1(&ss.Time)
			// if erand.BoolProb(float64(ss.PAlphaPlus), -1) {
			// 	ss.Net.TargToExt()
			// 	ss.Time.PlusPhase = true
			// }
		case 100:
			ss.Net.ActSt2(&ss.Time)
			ss.Net.ClearTargExt()
			ss.Time.PlusPhase = false
		}

		if cyc == minusCyc-1 { // do before view update
			if train {
				ss.DoAction(&ss.TrainEnv)
			} else {
				ss.DoAction(&ss.TestEnv)
			}
			ss.Net.MinusPhase(&ss.Time)
		}
		if ss.ViewOn {
			ss.UpdateViewTime(train, viewUpdt)
		}
	}
	ss.Time.NewPhase()
	if viewUpdt == axon.Phase {
		ss.UpdateView(train)
	}
	for cyc := 0; cyc < plusCyc; cyc++ { // do the plus phase
		ss.Net.Cycle(&ss.Time)
		// ss.LogTrnCyc(ss.TrnCycLog, ss.Time.Cycle)
		// if !ss.NoGui {
		// 	ss.RecSpikes(ss.Time.Cycle)
		// }
		ss.Time.CycleInc()

		if cyc == plusCyc-1 { // do before view update
			ss.Net.PlusPhase(&ss.Time)
			ss.Net.CTCtxt(&ss.Time) // update context at end
		}
		if ss.ViewOn {
			ss.UpdateViewTime(train, viewUpdt)
		}
	}

	ss.TrialStats(train)

	if train && ss.TrainEnv.Tick.Cur > 0 { // important: don't learn on first tick!
		ss.ErrLrMod.LrateMod(ss.Net.AsAxon(), float32(1-ss.TrlCosDiff))
		ss.Net.DWt()
	}

	if viewUpdt == axon.Phase || viewUpdt == axon.AlphaCycle || viewUpdt == axon.ThetaCycle {
		ss.UpdateView(train)
	}
}

// ApplyInputs applies input patterns from given envirbonment.
// It is good practice to have this be a separate method with appropriate
// args so that it can be used for various different contexts
// (training, testing, etc).
func (ss *Sim) ApplyInputs(en env.Env) {
	ss.Net.InitExt() // clear any existing inputs -- not strictly necessary if always
	// going to the same layers, but good practice and cheap anyway

	lays := []string{"V1", "S1e", "SCs", "SCd"}
	for _, lnm := range lays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		pats := en.State(ly.Nm)
		if pats != nil {
			ly.ApplyExt(pats)
		}
	}
}

func (ss *Sim) DoAction(en env.Env) {
	tick, _, _ := en.Counter(env.Tick)
	if tick != 0 {
		return
	}
	mdacts := ss.ValsTsr("MDe")
	md := ss.Net.LayerByName("MDe").(axon.AxonLayer).AsAxon()
	md.UnitValsTensor(mdacts, "Act")
	en.Action("MD", mdacts)
	scd := ss.Net.LayerByName("SCd").(axon.AxonLayer).AsAxon()
	pats := en.State(scd.Nm)
	if pats != nil {
		scd.ApplyExt(pats)
		md.ApplyExt(pats)
	}
}

// todo: action, plus phase apply SCd

// TrainTrial runs one trial of training using TrainEnv
func (ss *Sim) TrainTrial() {

	if ss.NeedsNewRun {
		ss.NewRun()
	}

	ss.TrainEnv.Step() // the Env encapsulates and manages all counter state
	if ss.SacTableView != nil {
		ss.SacTableView.UpdateTable()
	}

	// Key to query counters FIRST because current state is in NEXT epoch
	// if epoch counter has changed
	epc, _, chg := ss.TrainEnv.Counter(env.Epoch)
	if chg {
		ss.LogTrnEpc(ss.TrnEpcLog)
		ss.EpochSched(epc)
		if ss.ViewOn && ss.TrainUpdt > axon.AlphaCycle {
			ss.UpdateView(true)
		}
		if epc >= ss.MaxEpcs {
			// done with training..
			ss.RunEnd()
			if ss.TrainEnv.Run.Incr() { // we are done!
				ss.StopNow = true
				return
			} else {
				ss.NeedsNewRun = true
				return
			}
		}
	}

	// note: type must be in place before apply inputs
	ss.ApplyInputs(&ss.TrainEnv)
	ss.ThetaCyc(true) // train
	ss.LogTrnTrl(ss.TrnTrlLog)
	// if ss.RepsInterval > 0 && epc%ss.RepsInterval == 0 {
	// 	ss.LogTrnRepTrl(ss.TrnTrlRepLog)
	// }
	if ss.CurImgGrid != nil {
		ss.CurImgGrid.UpdateSig()
	}
}

// RunEnd is called at the end of a run -- save weights, record final log, etc here
func (ss *Sim) RunEnd() {
	ss.LogRun(ss.RunLog)
	if ss.SaveWts {
		fnm := ss.WeightsFileName()
		fmt.Printf("Saving Weights to: %v\n", fnm)
		ss.Net.SaveWtsJSON(gi.FileName(fnm))
		fmt.Printf("Weights saved..\n")
	}
}

// NewRun intializes a new run of the model, using the TrainEnv.Run counter
// for the new run value
func (ss *Sim) NewRun() {
	ss.InitRndSeed()
	run := ss.TrainEnv.Run.Cur
	ss.TrainEnv.Init(run)
	ss.TestEnv.Init(run)
	ss.Time.Reset()
	ss.InitWts(ss.Net)
	ss.InitStats()
	ss.TrnEpcLog.SetNumRows(0)
	ss.TrnTrlLog.SetNumRows(0)
	ss.TstEpcLog.SetNumRows(0)
	ss.TstTrlLog.SetNumRows(0)
	ss.NeedsNewRun = false
}

// InitStats initializes all the statistics, especially important for the
// cumulative epoch stats -- called at start of new run
func (ss *Sim) InitStats() {
	if len(ss.PulvLays) > 0 {
		return
	}
	ss.PulvLays = []string{}
	ss.HidLays = []string{}
	ss.InLays = []string{}
	ss.SuperLays = []string{"V1"}
	net := ss.Net
	for _, ly := range net.Layers {
		if ly.IsOff() {
			continue
		}
		if ly.Name() == "MTPos" {
			continue
		}
		switch ly.Type() {
		case emer.Input:
			ss.InLays = append(ss.InLays, ly.Name())
		case deep.TRC:
			ss.PulvLays = append(ss.PulvLays, ly.Name())
		case emer.Hidden:
			ss.SuperLays = append(ss.SuperLays, ly.Name())
			fallthrough
		case deep.CT:
			ss.HidLays = append(ss.HidLays, ly.Name())
		}
	}
	np := len(ss.PulvLays)
	ss.PulvCosDiff = make([]float64, np)
	ss.PulvUnitErr = make([]float64, np)
	ss.PulvTrlCosDiff = make([]float64, np)

	nh := len(ss.HidLays)
	ss.HidTrlCosDiff = make([]float64, nh)
}

// TrialStats computes the trial-level statistics.
func (ss *Sim) TrialStats(train bool) {
	for li, lnm := range ss.PulvLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		ss.PulvCosDiff[li] = float64(ly.CosDiff.Cos)
		ss.PulvUnitErr[li] = ly.PctUnitErr()
		if lnm == "LIPP" {
			ss.TrlCosDiff = float64(ly.CosDiff.Cos)
		}
	}
	ss.TrialCosDiff()
}

func (ss *Sim) TrialCosDiffLay(lnm string, varnm string) float64 {
	vtp := ss.ValsTsr(lnm + "TrlCosDiffP")
	vtc := ss.ValsTsr(lnm + "TrlCosDiffC")
	ly := ss.Net.LayerByName(lnm)
	ly.UnitValsTensor(vtc, varnm)
	cosdif := 0.0
	if len(vtp.Values) == len(vtc.Values) {
		cosdif = float64(metric.Correlation32(vtp.Values, vtc.Values))
	} else {
		vtp.CopyShapeFrom(vtc)
	}
	copy(vtp.Values, vtc.Values)
	return cosdif
}

func (ss *Sim) TrialCosDiff() {
	for li, lnm := range ss.HidLays {
		ss.HidTrlCosDiff[li] = ss.TrialCosDiffLay(lnm, "ActM")
	}
	for li, lnm := range ss.PulvLays {
		ss.PulvTrlCosDiff[li] = ss.TrialCosDiffLay(lnm, "ActP") // driver more interesting here
	}
}

// TrainEpoch runs training trials for remainder of this epoch
func (ss *Sim) TrainEpoch() {
	ss.StopNow = false
	curEpc := ss.TrainEnv.Epoch.Cur
	for {
		ss.TrainTrial()
		if ss.StopNow || ss.TrainEnv.Epoch.Cur != curEpc {
			break
		}
	}
	ss.Stopped()
}

// TrainRun runs training trials for remainder of run
func (ss *Sim) TrainRun() {
	ss.StopNow = false
	curRun := ss.TrainEnv.Run.Cur
	for {
		ss.TrainTrial()
		if ss.StopNow || ss.TrainEnv.Run.Cur != curRun {
			break
		}
	}
	ss.Stopped()
}

// Train runs the full training from this point onward
func (ss *Sim) Train() {
	ss.StopNow = false
	for {
		ss.TrainTrial()
		if ss.StopNow {
			break
		}
	}
	ss.Stopped()
}

// Stop tells the sim to stop running
func (ss *Sim) Stop() {
	ss.StopNow = true
}

// Stopped is called when a run method stops running -- updates the IsRunning flag and toolbar
func (ss *Sim) Stopped() {
	ss.IsRunning = false
	if ss.Win != nil {
		vp := ss.Win.WinViewport2D()
		if ss.ToolBar != nil {
			ss.ToolBar.UpdateActions()
		}
		vp.SetNeedsFullRender()
	}
}

// SaveWeights saves the network weights with the std wts filename
func (ss *Sim) SaveWeights() {
	if mpi.WorldRank() != 0 {
		return
	}
	fnm := ss.WeightsFileName()
	mpi.Printf("Saving Weights to: %s\n", fnm)
	ss.Net.SaveWtsJSON(gi.FileName(fnm))
}

// EpochSched implements the epoch-wise schedule
func (ss *Sim) EpochSched(epc int) {
	switch epc {
	case 100:
		ss.SaveWeights()
	case 250:
		ss.SaveWeights()
		// ss.Net.LrateSched(0.5)
		// mpi.Printf("dropped lrate to 0.5 at epoch: %d\n", epc)
	case 500:
		ss.SaveWeights()
		// ss.Net.LrateSched(0.2)
		// mpi.Printf("dropped lrate to 0.2 at epoch: %d\n", epc)
	case 750:
		// ss.Net.LrateSched(0.1)
		// mpi.Printf("dropped lrate to 0.1 at epoch: %d\n", epc)
	}
}

// OpenTrainedWts opens trained weights
// func (ss *Sim) OpenTrainedWts() {
// 	ab, err := Asset("objrec_train1.wts") // embedded in executable
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	ss.Net.ReadWtsJSON(bytes.NewBuffer(ab))
// 	// ss.Net.OpenWtsJSON("objrec_train1.wts.gz")
// }

////////////////////////////////////////////////////////////////////////////////////////////
// Testing

// TestTrial runs one trial of testing -- always sequentially presented inputs
func (ss *Sim) TestTrial(returnOnChg bool) {
	ss.TestEnv.Step()

	// Query counters FIRST
	_, _, chg := ss.TestEnv.Counter(env.Epoch)
	if chg {
		if ss.ViewOn && ss.TestUpdt > axon.AlphaCycle {
			ss.UpdateView(false)
		}
		ss.LogTstEpc(ss.TstEpcLog)
		if returnOnChg {
			return
		}
	}

	// note: type must be in place before apply inputs
	ss.ApplyInputs(&ss.TestEnv)
	ss.ThetaCyc(false) // !train
	// todo: actrf etc
	ss.LogTstTrl(ss.TstTrlLog)
}

// TestAll runs through the full set of testing items
func (ss *Sim) TestAll() {
	ss.TestEnv.Init(ss.TrainEnv.Run.Cur)
	ss.ActRFs.Reset()
	for {
		ss.TestTrial(true) // return on chg, don't present
		_, _, chg := ss.TestEnv.Counter(env.Epoch)
		if chg || ss.StopNow {
			break
		}
	}
	ss.ActRFs.Avg()
	ss.ActRFs.Norm()
	ss.ViewActRFs()
}

// RunTestAll runs through the full set of testing items, has stop running = false at end -- for gui
func (ss *Sim) RunTestAll() {
	ss.StopNow = false
	ss.TestAll()
	ss.Stopped()
}

// UpdtActRFs updates activation rf's -- only called during testing
func (ss *Sim) UpdtActRFs() {
	oly := ss.Net.LayerByName("Output")
	ovt := ss.ValsTsr("Output")
	oly.UnitValsTensor(ovt, "ActM")
	// if _, ok := ss.ValsTsrs["Image"]; !ok {
	// 	ss.ValsTsrs["Image"] = &ss.TestEnv.Vis.ImgTsr
	// }
	naf := len(ss.ActRFNms)
	if len(ss.ActRFs.RFs) != naf {
		for _, anm := range ss.ActRFNms {
			sp := strings.Split(anm, ":")
			lnm := sp[0]
			ly := ss.Net.LayerByName(lnm)
			if ly == nil {
				continue
			}
			lvt := ss.ValsTsr(lnm)
			ly.UnitValsTensor(lvt, "ActM")
			tnm := sp[1]
			tvt := ss.ValsTsr(tnm)
			ss.ActRFs.AddRF(anm, lvt, tvt)
			// af.NormRF.SetMetaData("min", "0")
		}
	}
	for _, anm := range ss.ActRFNms {
		sp := strings.Split(anm, ":")
		lnm := sp[0]
		ly := ss.Net.LayerByName(lnm)
		if ly == nil {
			continue
		}
		lvt := ss.ValsTsr(lnm)
		ly.UnitValsTensor(lvt, "ActM")
		tnm := sp[1]
		tvt := ss.ValsTsr(tnm)
		ss.ActRFs.Add(anm, lvt, tvt, 0.01) // thr prevent weird artifacts
	}
}

// ViewActRFs displays act rfs
func (ss *Sim) ViewActRFs() {
	if ss.ActRFGrids == nil {
		return
	}
	for _, nm := range ss.ActRFNms {
		tg := ss.ActRFGrids[nm]
		if tg.Tensor == nil {
			rf := ss.ActRFs.RFByName(nm)
			tg.SetTensor(&rf.NormRF)
		} else {
			tg.UpdateSig()
		}
	}
}

/////////////////////////////////////////////////////////////////////////
//   Params setting

// ParamsName returns name of current set of parameters
func (ss *Sim) ParamsName() string {
	if ss.ParamSet == "" {
		return "Base"
	}
	return ss.ParamSet
}

// SetParams sets the params for "Base" and then current ParamSet.
// If sheet is empty, then it applies all avail sheets (e.g., Network, Sim)
// otherwise just the named sheet
// if setMsg = true then we output a message for each param that was set.
func (ss *Sim) SetParams(sheet string, setMsg bool) error {
	if sheet == "" {
		// this is important for catching typos and ensuring that all sheets can be used
		ss.Params.ValidateSheets([]string{"Network", "Sim"})
	}
	err := ss.SetParamsSet("Base", sheet, setMsg)
	if ss.ParamSet != "" && ss.ParamSet != "Base" {
		err = ss.SetParamsSet(ss.ParamSet, sheet, setMsg)
	}
	return err
}

// SetParamsSet sets the params for given params.Set name.
// If sheet is empty, then it applies all avail sheets (e.g., Network, Sim)
// otherwise just the named sheet
// if setMsg = true then we output a message for each param that was set.
func (ss *Sim) SetParamsSet(setNm string, sheet string, setMsg bool) error {
	pset, err := ss.Params.SetByNameTry(setNm)
	if err != nil {
		return err
	}
	if sheet == "" || sheet == "Network" {
		netp, ok := pset.Sheets["Network"]
		if ok {
			ss.Net.ApplyParams(netp, setMsg)
		}
	}

	if sheet == "" || sheet == "Sim" {
		simp, ok := pset.Sheets["Sim"]
		if ok {
			simp.Apply(ss, setMsg)
		}
	}
	// note: if you have more complex environments with parameters, definitely add
	// sheets for them, e.g., "TrainEnv", "TestEnv" etc
	return err
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Logging

// ValsTsr gets value tensor of given name, creating if not yet made
func (ss *Sim) ValsTsr(name string) *etensor.Float32 {
	if ss.ValsTsrs == nil {
		ss.ValsTsrs = make(map[string]*etensor.Float32)
	}
	tsr, ok := ss.ValsTsrs[name]
	if !ok {
		tsr = &etensor.Float32{}
		ss.ValsTsrs[name] = tsr
	}
	return tsr
}

// RunName returns a name for this run that combines Tag and Params -- add this to
// any file names that are saved.
func (ss *Sim) RunName() string {
	rn := ""
	if ss.Tag != "" {
		rn += ss.Tag + "_"
	}
	rn += ss.ParamsName()
	if ss.StartRun > 0 {
		rn += fmt.Sprintf("_%03d", ss.StartRun)
	}
	return rn
}

// RunEpochName returns a string with the run and epoch numbers with leading zeros, suitable
// for using in weights file names.  Uses 3, 5 digits for each.
func (ss *Sim) RunEpochName(run, epc int) string {
	return fmt.Sprintf("%03d_%05d", run, epc)
}

// WeightsFileName returns default current weights file name
func (ss *Sim) WeightsFileName() string {
	return ss.Net.Nm + "_" + ss.RunName() + "_" + ss.RunEpochName(ss.TrainEnv.Run.Cur, ss.TrainEnv.Epoch.Cur) + ".wts.gz"
}

// LogFileName returns default log file name
func (ss *Sim) LogFileName(lognm string) string {
	nm := ss.Net.Nm + "_" + ss.RunName() + "_" + lognm
	if mpi.WorldRank() > 0 {
		nm += fmt.Sprintf("_%d", mpi.WorldRank())
	}
	nm += ".tsv"
	return nm
}

//////////////////////////////////////////////
//  TrnTrlLog

// LogTrnTrl adds data from current trial to the TrnTrlLog table.
func (ss *Sim) LogTrnTrl(dt *etable.Table) {
	epc := ss.TrainEnv.Epoch.Cur
	trl := ss.TrainEnv.Trial.Cur
	tick := ss.TrainEnv.Tick.Cur
	row := dt.Rows

	if row > 1 { // reset at new epoch
		lstepc := int(dt.CellFloat("Epoch", row-1))
		if lstepc != epc {
			dt.SetNumRows(0)
			row = 0
		}
	}
	if dt.Rows <= row {
		dt.SetNumRows(row + 1)
	}

	dt.SetCellFloat("Run", row, float64(ss.TrainEnv.Run.Cur))
	dt.SetCellFloat("Epoch", row, float64(epc))
	dt.SetCellFloat("Trial", row, float64(trl))
	dt.SetCellFloat("Tick", row, float64(tick))
	dt.SetCellFloat("Idx", row, float64(row))
	dt.SetCellString("Obj", row, "na")
	dt.SetCellString("TrialName", row, ss.TrainEnv.String())

	for li, lnm := range ss.PulvLays {
		dt.SetCellFloat(lnm+"_CosDiff", row, ss.PulvCosDiff[li])
		dt.SetCellFloat(lnm+"_TrlCosDiff", row, ss.PulvTrlCosDiff[li])
		dt.SetCellFloat(lnm+"_UnitErr", row, ss.PulvUnitErr[li])
	}
	for li, lnm := range ss.HidLays {
		dt.SetCellFloat(lnm+"_TrlCosDiff", row, ss.HidTrlCosDiff[li])
	}

	if ss.LastTrlTime.IsZero() {
		ss.LastTrlMSec = 0
	} else {
		iv := time.Now().Sub(ss.LastTrlTime)
		ss.LastTrlMSec = float64(iv) / (float64(time.Millisecond))
	}
	ss.LastTrlTime = time.Now()

	// mpi.Printf("trl: %d %d %d: msec: %5.0f \t obj:%s\n", epc, trl, tick, ss.LastTrlMSec, ss.TrainEnv.String())

	if ss.TrnTrlFile != nil && (!ss.UseMPI || ss.SaveProcLog) { // otherwise written at end of epoch, integrated
		if ss.TrainEnv.Run.Cur == ss.StartRun && epc == 0 && row == 0 {
			dt.WriteCSVHeaders(ss.TrnTrlFile, etable.Tab)
		}
		dt.WriteCSVRow(ss.TrnTrlFile, row, etable.Tab)
	}

	// note: essential to use Go version of update when called from another goroutine
	ss.TrnTrlPlot.GoUpdate()
}

func (ss *Sim) ConfigTrnTrlLog(dt *etable.Table) {
	dt.SetMetaData("name", "TrnTrlLog")
	dt.SetMetaData("desc", "Record of training per input pattern")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Run", etensor.INT64, nil, nil},
		{"Epoch", etensor.INT64, nil, nil},
		{"Trial", etensor.INT64, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"Idx", etensor.INT64, nil, nil},
		{"Obj", etensor.STRING, nil, nil},
		{"TrialName", etensor.STRING, nil, nil},
	}
	for _, lnm := range ss.PulvLays {
		sch = append(sch, etable.Column{lnm + "_CosDiff", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_TrlCosDiff", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_UnitErr", etensor.FLOAT64, nil, nil})
	}
	for _, lnm := range ss.HidLays {
		sch = append(sch, etable.Column{lnm + "_TrlCosDiff", etensor.FLOAT64, nil, nil})
	}

	dt.SetFromSchema(sch, 0)
}

func (ss *Sim) ConfigTrnTrlPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Saccade Predictive Learning Train Trial Plot"
	plt.Params.XAxisCol = "Idx"
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Run", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Epoch", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Trial", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Tick", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Idx", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Obj", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("TrialName", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)

	for _, lnm := range ss.PulvLays {
		plt.SetColParams(lnm+"_CosDiff", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
		plt.SetColParams(lnm+"_TrlCosDiff", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 1)
		plt.SetColParams(lnm+"_UnitErr", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 1)
	}
	for _, lnm := range ss.HidLays {
		plt.SetColParams(lnm+"_TrlCosDiff", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 1)
	}
	return plt
}

//////////////////////////////////////////////
//  TrnTrlRepLog

// CenterPoolsIdxs returns the indexes for 2x2 center pools (including sub-pools):
// nu = number of units per pool, sis = starting indexes
func (ss *Sim) CenterPoolsIdxs(ly *axon.Layer) (nu int, sis []int) {
	nu = ly.Shp.Dim(2) * ly.Shp.Dim(3)
	npy := ly.Shp.Dim(0)
	npx := ly.Shp.Dim(1)
	npxact := npx
	nsp := 1
	if ss.SubPools {
		npy /= 2
		npx /= 2
		nsp = 2
	}
	cpy := (npy - 1) / 2
	cpx := (npx - 1) / 2
	if npx <= 2 {
		cpx = 0
	}
	if npy <= 2 {
		cpy = 0
	}

	for py := 0; py < 2; py++ {
		for px := 0; px < 2; px++ {
			for sy := 0; sy < nsp; sy++ {
				for sx := 0; sx < nsp; sx++ {
					y := (py+cpy)*nsp + sy
					x := (px+cpx)*nsp + sx
					si := (y*npxact + x) * nu
					sis = append(sis, si)
				}
			}
		}
	}
	return
}

// CopyCenterPools copy 2 center pools of ActM to tensor
func (ss *Sim) CopyCenterPools(ly *axon.Layer, vl *etensor.Float32) {
	nu, sis := ss.CenterPoolsIdxs(ly)
	vl.SetShape([]int{len(sis) * nu}, nil, nil)
	ti := 0
	for _, si := range sis {
		for ni := 0; ni < nu; ni++ {
			vl.Values[ti] = ly.Neurons[si+ni].ActM
			ti++
		}
	}
}

// LogTrnRepTrl adds data from current trial to the TrnTrlRepLog table.
func (ss *Sim) LogTrnRepTrl(dt *etable.Table) {
	epc := ss.TrainEnv.Epoch.Cur
	trl := ss.TrainEnv.Trial.Cur
	tick := ss.TrainEnv.Tick.Cur
	row := dt.Rows

	if row > 1 { // reset at new epoch
		lstepc := int(dt.CellFloat("Epoch", row-1))
		if lstepc != epc {
			dt.SetNumRows(0)
			row = 0
		}
	}
	if dt.Rows <= row {
		dt.SetNumRows(row + 1)
	}

	dt.SetCellFloat("Run", row, float64(ss.TrainEnv.Run.Cur))
	dt.SetCellFloat("Epoch", row, float64(epc))
	dt.SetCellFloat("Trial", row, float64(trl))
	dt.SetCellFloat("Tick", row, float64(tick))
	dt.SetCellFloat("Idx", row, float64(row))
	dt.SetCellString("Obj", row, "na")
	dt.SetCellString("TrialName", row, ss.TrainEnv.String())

	for _, lnm := range ss.HidLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		lvt := ss.ValsTsr(lnm)
		if ly.Is4D() && ly.Shp.Dim(0) > 2 && ly.Shp.Dim(2) > 2 && !strings.HasPrefix(ly.Nm, "TE") {
			ss.CopyCenterPools(ly, lvt)
			dt.SetCellTensor(lnm, row, lvt)
		} else {
			ly.UnitValsTensor(lvt, "ActM")
			dt.SetCellTensor(lnm, row, lvt)
		}
	}

	// if ss.TrnTrlFile != nil && (!ss.UseMPI || ss.SaveProcLog) { // otherwise written at end of epoch, integrated
	// 	if ss.TrainEnv.Run.Cur == ss.StartRun && epc == 0 && row == 0 {
	// 		dt.WriteCSVHeaders(ss.TrnTrlFile, etable.Tab)
	// 	}
	// 	dt.WriteCSVRow(ss.TrnTrlFile, row, etable.Tab)
	// }

	// note: essential to use Go version of update when called from another goroutine
	// ss.TrnTrlPlot.GoUpdate()
}

func (ss *Sim) ConfigTrnTrlRepLog(dt *etable.Table) {
	dt.SetMetaData("name", "TrnTrlRepLog")
	dt.SetMetaData("desc", "Record of training per input pattern")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Run", etensor.INT64, nil, nil},
		{"Epoch", etensor.INT64, nil, nil},
		{"Trial", etensor.INT64, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"Idx", etensor.INT64, nil, nil},
		{"Obj", etensor.STRING, nil, nil},
		{"TrialName", etensor.STRING, nil, nil},
	}
	for _, lnm := range ss.HidLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		if ly.Is4D() && ly.Shp.Dim(0) > 2 && ly.Shp.Dim(2) > 2 && !strings.HasPrefix(ly.Nm, "TE") {
			nu, sis := ss.CenterPoolsIdxs(ly)
			sch = append(sch, etable.Column{lnm, etensor.FLOAT64, []int{len(sis) * nu}, nil})
		} else {
			sch = append(sch, etable.Column{lnm, etensor.FLOAT64, ly.Shp.Shp, nil})
		}
	}
	dt.SetFromSchema(sch, 0)
}

//////////////////////////////////////////////
//  TrnEpcLog

// HogDead computes the proportion of units in given layer name with ActAvg over hog thr
// and under dead threshold
func (ss *Sim) HogDead(lnm string) (hog, dead float64) {
	ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
	n := 0
	if ly.Is4D() {
		npy := ly.Shp.Dim(0)
		npx := ly.Shp.Dim(1)
		nny := ly.Shp.Dim(2)
		nnx := ly.Shp.Dim(3)
		nn := nny * nnx
		if npy == 8 { // exclude periphery
			n = 16 * nn
			for py := 2; py < 6; py++ {
				for px := 2; px < 6; px++ {
					pi := (py*npx + px) * nn
					for ni := 0; ni < nn; ni++ {
						nrn := &ly.Neurons[pi+ni]
						if nrn.ActAvg > 0.3 {
							hog += 1
						} else if nrn.ActAvg < 0.01 {
							dead += 1
						}
					}
				}
			}
		} else if ly.Shp.Dim(0) == 4 && ly.Nm[:2] != "TE" {
			n = 4 * nn
			for py := 1; py < 3; py++ {
				for px := 1; px < 3; px++ {
					pi := (py*npx + px) * nn
					for ni := 0; ni < nn; ni++ {
						nrn := &ly.Neurons[pi+ni]
						if nrn.ActAvg > 0.3 {
							hog += 1
						} else if nrn.ActAvg < 0.01 {
							dead += 1
						}
					}
				}
			}
		}
	}
	if n == 0 {
		n = len(ly.Neurons)
		for ni := range ly.Neurons {
			nrn := &ly.Neurons[ni]
			if nrn.ActAvg > 0.3 {
				hog += 1
			} else if nrn.ActAvg < 0.01 {
				dead += 1
			}
		}
	}
	hog /= float64(n)
	dead /= float64(n)
	return
}

// LogTrnEpc adds data from current epoch to the TrnEpcLog table.
// computes epoch averages prior to logging.
func (ss *Sim) LogTrnEpc(dt *etable.Table) {
	// if mpi.WorldRank() == 0 {
	// 	ss.Net.TimerReport()
	// 	ss.Net.ThrTimerReset()
	// }

	row := dt.Rows
	dt.SetNumRows(row + 1)

	trl := ss.TrnTrlLog
	if ss.UseMPI {
		empi.GatherTableRows(ss.TrnTrlLogAll, ss.TrnTrlLog, ss.Comm)
		trl = ss.TrnTrlLogAll
	}

	epc := ss.TrainEnv.Epoch.Prv // this is triggered by increment so use previous value
	nt := float64(trl.Rows)

	if ss.LastEpcTime.IsZero() {
		ss.EpcPerTrlMSec = 0
	} else {
		iv := time.Now().Sub(ss.LastEpcTime)
		ss.EpcPerTrlMSec = float64(iv) / (nt * float64(time.Millisecond))
	}
	ss.LastEpcTime = time.Now()

	dt.SetCellFloat("Run", row, float64(ss.TrainEnv.Run.Cur))
	dt.SetCellFloat("Epoch", row, float64(epc))
	dt.SetCellFloat("PerTrlMSec", row, ss.EpcPerTrlMSec)

	for _, lnm := range ss.HidLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		hog, dead := ss.HogDead(lnm)
		dt.SetCellFloat(lnm+"_Dead", row, dead)
		dt.SetCellFloat(lnm+"_Hog", row, hog)
		dt.SetCellFloat(lnm+"_MaxGeM", row, float64(ly.ActAvg.AvgMaxGeM))
		dt.SetCellFloat(lnm+"_ActAvg", row, float64(ly.ActAvg.ActMAvg))
		dt.SetCellFloat(lnm+"_GiMult", row, float64(ly.ActAvg.GiMult))
	}

	for _, lnm := range ss.InLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		dt.SetCellFloat(lnm+"_ActAvg", row, float64(ly.ActAvg.ActMAvg))
	}

	tix := etable.NewIdxView(trl)
	spl := split.GroupBy(tix, []string{"Tick"})

	// average trial cos diff
	t2tix := etable.NewIdxView(trl)
	t2tix.Filter(func(et *etable.Table, row int) bool {
		tck := int(et.CellFloat("Tick", row))
		return tck >= 2
	})
	t0tix := etable.NewIdxView(trl)
	t0tix.Filter(func(et *etable.Table, row int) bool {
		tck := int(et.CellFloat("Tick", row))
		return tck == 0
	})
	t1tix := etable.NewIdxView(trl)
	t1tix.Filter(func(et *etable.Table, row int) bool {
		tck := int(et.CellFloat("Tick", row))
		return tck == 1
	})

	// np := len(ss.PulvLays)
	for _, lnm := range ss.PulvLays {
		_, err := split.AggTry(spl, lnm+"_CosDiff", agg.AggMean)
		if err != nil {
			log.Println(err)
		}
		split.AggTry(spl, lnm+"_UnitErr", agg.AggMean)
	}
	tags := spl.AggsToTable(etable.ColNameOnly)
	for li, lnm := range ss.PulvLays {
		for tck := 0; tck < ss.MaxTicks; tck++ {
			val := tags.Cols[1+2*li].FloatVal1D(tck)
			dt.SetCellFloat(fmt.Sprintf("%s_CosDiff_%d", lnm, tck), row, val)
			// val = tags.Cols[2+2*li].FloatVal1D(tck)
			// dt.SetCellFloat(fmt.Sprintf("%s_UnitErr_%d", lnm, tck), row, val)
		}
		cdif := agg.Agg(t2tix, lnm+"_TrlCosDiff", agg.AggMean)
		dt.SetCellFloat(lnm+"_TrlCosDiff", row, cdif[0])
		c0dif := agg.Agg(t0tix, lnm+"_TrlCosDiff", agg.AggMean)
		dt.SetCellFloat(lnm+"_TrlCosDiff0", row, c0dif[0])
	}

	for _, lnm := range ss.HidLays {
		cdif := agg.Agg(t2tix, lnm+"_TrlCosDiff", agg.AggMean)
		dt.SetCellFloat(lnm+"_TrlCosDiff", row, cdif[0])
		if strings.HasSuffix(lnm, "CT") { // ct layer has overhang of 1 trial
			c0dif := agg.Agg(t1tix, lnm+"_TrlCosDiff", agg.AggMean)
			dt.SetCellFloat(lnm+"_TrlCosDiff0", row, c0dif[0])
		} else {
			c0dif := agg.Agg(t0tix, lnm+"_TrlCosDiff", agg.AggMean)
			dt.SetCellFloat(lnm+"_TrlCosDiff0", row, c0dif[0])
		}
	}

	for li, lnm := range ss.PulvLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		dt.SetCellFloat(lnm+"_MaxGeM", row, float64(ly.ActAvg.AvgMaxGeM))
		dt.SetCellFloat(lnm+"_ActAvg", row, float64(ly.ActAvg.ActMAvg))
		dt.SetCellFloat(lnm+"_GiMult", row, float64(ly.ActAvg.GiMult))
		for tck := 0; tck < ss.MaxTicks; tck++ {
			val := tags.Cols[1+2*li].FloatVal1D(tck)
			dt.SetCellFloat(fmt.Sprintf("%s_CosDiff_%d", lnm, tck), row, val)
			// val = tags.Cols[2+2*li].FloatVal1D(tck)
			// dt.SetCellFloat(fmt.Sprintf("%s_UnitErr_%d", lnm, tck), row, val)
		}
	}

	if ss.RepsInterval > 0 && epc%ss.RepsInterval == 0 {
		reps := etable.NewIdxView(ss.TrnTrlRepLog)
		if ss.UseMPI {
			empi.GatherTableRows(ss.TrnTrlRepLogAll, ss.TrnTrlRepLog, ss.Comm)
			reps = etable.NewIdxView(ss.TrnTrlRepLogAll)
		}
		// reps.SortColName("Obj", true)
		for _, lnm := range ss.HidLays {
			ss.PCA.TableCol(reps, lnm, metric.Covariance64)
			var nstr float64
			ln := len(ss.PCA.Values)
			for i, v := range ss.PCA.Values {
				// fmt.Printf("%s\t\t %d  %g\n", lnm, i, v)
				if v >= 0.01 {
					nstr = float64(ln - i)
					break
				}
			}
			if ln < 11 {
				fmt.Printf("hid lay < 11: %d  %s\n", ln, lnm)
				continue
			}
			var top5, next5 float64
			for i := 0; i < 5; i++ {
				top5 += ss.PCA.Values[ln-1-i]
				next5 += ss.PCA.Values[ln-6-i]
			}
			sum := norm.Sum64(ss.PCA.Values)
			ravg := (sum - (top5 + next5)) / float64(ln-10)
			dt.SetCellFloat(lnm+"_PCA_NStrong", row, nstr)
			dt.SetCellFloat(lnm+"_PCA_Top5", row, top5/5)
			dt.SetCellFloat(lnm+"_PCA_Next5", row, next5/5)
			dt.SetCellFloat(lnm+"_PCA_Rest", row, ravg)
		}
	} else {
		if row > 0 {
			for _, lnm := range ss.HidLays {
				dt.SetCellFloat(lnm+"_PCA_NStrong", row, dt.CellFloat(lnm+"_PCA_NStrong", row-1))
				dt.SetCellFloat(lnm+"_PCA_Top5", row, dt.CellFloat(lnm+"_PCA_Top5", row-1))
				dt.SetCellFloat(lnm+"_PCA_Next5", row, dt.CellFloat(lnm+"_PCA_Next5", row-1))
				dt.SetCellFloat(lnm+"_PCA_Rest", row, dt.CellFloat(lnm+"_PCA_Rest", row-1))
			}
		}
	}

	// note: essential to use Go version of update when called from another goroutine
	ss.TrnEpcPlot.GoUpdate()
	if ss.TrnEpcFile != nil {
		if ss.TrainEnv.Run.Cur == ss.StartRun && epc == 0 {
			dt.WriteCSVHeaders(ss.TrnEpcFile, etable.Tab)
		}
		dt.WriteCSVRow(ss.TrnEpcFile, row, etable.Tab)
	}

	if ss.TrnTrlFile != nil && !(!ss.UseMPI || ss.SaveProcLog) { // saved at trial level otherwise
		if ss.TrainEnv.Run.Cur == ss.StartRun && epc == 0 {
			trl.WriteCSVHeaders(ss.TrnTrlFile, etable.Tab)
		}
		for ri := 0; ri < trl.Rows; ri++ {
			trl.WriteCSVRow(ss.TrnTrlFile, ri, etable.Tab)
		}
	}
}

func (ss *Sim) ConfigTrnEpcLog(dt *etable.Table) {
	dt.SetMetaData("name", "TrnEpcLog")
	dt.SetMetaData("desc", "Record of performance over epochs of training")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Run", etensor.INT64, nil, nil},
		{"Epoch", etensor.INT64, nil, nil},
		{"PerTrlMSec", etensor.FLOAT64, nil, nil},
	}
	for _, lnm := range ss.HidLays {
		sch = append(sch, etable.Column{lnm + "_Dead", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_Hog", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_MaxGeM", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_ActAvg", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_GiMult", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_TrlCosDiff", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_TrlCosDiff0", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_PCA_NStrong", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_PCA_Top5", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_PCA_Next5", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_PCA_Rest", etensor.FLOAT64, nil, nil})
	}
	for _, lnm := range ss.PulvLays {
		sch = append(sch, etable.Column{lnm + "_TrlCosDiff", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_TrlCosDiff0", etensor.FLOAT64, nil, nil})
	}
	for _, lnm := range ss.InLays {
		sch = append(sch, etable.Column{lnm + "_ActAvg", etensor.FLOAT64, nil, nil})
	}

	for _, lnm := range ss.PulvLays {
		sch = append(sch, etable.Column{lnm + "_MaxGeM", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_ActAvg", etensor.FLOAT64, nil, nil})
		sch = append(sch, etable.Column{lnm + "_GiMult", etensor.FLOAT64, nil, nil})
	}

	for tck := 0; tck < ss.MaxTicks; tck++ {
		for _, lnm := range ss.PulvLays {
			sch = append(sch, etable.Column{fmt.Sprintf("%s_CosDiff_%d", lnm, tck), etensor.FLOAT64, nil, nil})
			// sch = append(sch, etable.Column{fmt.Sprintf("%s_UnitErr_%d", lnm, tck), etensor.FLOAT64, nil, nil})
		}
	}
	dt.SetFromSchema(sch, 0)
}

func (ss *Sim) ConfigTrnEpcPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Saccade Predictive Learning Epoch Plot"
	plt.Params.XAxisCol = "Epoch"
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Run", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Epoch", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("PerTrlMSec", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)

	for _, lnm := range ss.HidLays {
		plt.SetColParams(lnm+"_Dead", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 1)
		plt.SetColParams(lnm+"_Hog", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 1)
		plt.SetColParams(lnm+"_MaxGeM", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
		plt.SetColParams(lnm+"_ActAvg", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
		plt.SetColParams(lnm+"_TrlCosDiff", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
		plt.SetColParams(lnm+"_TrlCosDiff0", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
	}
	for _, lnm := range ss.PulvLays {
		plt.SetColParams(lnm+"_TrlCosDiff", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
		plt.SetColParams(lnm+"_TrlCosDiff0", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
	}
	for _, lnm := range ss.InLays {
		plt.SetColParams(lnm+"_ActAvg", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
	}

	for tck := 0; tck < ss.MaxTicks; tck++ {
		for _, lnm := range ss.PulvLays {
			plt.SetColParams(fmt.Sprintf("%s_CosDiff_%d", lnm, tck), eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
			// plt.SetColParams(fmt.Sprintf("%s_UnitErr_%d", lnm, tck), eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)
		}
	}
	return plt
}

//////////////////////////////////////////////
//  TstTrlLog

// LogTstTrl adds data from current trial to the TstTrlLog table.
func (ss *Sim) LogTstTrl(dt *etable.Table) {
	epc := ss.TrainEnv.Epoch.Prv // this is triggered by increment so use previous value
	trl := ss.TestEnv.Trial.Cur
	row := dt.Rows

	if dt.Rows <= row {
		dt.SetNumRows(row + 1)
	}

	dt.SetCellFloat("Run", row, float64(ss.TrainEnv.Run.Cur))
	dt.SetCellFloat("Epoch", row, float64(epc))
	dt.SetCellFloat("Trial", row, float64(trl))
	dt.SetCellString("Obj", row, "na")
	dt.SetCellString("TrialName", row, ss.TestEnv.String())

	for _, lnm := range ss.LayStatNms {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		if ly.IsOff() {
			continue
		}
		dt.SetCellFloat(ly.Nm+" ActM.Avg", row, float64(ly.Pools[0].ActM.Avg))
	}
	// note: essential to use Go version of update when called from another goroutine
	ss.TstTrlPlot.GoUpdate()
}

func (ss *Sim) ConfigTstTrlLog(dt *etable.Table) {
	// inp := ss.Net.LayerByName("V1").(axon.AxonLayer).AsAxon()
	// out := ss.Net.LayerByName("Output").(axon.AxonLayer).AsAxon()

	dt.SetMetaData("name", "TstTrlLog")
	dt.SetMetaData("desc", "Record of testing per input pattern")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Run", etensor.INT64, nil, nil},
		{"Epoch", etensor.INT64, nil, nil},
		{"Trial", etensor.INT64, nil, nil},
		{"Obj", etensor.STRING, nil, nil},
		{"TrialName", etensor.STRING, nil, nil},
	}
	for _, lnm := range ss.LayStatNms {
		sch = append(sch, etable.Column{lnm + " ActM.Avg", etensor.FLOAT64, nil, nil})
	}
	dt.SetFromSchema(sch, 0)
}

func (ss *Sim) ConfigTstTrlPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Saccade Predictive Learning Test Trial Plot"
	plt.Params.XAxisCol = "Trial"
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Run", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Epoch", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Trial", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Obj", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("TrialName", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)

	for _, lnm := range ss.LayStatNms {
		plt.SetColParams(lnm+" ActM.Avg", eplot.Off, eplot.FixMin, 0, eplot.FixMax, 0.5)
	}
	return plt
}

//////////////////////////////////////////////
//  TstEpcLog

func (ss *Sim) LogTstEpc(dt *etable.Table) {
	trl := ss.TstTrlLog
	if ss.UseMPI {
		empi.GatherTableRows(ss.TstTrlLogAll, ss.TstTrlLog, ss.Comm)
		trl = ss.TstTrlLogAll
	}
	tix := etable.NewIdxView(trl)
	// epc := ss.TrainEnv.Epoch.Prv // ?

	spl := split.GroupBy(tix, []string{"Obj"})
	_, err := split.AggTry(spl, "Err", agg.AggMean)
	if err != nil {
		log.Println(err)
	}
	objs := spl.AggsToTable(etable.AddAggName)
	no := objs.Rows
	dt.SetNumRows(no)
	for i := 0; i < no; i++ {
		dt.SetCellFloat("Obj", i, float64(i))
		dt.SetCellFloat("PctErr", i, objs.Cols[1].FloatVal1D(i))
	}
	ss.TstEpcPlot.GoUpdate()
}

func (ss *Sim) ConfigTstEpcLog(dt *etable.Table) {
	dt.SetMetaData("name", "TstEpcLog")
	dt.SetMetaData("desc", "Summary stats for testing trials")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	dt.SetFromSchema(etable.Schema{
		{"Obj", etensor.INT64, nil, nil},
		{"PctErr", etensor.FLOAT64, nil, nil},
	}, 0)
}

func (ss *Sim) ConfigTstEpcPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Saccade Predictive Learning Testing Epoch Plot"
	plt.Params.XAxisCol = "Obj"
	plt.Params.Type = eplot.Bar
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Obj", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("PctErr", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	return plt
}

//////////////////////////////////////////////
//  RunLog

// LogRun adds data from current run to the RunLog table.
func (ss *Sim) LogRun(dt *etable.Table) {
	run := ss.TrainEnv.Run.Cur // this is NOT triggered by increment yet -- use Cur
	row := dt.Rows
	dt.SetNumRows(row + 1)

	epclog := ss.TrnEpcLog
	epcix := etable.NewIdxView(epclog)
	// compute mean over last N epochs for run level
	nlast := 5
	if nlast > epcix.Len()-1 {
		nlast = epcix.Len() - 1
	}
	epcix.Idxs = epcix.Idxs[epcix.Len()-nlast:]

	// params := ss.Params.Name
	params := "params"

	// todo: fix or will crash..
	dt.SetCellFloat("Run", row, float64(run))
	dt.SetCellString("Params", row, params)

	// runix := etable.NewIdxView(dt)
	// spl := split.GroupBy(runix, []string{"Params"})
	// split.Desc(spl, "FirstZero")
	// split.Desc(spl, "PctCor")
	// ss.RunStats = spl.AggsToTable(etable.AddAggName)

	// note: essential to use Go version of update when called from another goroutine
	ss.RunPlot.GoUpdate()
	if ss.RunFile != nil {
		if row == 0 {
			dt.WriteCSVHeaders(ss.RunFile, etable.Tab)
		}
		dt.WriteCSVRow(ss.RunFile, row, etable.Tab)
	}
}

func (ss *Sim) ConfigRunLog(dt *etable.Table) {
	dt.SetMetaData("name", "RunLog")
	dt.SetMetaData("desc", "Record of performance at end of training")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	dt.SetFromSchema(etable.Schema{
		{"Run", etensor.INT64, nil, nil},
		{"Params", etensor.STRING, nil, nil},
	}, 0)
}

func (ss *Sim) ConfigRunPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Saccade Predictive Learning Run Plot"
	plt.Params.XAxisCol = "Run"
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Run", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	return plt
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

func (ss *Sim) ConfigNetView(nv *netview.NetView) {
	nv.ViewDefaults()
	cam := &(nv.Scene().Camera)
	cam.Pose.Pos.Set(0.0, 2.4, 2.4)
	cam.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})
	// cam.Pose.Quat.SetFromAxisAngle(mat32.Vec3{-1, 0, 0}, 0.4077744)
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("saccade")
	gi.SetAppAbout(`saccade does deep predictive learning on saccade eye movements. See <a href="https://github.com/ccnlab/deep-obj-cat/blob/master/sims/saccade/README.md">README.md on GitHub</a>.</p>`)

	win := gi.NewMainWindow("saccade", "Saccade", width, height)
	ss.Win = win

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tbar := gi.AddNewToolBar(mfr, "tbar")
	tbar.SetStretchMaxWidth()
	ss.ToolBar = tbar

	split := gi.AddNewSplitView(mfr, "split")
	split.Dim = mat32.X
	split.SetStretchMax()

	sv := giv.AddNewStructView(split, "sv")
	sv.SetStruct(ss)

	tv := gi.AddNewTabView(split, "tv")

	nv := tv.AddNewTab(netview.KiT_NetView, "NetView").(*netview.NetView)
	nv.Params.Defaults()
	nv.Params.LayNmSize = 0.03
	nv.Params.MaxRecs = 104
	nv.Var = "Act"
	nv.SetNet(ss.Net)
	ss.NetView = nv
	ss.ConfigNetView(nv)

	plt := tv.AddNewTab(eplot.KiT_Plot2D, "TrnTrlPlot").(*eplot.Plot2D)
	ss.TrnTrlPlot = ss.ConfigTrnTrlPlot(plt, ss.TrnTrlLog)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "TrnEpcPlot").(*eplot.Plot2D)
	ss.TrnEpcPlot = ss.ConfigTrnEpcPlot(plt, ss.TrnEpcLog)

	tg := tv.AddNewTab(etview.KiT_TensorGrid, "Image").(*etview.TensorGrid)
	tg.SetStretchMax()
	tg.Disp.Defaults()
	// tg.Disp.Image = true
	tg.Disp.ColorMap = giv.ColorMapName("DarkLight")
	ss.CurImgGrid = tg
	// tg.SetTensor(&ss.TrainEnv.V1Hi.ImgTsr)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "TstTrlPlot").(*eplot.Plot2D)
	ss.TstTrlPlot = ss.ConfigTstTrlPlot(plt, ss.TstTrlLog)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "TstEpcPlot").(*eplot.Plot2D)
	ss.TstEpcPlot = ss.ConfigTstEpcPlot(plt, ss.TstEpcLog)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "RunPlot").(*eplot.Plot2D)
	ss.RunPlot = ss.ConfigRunPlot(plt, ss.RunLog)

	ss.SacTableView = tv.AddNewTab(etview.KiT_TableView, "EnvTable").(*etview.TableView)
	ss.SacTableView.SetTable(&ss.TrainEnv.Table, nil)

	split.SetSplits(.2, .8)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update", Tooltip: "Initialize everything including network weights, and start over.  Also applies current params.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Init()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "Train", Icon: "run", Tooltip: "Starts the network training, picking up from wherever it may have left off.  If not stopped, training will complete the specified number of Runs through the full number of Epochs of training, with testing automatically occuring at the specified interval.",
		UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!ss.IsRunning)
		}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !ss.IsRunning {
			ss.IsRunning = true
			tbar.UpdateActions()
			go ss.Train()
		}
	})

	tbar.AddAction(gi.ActOpts{Label: "Stop", Icon: "stop", Tooltip: "Interrupts running.  Hitting Train again will pick back up where it left off.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Stop()
	})

	tbar.AddAction(gi.ActOpts{Label: "Step Trial", Icon: "step-fwd", Tooltip: "Advances one training trial at a time.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !ss.IsRunning {
			ss.IsRunning = true
			ss.TrainTrial()
			ss.IsRunning = false
			vp.SetNeedsFullRender()
		}
	})

	tbar.AddAction(gi.ActOpts{Label: "Step Epoch", Icon: "fast-fwd", Tooltip: "Advances one epoch (complete set of training patterns) at a time.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !ss.IsRunning {
			ss.IsRunning = true
			tbar.UpdateActions()
			go ss.TrainEpoch()
		}
	})

	tbar.AddAction(gi.ActOpts{Label: "Step Run", Icon: "fast-fwd", Tooltip: "Advances one full training Run at a time.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !ss.IsRunning {
			ss.IsRunning = true
			tbar.UpdateActions()
			go ss.TrainRun()
		}
	})

	tbar.AddSeparator("spcl")

	// tbar.AddAction(gi.ActOpts{Label: "Config Net", Icon: "update", Tooltip: "configure and build the network", UpdateFunc: func(act *gi.Action) {
	// 	act.SetActiveStateUpdt(!ss.IsRunning)
	// }}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
	// 	ss.ConfigNet(ss.Net)
	// 	ss.NetView.Config()
	// 	vp.SetNeedsFullRender()
	// })

	tbar.AddAction(gi.ActOpts{Label: "Open Trained Wts", Icon: "update", Tooltip: "open weights trained on first phase of training (excluding 'novel' objects)", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		// ss.OpenTrainedWts()
		vp.SetNeedsFullRender()
	})

	tbar.AddSeparator("test")

	tbar.AddAction(gi.ActOpts{Label: "Test Trial", Icon: "step-fwd", Tooltip: "Runs the next testing trial.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !ss.IsRunning {
			ss.IsRunning = true
			ss.TestTrial(false) // don't break on chg
			ss.IsRunning = false
			vp.SetNeedsFullRender()
		}
	})

	tbar.AddAction(gi.ActOpts{Label: "Test All", Icon: "fast-fwd", Tooltip: "Tests all of the testing trials.", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!ss.IsRunning)
	}}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !ss.IsRunning {
			ss.IsRunning = true
			tbar.UpdateActions()
			go ss.RunTestAll()
		}
	})

	tbar.AddSeparator("log")

	tbar.AddAction(gi.ActOpts{Label: "Reset RunLog", Icon: "update", Tooltip: "Reset the accumulated log of all Runs, which are tagged with the ParamSet used"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.RunLog.SetNumRows(0)
			ss.RunPlot.Update()
		})

	tbar.AddSeparator("misc")

	tbar.AddAction(gi.ActOpts{Label: "New Seed", Icon: "new", Tooltip: "Generate a new initial random seed to get different results.  By default, Init re-establishes the same initial seed every time."}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.NewRndSeed()
		})

	tbar.AddAction(gi.ActOpts{Label: "README", Icon: "file-markdown", Tooltip: "Opens your browser on the README file that contains instructions for how to run this model."}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			gi.OpenURL("https://github.com/ccnlab/deep-obj-cat/blob/master/sims/saccade/README.md")
		})

	vp.UpdateEndNoSig(updt)

	// main menu
	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(win)

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(win)

	// note: Command in shortcuts is automatically translated into Control for
	// Linux, Windows or Meta for MacOS
	// fmen := win.MainMenu.ChildByName("File", 0).(*gi.Action)
	// fmen.Menu.AddAction(gi.ActOpts{Label: "Open", Shortcut: "Command+O"},
	// 	win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
	// 		FileViewOpenSVG(vp)
	// 	})
	// fmen.Menu.AddSeparator("csep")
	// fmen.Menu.AddAction(gi.ActOpts{Label: "Close Window", Shortcut: "Command+W"},
	// 	win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
	// 		win.Close()
	// 	})

	/*
		inQuitPrompt := false
		gi.SetQuitReqFunc(func() {
			if inQuitPrompt {
				return
			}
			inQuitPrompt = true
			gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Quit?",
				Prompt: "Are you <i>sure</i> you want to quit and lose any unsaved params, weights, logs, etc?"}, gi.AddOk, gi.AddCancel,
				win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					if sig == int64(gi.DialogAccepted) {
						gi.Quit()
					} else {
						inQuitPrompt = false
					}
				})
		})

		// gi.SetQuitCleanFunc(func() {
		// 	mpi.Printf("Doing final Quit cleanup here..\n")
		// })

		inClosePrompt := false
		win.SetCloseReqFunc(func(w *gi.Window) {
			if inClosePrompt {
				return
			}
			inClosePrompt = true
			gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Close Window?",
				Prompt: "Are you <i>sure</i> you want to close the window?  This will Quit the App as well, losing all unsaved params, weights, logs, etc"}, gi.AddOk, gi.AddCancel,
				win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					if sig == int64(gi.DialogAccepted) {
						gi.Quit()
					} else {
						inClosePrompt = false
					}
				})
		})
	*/

	win.SetCloseCleanFunc(func(w *gi.Window) {
		go gi.Quit() // once main window is closed, quit
	})

	win.MainMenuUpdated()
	return win
}

// These props register Save methods so they can be used
var SimProps = ki.Props{
	"CallMethods": ki.PropSlice{
		{"OpenSimMat", ki.Props{
			"desc": "Open a TEsim TE similarity matrix in standard object order",
			"icon": "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".tsv",
				}},
			},
		}},
		{"OpenCatActs", ki.Props{
			"desc": "Open a catact file with layer activity by category, and then run the RSA analysis on it -- see RSA for results",
			"icon": "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".tsv",
				}},
			},
		}},
	},
}

func (ss *Sim) CmdArgs() {
	ss.NoGui = true
	var nogui bool
	var saveEpcLog bool
	var saveTrlLog bool
	var saveRunLog bool
	var note string
	flag.StringVar(&ss.ParamSet, "params", "", "ParamSet name to use -- must be valid name as listed in compiled-in params or loaded params")
	flag.StringVar(&ss.Tag, "tag", "", "extra tag to add to file names saved from this run")
	flag.StringVar(&note, "note", "", "user note -- describe the run params etc")
	flag.IntVar(&ss.StartRun, "run", 0, "starting run number -- determines the random seed -- runs counts from there -- can do all runs in parallel by launching separate jobs with each run, runs = 1")
	flag.IntVar(&ss.MaxRuns, "runs", 1, "number of runs to do (note that MaxEpcs is in paramset)")
	flag.BoolVar(&ss.LogSetParams, "setparams", false, "if true, print a record of each parameter that is set")
	flag.BoolVar(&ss.SaveWts, "wts", false, "if true, save final weights after each run")
	flag.BoolVar(&saveEpcLog, "epclog", true, "if true, save train epoch log to file")
	flag.BoolVar(&saveTrlLog, "trllog", false, "if true, save train trial log to file")
	flag.BoolVar(&saveRunLog, "runlog", true, "if true, save run epoch log to file")
	flag.BoolVar(&ss.SaveProcLog, "proclog", false, "if true, save log files separately for each processor (for debugging)")
	flag.BoolVar(&nogui, "nogui", true, "if not passing any other args and want to run nogui, use nogui")
	flag.BoolVar(&ss.UseMPI, "mpi", false, "if set, use MPI for distributed computation")
	flag.Parse()

	if ss.UseMPI {
		ss.MPIInit()
	}

	// key for Config and Init to be after MPIInit
	ss.Config()
	ss.Init()

	if note != "" {
		mpi.Printf("note: %s\n", note)
	}
	if ss.ParamSet != "" {
		mpi.Printf("Using ParamSet: %s\n", ss.ParamSet)
	}

	if saveEpcLog && (ss.SaveProcLog || mpi.WorldRank() == 0) {
		var err error
		fnm := ss.LogFileName("epc")
		ss.TrnEpcFile, err = os.Create(fnm)
		if err != nil {
			log.Println(err)
			ss.TrnEpcFile = nil
		} else {
			mpi.Printf("Saving epoch log to: %v\n", fnm)
			defer ss.TrnEpcFile.Close()
		}
	}
	if saveTrlLog && (ss.SaveProcLog || mpi.WorldRank() == 0) {
		var err error
		fnm := ss.LogFileName("trl")
		ss.TrnTrlFile, err = os.Create(fnm)
		if err != nil {
			log.Println(err)
			ss.TrnTrlFile = nil
		} else {
			mpi.Printf("Saving trial log to: %v\n", fnm)
			defer ss.TrnTrlFile.Close()
		}
	}
	if saveRunLog && (ss.SaveProcLog || mpi.WorldRank() == 0) {
		var err error
		fnm := ss.LogFileName("run")
		ss.RunFile, err = os.Create(fnm)
		if err != nil {
			log.Println(err)
			ss.RunFile = nil
		} else {
			mpi.Printf("Saving run log to: %v\n", fnm)
			defer ss.RunFile.Close()
		}
	}
	if ss.SaveWts {
		if mpi.WorldRank() != 0 {
			ss.SaveWts = false
		}
		mpi.Printf("Saving final weights per run\n")
	}
	mpi.Printf("Running %d Runs starting at %d\n", ss.MaxRuns, ss.StartRun)
	ss.TrainEnv.Run.Set(ss.StartRun)
	ss.TrainEnv.Run.Max = ss.StartRun + ss.MaxRuns
	ss.NewRun()
	ss.Train()
	ss.MPIFinalize()
}

////////////////////////////////////////////////////////////////////
//  MPI code

// MPIInit initializes MPI
func (ss *Sim) MPIInit() {
	mpi.Init()
	var err error
	ss.Comm, err = mpi.NewComm(nil) // use all procs
	if err != nil {
		log.Println(err)
		ss.UseMPI = false
	} else {
		mpi.Printf("MPI running on %d procs\n", mpi.WorldSize())
	}
}

// MPIFinalize finalizes MPI
func (ss *Sim) MPIFinalize() {
	if ss.UseMPI {
		mpi.Finalize()
	}
}

// CollectDWts collects the weight changes from all synapses into AllDWts
func (ss *Sim) CollectDWts(net *axon.Network) {
	net.CollectDWts(&ss.AllDWts) // plug in number from printout below, to avoid realloc
}

// MPIWtFmDWt updates weights from weight changes, using MPI to integrate
// DWt changes across parallel nodes, each of which are learning on different
// sequences of inputs.
func (ss *Sim) MPIWtFmDWt() {
	if ss.UseMPI {
		ss.CollectDWts(&ss.Net.Network)
		ndw := len(ss.AllDWts)
		if len(ss.SumDWts) != ndw {
			ss.SumDWts = make([]float32, ndw)
		}
		ss.Comm.AllReduceF32(mpi.OpSum, ss.SumDWts, ss.AllDWts)
		ss.Net.SetDWts(ss.SumDWts, mpi.WorldSize())
	}
	ss.Net.WtFmDWt()
}
