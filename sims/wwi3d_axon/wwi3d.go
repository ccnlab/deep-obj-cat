// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
wwi3d does deep predictive learning of 3D objects tumbling through space, with
periodic saccadic eye movements, providing plenty of opportunity for prediction errors.
wwi = what, where integration: both pathways combine to predict object --
*where* (dorsal) pathway is trained first and residual prediction error trains *what* pathway.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
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
	LIPOnly          bool            `desc:"if true, only build, train the LIP portion"`
	BinarizeV1       bool            `desc:"if true, V1 inputs are binarized -- todo: test continued need for this"`
	TrnTrlLog        *etable.Table   `view:"no-inline" desc:"training trial-level log data"`
	TrnTrlLogAll     *etable.Table   `view:"no-inline" desc:"all training trial-level log data (aggregated from MPI)"`
	TrnTrlRepLog     *etable.Table   `view:"no-inline" desc:"training trial-level reps log data"`
	TrnTrlRepLogAll  *etable.Table   `view:"no-inline" desc:"training trial-level reps log data"`
	CatLayActs       *etable.Table   `view:"no-inline" desc:"super layer activations per category / object"`
	CatLayActsDest   *etable.Table   `view:"no-inline" desc:"MPI dest super layer activations per category / object"`
	RSA              RSA             `view:"no-inline" desc:"RSA data"`
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
	Prjn8x8Skp4      *prjn.PoolTile  `view:"2x Standard feedforward topographic projection, recv = 1/2 send size"`
	Prjn8x8Skp4Recip *prjn.PoolTile  `view:"Reciprocal"`
	Prjn2x2Skp2      *prjn.PoolTile  `view:"sparser skip 2 -- no overlap"`
	Prjn2x2Skp2Recip *prjn.PoolTile  `view:"Reciprocal"`
	Prjn4x4Skp4      *prjn.PoolTile  `view:"no ovlp for smaller layers"`
	Prjn4x4Skp4Recip *prjn.PoolTile  `view:"Reciprocal"`
	Prjn3x3Skp1      *prjn.PoolTile  `view:"Standard same-to-same size topographic projection"`
	Prjn5x5Skp1      *prjn.PoolTile  `view:"Standard same-to-same size topographic projection"`
	PrjnSigTopo      *prjn.PoolTile  `view:"sigmoidal topographic projection used in LIP saccade remapping layers"`
	PrjnGaussTopo    *prjn.PoolTile  `view:"gaussian topographic projection used in LIP saccade remapping layers"`
	StartRun         int             `desc:"starting run number -- typically 0 but can be set in command args for parallel runs on a cluster"`
	MaxRuns          int             `desc:"maximum number of model runs to perform (starting from StartRun)"`
	MaxEpcs          int             `desc:"maximum number of epochs to run per model run"`
	MaxTrls          int             `desc:"maximum number of training trials per epoch (each trial is MaxTicks ticks)"`
	MaxTicks         int             `desc:"max number of ticks, for logs, stats"`
	NZeroStop        int             `desc:"if a positive number, training will stop after this many epochs with zero SSE"`
	RepsInterval     int             `desc:"how often to analyze the representations"`
	TrainEnv         Obj3DSacEnv     `desc:"Training environment -- 3D Object training"`
	TestEnv          Obj3DSacEnv     `desc:"Testing environment -- testing 3D Objects"`
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
	ss.LIPOnly = false
	ss.BinarizeV1 = false
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
	ss.RSA.Interval = 10

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

	ss.Prjn8x8Skp4 = prjn.NewPoolTile()
	ss.Prjn8x8Skp4.Size.Set(8, 8)
	ss.Prjn8x8Skp4.Skip.Set(4, 4)
	ss.Prjn8x8Skp4.Start.Set(-2, -2)
	ss.Prjn8x8Skp4.TopoRange.Min = 0.8 // note: none of these make a very big diff

	ss.Prjn8x8Skp4Recip = prjn.NewPoolTile()
	ss.Prjn8x8Skp4Recip.Size.Set(8, 8)
	ss.Prjn8x8Skp4Recip.Skip.Set(4, 4)
	ss.Prjn8x8Skp4Recip.Start.Set(-2, -2)
	ss.Prjn8x8Skp4Recip.TopoRange.Min = 0.8 // note: none of these make a very big diff
	ss.Prjn8x8Skp4Recip.Recip = true

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

	ss.Prjn4x4Skp4 = prjn.NewPoolTile()
	ss.Prjn4x4Skp4.Size.Set(4, 4)
	ss.Prjn4x4Skp4.Skip.Set(4, 4)
	ss.Prjn4x4Skp4.Start.Set(0, 0)
	ss.Prjn4x4Skp4.TopoRange.Min = 0.8

	ss.Prjn4x4Skp4Recip = prjn.NewPoolTile()
	ss.Prjn4x4Skp4Recip.Size.Set(4, 4)
	ss.Prjn4x4Skp4Recip.Skip.Set(4, 4)
	ss.Prjn4x4Skp4Recip.Start.Set(0, 0)
	ss.Prjn4x4Skp4Recip.TopoRange.Min = 0.8
	ss.Prjn4x4Skp4Recip.Recip = true

	ss.Prjn3x3Skp1 = prjn.NewPoolTile()
	ss.Prjn3x3Skp1.Size.Set(3, 3)
	ss.Prjn3x3Skp1.Skip.Set(1, 1)
	ss.Prjn3x3Skp1.Start.Set(-1, -1)
	ss.Prjn3x3Skp1.TopoRange.Min = 0.8 // note: none of these make a very big diff

	ss.Prjn5x5Skp1 = prjn.NewPoolTile()
	ss.Prjn5x5Skp1.Size.Set(5, 5)
	ss.Prjn5x5Skp1.Skip.Set(1, 1)
	ss.Prjn5x5Skp1.Start.Set(-2, -2)
	ss.Prjn5x5Skp1.TopoRange.Min = 0.8 // note: none of these make a very big diff

	ss.PrjnSigTopo = prjn.NewPoolTile()
	ss.PrjnSigTopo.GaussOff()
	ss.PrjnSigTopo.Size.Set(1, 1)
	ss.PrjnSigTopo.Skip.Set(0, 0)
	ss.PrjnSigTopo.Start.Set(0, 0)
	ss.PrjnSigTopo.TopoRange.Min = 0.6
	ss.PrjnSigTopo.SigFull.On = true
	ss.PrjnSigTopo.SigFull.Gain = 0.05
	ss.PrjnSigTopo.SigFull.CtrMove = 0.5

	ss.PrjnGaussTopo = prjn.NewPoolTile()
	ss.PrjnGaussTopo.Size.Set(1, 1)
	ss.PrjnGaussTopo.Skip.Set(0, 0)
	ss.PrjnGaussTopo.Start.Set(0, 0)
	ss.PrjnGaussTopo.TopoRange.Min = 0.6
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
	ss.ConfigCatLayActs(ss.CatLayActs)
	if ss.UseMPI {
		ss.ConfigCatLayActs(ss.CatLayActsDest)
	}
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
		if ss.LIPOnly {
			ss.MaxEpcs = 100
		} else {
			ss.MaxEpcs = 999 // 500
		}
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
	ss.TrainEnv.V1Med.Binarize = ss.BinarizeV1
	ss.TrainEnv.V1Hi.Binarize = ss.BinarizeV1

	ss.TestEnv.Nm = "TestEnv"
	ss.TestEnv.Dsc = "testing params and state"
	ss.TestEnv.Defaults()
	ss.TestEnv.Path = "images/test"
	ss.TestEnv.Trial.Max = 500
	ss.TestEnv.V1Med.Binarize = ss.BinarizeV1
	ss.TestEnv.V1Hi.Binarize = ss.BinarizeV1

	ss.TrainEnv.Init(0)
	ss.TestEnv.Init(0)
	if ss.UseMPI { // filter trials to subset for each proc
		st, ed, _ := empi.AllocN(ss.MaxTrls)
		ss.TrainEnv.IdxView = etable.NewIdxView(ss.TrainEnv.Table)
		ss.TrainEnv.IdxView.Filter(func(et *etable.Table, row int) bool {
			trl := int(et.CellFloat("Trial", row))
			return trl >= st && trl < ed
		})
		mpi.Printf("trial allocs: %d .. %d  idx len: %d\n", st, ed, ss.TrainEnv.IdxView.Len())
	}
	ss.TrainEnv.Validate()
	ss.TestEnv.Validate()
}

func (ss *Sim) ConfigNet(net *deep.Network) {
	net.InitName(net, "WWI3D")
	ss.ConfigNetLIP(net)

	if !ss.LIPOnly {
		ss.ConfigNetRest(net)
	}

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

// ConfigNetLIP configures just the V1 and LIP dorsal path part
func (ss *Sim) ConfigNetLIP(net *deep.Network) {
	v1m := net.AddLayer4D("V1m", 8, 8, 5, 4, emer.Input)
	v1h := net.AddLayer4D("V1h", 16, 16, 5, 4, emer.Input)

	lip, lipct, lipp := net.AddSuperCTTRC4D("LIP", 16, 16, 1, 1) // 4, 4 tiny bit better than 2,2
	lipp.SetName("LIPP")
	// lipct.Shape().SetShape([]int{16, 16, 1, 1}, nil, nil)
	// lipp.Shape().SetShape([]int{16, 16, 1, 1}, nil, nil)

	mtpos := net.AddLayer4D("MTPos", 16, 16, 1, 1, emer.Hidden)

	lipp.(*deep.TRCLayer).Driver = "MTPos"

	eyepos := net.AddLayer2D("EyePos", 21, 21, emer.Input)
	sacplan := net.AddLayer2D("SacPlan", 11, 11, emer.Input)
	sac := net.AddLayer2D("Saccade", 11, 11, emer.Input)
	objvel := net.AddLayer2D("ObjVel", 11, 11, emer.Input)

	v1m.SetClass("V1")
	v1h.SetClass("V1")

	mtpos.SetClass("LIP")
	lip.SetClass("LIP")
	lipct.SetClass("LIP")
	lipp.SetClass("LIP")
	sacplan.SetClass("PopIn")
	sac.SetClass("PopIn")
	objvel.SetClass("PopIn")

	full := prjn.NewFull()
	pone2one := prjn.NewPoolOneToOne()

	net.ConnectLayers(v1h, mtpos, pone2one, emer.Forward).SetClass("V1MT")
	net.ConnectLayers(v1m, mtpos, ss.Prjn4x4Skp2Recip, emer.Forward).SetClass("V1MT")
	net.ConnectLayers(mtpos, lip, ss.Prjn3x3Skp1, emer.Forward).SetClass("MTLIP") // was pone2one
	// net.ConnectCtxtToCT(lipct, lipct, full).SetClass("CTSelfLIP")           // only helpful with rel = 2

	lipct.RecvPrjns().SendName("LIP").SetPattern(full) // critical to have full input

	// topo CT <-> P is critical
	net.ConnectLayers(lipct, lipp, ss.Prjn3x3Skp1, emer.Forward).SetClass("CTToPulv")
	net.ConnectLayers(lipp, lipct, ss.Prjn3x3Skp1, emer.Forward).SetClass("FmPulv") //  FmLIP
	net.ConnectLayers(lipp, lip, full, emer.Forward).SetClass("FmPulv")             //  FmLIP

	// lip.RecvPrjns().SendName("LIPP").SetClass("FmPulv FmLIP")
	// lipct.RecvPrjns().SendName("LIPP").SetClass("FmPulv FmLIP")
	// lipct.RecvPrjns().SendName("LIP").SetClass("CTCtxtStd")

	// lip can really be a pure position rep
	// net.ConnectLayers(eyepos, lip, full, emer.Forward)  // InitWts sets ss.PrjnGaussTopo
	// net.ConnectLayers(sacplan, lip, full, emer.Forward) // InitWts sets ss.PrjnSigTopo
	// net.ConnectLayers(objvel, lip, full, emer.Forward)  // InitWts sets ss.PrjnSigTopo

	net.ConnectLayers(eyepos, lipct, full, emer.Forward) // InitWts sets ss.PrjnGaussTopo
	net.ConnectLayers(sac, lipct, full, emer.Forward)    // InitWts sets ss.PrjnSigTopo
	net.ConnectLayers(objvel, lipct, full, emer.Forward) // InitWts sets ss.PrjnSigTopo

	//	Position
	v1h.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v1m.Name(), YAlign: relpos.Front, Space: 2})

	lip.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: v1m.Name(), XAlign: relpos.Left, YAlign: relpos.Front})
	lipct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: lip.Name(), XAlign: relpos.Left, Space: 10})
	lipp.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: lipct.Name(), XAlign: relpos.Left, Space: 10})
	mtpos.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: lipp.Name(), XAlign: relpos.Left, Space: 10})

	eyepos.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: lip.Name(), YAlign: relpos.Front, Space: 2})
	sacplan.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: eyepos.Name(), XAlign: relpos.Left, Space: 10})
	sac.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: sacplan.Name(), XAlign: relpos.Left, Space: 10})
	objvel.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: sac.Name(), XAlign: relpos.Left, Space: 10})

}

// ConfigNetRest configures the rest of the network
func (ss *Sim) ConfigNetRest(net *deep.Network) {
	// note: important for pulvinar to be created first, for weight symmetry init
	v1hp := deep.AddTRCLayer4D(net.AsAxon(), "V1hP", 16, 16, 5, 4)
	v1hp.SetClass("V1")
	v1hp.Driver = "V1h"

	v1mp := deep.AddTRCLayer4D(net.AsAxon(), "V1mP", 8, 8, 5, 4)
	v1mp.SetClass("V1")
	v1mp.Driver = "V1m"

	v2, v2ct := net.AddSuperCT4D("V2", 8, 8, 10, 10)

	v3, v3ct := net.AddSuperCT4D("V3", 4, 4, 10, 10)

	dp, dpct := net.AddSuperCT4D("DP", 1, 1, 10, 10)

	v4, v4ct := net.AddSuperCT4D("V4", 4, 4, 10, 10)

	teo, teoct := net.AddSuperCT4D("TEO", 4, 4, 10, 10) // 2x2 doesn't work with big V2 topo prjn

	te, tect := net.AddSuperCT4D("TE", 2, 2, 10, 10)

	v2.SetClass("V2")
	v2ct.SetClass("V2")
	// v2p.SetClass("V2")

	v3.SetClass("V3")
	v3ct.SetClass("V3")
	// v3p.SetClass("V3")

	v4.SetClass("V4")
	v4ct.SetClass("V4")
	// v4p.SetClass("V4")

	dp.SetClass("DP")
	dpct.SetClass("DP")
	// dpp.SetClass("DP")

	teo.SetClass("TEO")
	teoct.SetClass("TEO")
	// teop.SetClass("TEO")

	te.SetClass("TE")
	tect.SetClass("TE")
	// tep.SetClass("TE")

	v1m := net.LayerByName("V1m")
	v1h := net.LayerByName("V1h")
	lip := net.LayerByName("LIP")
	lipct := net.LayerByName("LIPCT")
	eyepos := net.LayerByName("EyePos")

	// lesion stuff here
	/*
		dp.SetOff(true)
		dpct.SetOff(true)
		dpp.SetOff(true)

		v3.SetOff(true)
		v3ct.SetOff(true)
		v3p.SetOff(true)

		lip.SetOff(true)
		lipct.SetOff(true)
		lipp := net.LayerByName("LIPP")
		lipp.SetOff(true)
	*/

	full := prjn.NewFull()
	pone2one := prjn.NewPoolOneToOne()
	one2one := prjn.NewOneToOne()
	sameu := prjn.NewPoolSameUnit()
	sameu.SelfCon = false
	_ = one2one
	rndcut := prjn.NewUnifRnd()
	rndcut.PCon = 0.1
	_ = rndcut

	// basic super cons
	net.ConnectLayers(v1m, v2, ss.Prjn3x3Skp1, emer.Forward).SetClass("V1V2")
	net.ConnectLayers(v1h, v2, ss.Prjn4x4Skp2, emer.Forward).SetClass("V1V2")

	_, v4v2 := net.BidirConnectLayers(v2, v4, ss.Prjn4x4Skp2)
	v4v2.SetPattern(ss.Prjn4x4Skp2Recip)

	_, v3v2 := net.BidirConnectLayers(v2, v3, ss.Prjn4x4Skp2)
	v3v2.SetClass("BackMax") // "BackMax") // this is critical!
	v3v2.SetPattern(ss.Prjn4x4Skp2Recip)

	_, dpv3 := net.BidirConnectLayers(v3, dp, full)
	dpv3.SetClass("BackStrong") // likely key (in 233) -- retest

	_, teov4 := net.BidirConnectLayers(v4, teo, ss.Prjn3x3Skp1) // 3x3 > full
	teov4.SetClass("BackStrong")                                // todo: test

	_, teteo := net.BidirConnectLayers(teo, te, ss.Prjn4x4Skp2) // 4x4 > full
	teteo.SetPattern(ss.Prjn4x4Skp2Recip)

	// non-basic cons

	////////////////////
	// to LIP -- weak from v2, v3

	net.ConnectLayers(v2, lip, ss.Prjn2x2Skp2Recip, emer.Forward).SetClass("FwdWeak")
	net.ConnectLayers(v3, lip, ss.Prjn4x4Skp4Recip, emer.Forward).SetClass("FwdWeak")

	net.ConnectLayers(v2ct, lipct, ss.Prjn2x2Skp2Recip, emer.Forward).SetClass("FwdWeak")
	net.ConnectLayers(v3ct, lipct, ss.Prjn4x4Skp4Recip, emer.Forward).SetClass("FwdWeak")

	// Pulvinar connections
	net.ConnectLayers(v2ct, v1mp, pone2one, emer.Back).SetClass("BackToPulv1")
	net.ConnectLayers(v3ct, v1mp, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("BackToPulv2")
	net.ConnectLayers(v4ct, v1mp, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("BackToPulv2")
	net.ConnectLayers(teoct, v1mp, full, emer.Back).SetClass("BackToPulv") // orig is scheduled

	net.ConnectLayers(v2ct, v1hp, ss.Prjn2x2Skp2Recip, emer.Back).SetClass("BackToPulv1")
	net.ConnectLayers(v3ct, v1hp, ss.Prjn8x8Skp4Recip, emer.Back).SetClass("BackToPulv2")
	net.ConnectLayers(v4ct, v1hp, ss.Prjn8x8Skp4Recip, emer.Back).SetClass("BackToPulv2")
	net.ConnectLayers(teoct, v1hp, full, emer.Back).SetClass("BackToPulv") // orig is scheduled

	/*
		// note: v3ct -> v3p is automatic, with one-to-one -- not in standard one!
		v3p.RecvPrjns().SendName(v3ct.Name()).SetOff(true)
		net.ConnectLayers(v2ct, v3p, ss.Prjn4x4Skp2, emer.Back).SetClass("BackToPulv5") // actually FF
		net.ConnectLayers(dpct, v3p, full, emer.Back).SetClass("BackToPulv2")
		net.ConnectLayers(teoct, v3p, full, emer.Back).SetClass("BackToPulv") // was uniform rnd but p = 1

		dpp.RecvPrjns().SendName(dpct.Name()).SetPattern(full)
		dpp.RecvPrjns().SendName(dpct.Name()).SetClass("BackToPulv2")
		net.ConnectLayers(v2ct, dpp, full, emer.Back).SetClass("BackToPulv2")  // actually FF
		net.ConnectLayers(v3ct, dpp, full, emer.Back).SetClass("BackToPulv5")  // actually FF
		net.ConnectLayers(teoct, dpp, full, emer.Back).SetClass("BackToPulv2") // was uniform rnd but p = 1

		v4p.RecvPrjns().SendName(v4ct.Name()).SetClass("BackToPulv2")                   // gp1to1
		net.ConnectLayers(v2ct, v4p, ss.Prjn4x4Skp2, emer.Back).SetClass("BackToPulv5") // actually FF
		net.ConnectLayers(v3ct, v4p, ss.Prjn3x3Skp1, emer.Back).SetClass("BackToPulv5") // actually FF
		net.ConnectLayers(teoct, v4p, full, emer.Back).SetClass("BackToPulv2")          // was uniform rnd but p = 1

		teop.RecvPrjns().SendName(teoct.Name()).SetClass("BackToPulv2")        // gp1to1
		net.ConnectLayers(v3ct, teop, full, emer.Back).SetClass("BackToPulv2") // actually FF
		net.ConnectLayers(v4ct, teop, full, emer.Back).SetClass("BackToPulv5") // actually FF
		net.ConnectLayers(tect, teop, full, emer.Back).SetClass("BackToPulv5") // was uniform rnd but p = 1

		tep.RecvPrjns().SendName(tect.Name()).SetClass("BackToPulv2")          // gp1to1
		net.ConnectLayers(v4ct, tep, full, emer.Back).SetClass("BackToPulv5")  // actually FF
		net.ConnectLayers(teoct, tep, full, emer.Back).SetClass("BackToPulv2") // was uniform rnd but p = 1
	*/

	////////////////////
	// to V2

	// net.ConnectLayers(v2, v2, sameu, emer.Lateral)

	net.ConnectLayers(v1mp, v2, pone2one, emer.Forward).SetClass("FmPulv02")
	net.ConnectLayers(v1hp, v2, ss.Prjn2x2Skp2, emer.Forward).SetClass("FmPulv02")

	net.ConnectCtxtToCT(v2ct, v2ct, ss.Prjn3x3Skp1).SetClass("CTSelfLower") // was pone2one
	v2ct.RecvPrjns().SendName(v2.Name()).SetPattern(ss.Prjn3x3Skp1).SetClass("CTFmSuperLower")

	net.ConnectLayers(lip, v2, ss.Prjn2x2Skp2, emer.Back).SetClass("BackMax FmLIP") // key top-down attn .5 > .2
	net.ConnectLayers(teoct, v2, ss.Prjn4x4Skp2Recip, emer.Back)                    // key! .1 def

	// net.ConnectLayers(teo, v2, ss.Prjn4x4Skp2Recip, emer.Back) // too strong of top-down

	// v2ct
	net.ConnectLayers(v1mp, v2ct, ss.Prjn3x3Skp1, emer.Forward).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, v2ct, ss.Prjn4x4Skp2, emer.Forward).SetClass("FmPulv2")

	net.ConnectLayers(lipct, v2ct, ss.Prjn2x2Skp2, emer.Back).SetClass("CTBackMax FmLIP")
	net.ConnectLayers(v3ct, v2ct, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("CTBackMax")
	net.ConnectLayers(v4ct, v2ct, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("CTBackMax")

	// net.ConnectLayers(teoct, v2ct, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("CTBackMax") // not beneficial

	net.ConnectLayers(v3, v2ct, ss.Prjn2x2Skp2Recip, emer.Back).SetClass("SToCTMax")  // s -> ct leak
	net.ConnectLayers(teo, v2ct, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("SToCTMax") // s -> ct leak -- key @ max

	// CTBack generically worse, generally important for cosdiff
	// net.ConnectLayers(v3ct, v2p, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("BackToPulv")
	// net.ConnectLayers(v4ct, v2p, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("BackToPulv") // better without?  not clear

	////////////////////
	// to V3

	// net.ConnectLayers(v3, v3, sameu, emer.Lateral)

	net.ConnectLayers(v1mp, v3, ss.Prjn4x4Skp2, emer.Back).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, v3, ss.Prjn8x8Skp4, emer.Back).SetClass("FmPulv2")
	// net.ConnectLayers(dpp, v3, full, emer.Back).SetClass("FmPulv05") // todo: remove?

	net.ConnectCtxtToCT(v3ct, v3ct, ss.Prjn3x3Skp1).SetClass("CTSelfLower") // was pone2one
	v3ct.RecvPrjns().SendName(v3.Name()).SetPattern(ss.Prjn3x3Skp1).SetClass("CTFmSuperLower")

	net.ConnectLayers(v4, v3, ss.Prjn3x3Skp1, emer.Back).SetClass("BackStrong")
	net.ConnectLayers(lip, v3, ss.Prjn2x2Skp2, emer.Back).SetClass("FmLIP")

	net.ConnectLayers(teo, v3, ss.Prjn3x3Skp1, emer.Back)
	net.ConnectLayers(teoct, v3, ss.Prjn3x3Skp1, emer.Back)

	// v3ct

	net.ConnectLayers(v1mp, v3ct, ss.Prjn4x4Skp2, emer.Back).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, v3ct, ss.Prjn8x8Skp4, emer.Back).SetClass("FmPulv2")
	// net.ConnectLayers(dpp, v3ct, full, emer.Back).SetClass("FmPulv2")

	net.ConnectLayers(lipct, v3ct, ss.Prjn2x2Skp2, emer.Back).SetClass("CTBack FmLIP")
	net.ConnectLayers(dpct, v3ct, full, emer.Back).SetClass("CTBack")
	net.ConnectLayers(v4ct, v3ct, ss.Prjn3x3Skp1, emer.Back).SetClass("CTBack")

	// todo: retest again:
	net.ConnectLayers(dp, v3ct, full, emer.Back).SetClass("SToCT")
	net.ConnectLayers(v4, v3ct, ss.Prjn3x3Skp1, emer.Back).SetClass("SToCT") // s -> ct, 3x3 ok

	// net.ConnectLayers(dpct, v3p, full, emer.Back).SetClass("BackToPulv") // not much effect on cosdiff
	// net.ConnectLayers(v2ct, v3p, ss.Prjn4x4Skp2, emer.Forward).SetClass("FwdToPulv") // has major effect on cosdiff

	////////////////////
	// to DP

	net.ConnectLayers(v1mp, dp, full, emer.Back).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, dp, full, emer.Back).SetClass("FmPulv2")
	// net.ConnectLayers(v3p, dp, full, emer.Back).SetClass("FmPulv")
	// net.ConnectLayers(teop, dp, full, emer.Back).SetClass("FmPulv") // todo: test

	net.ConnectCtxtToCT(dpct, dpct, full).SetClass("CTSelfLower") // not much effect, but consistent

	// net.ConnectLayers(v2, dp, full, emer.Forward) // no effect, expensive

	net.ConnectLayers(teo, dp, full, emer.Back) // todo: test again

	// dpct
	net.ConnectLayers(v1mp, dpct, full, emer.Back).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, dpct, full, emer.Back).SetClass("FmPulv2")

	net.ConnectLayers(teoct, dpct, full, emer.Back).SetClass("CTBack")

	////////////////////
	// to V4

	// net.ConnectLayers(v4, v4, sameu, emer.Lateral)

	net.ConnectLayers(v1mp, v4, ss.Prjn4x4Skp2, emer.Back).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, v4, ss.Prjn8x8Skp4, emer.Back).SetClass("FmPulv2")

	v4ct.RecvPrjns().SendName(v4.Name()).SetPattern(ss.Prjn3x3Skp1).SetClass("CTFmSuper")
	net.ConnectCtxtToCT(v4ct, v4ct, ss.Prjn3x3Skp1).SetClass("CTSelfLower") // was pone2one

	// net.ConnectLayers(teoct, v4, ss.Prjn3x3Skp1, emer.Back).SetClass("CTBack") // very not beneficial

	// Prjn4x4Skp2Recip is same as full, but has topo scales -- better than full
	net.ConnectLayers(te, v4, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("BackStrong")

	// v4ct
	net.ConnectLayers(v1mp, v4ct, ss.Prjn4x4Skp2, emer.Back).SetClass("FmPulv2")
	net.ConnectLayers(v1hp, v4ct, ss.Prjn8x8Skp4, emer.Back).SetClass("FmPulv2")

	net.ConnectLayers(teoct, v4ct, ss.Prjn3x3Skp1, emer.Back).SetClass("CTBack")
	net.ConnectLayers(teo, v4ct, ss.Prjn3x3Skp1, emer.Back).SetClass("SToCT") // s -> ct -- important

	// Prjn4x4Skp2Recip is same as full, but has topo scales -- better
	net.ConnectLayers(tect, v4ct, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("CTBack")

	// net.ConnectLayers(v2ct, v4ct, ss.Prjn4x4Skp2, emer.Forward).SetClass("CTBack") // instead of direct to v2p -- not helpful

	// net.ConnectLayers(teoct, v4p, ss.Prjn3x3Skp1, emer.Back) // not much additional benefit for cosdiff

	// net.ConnectLayers(v2ct, v4p, ss.Prjn4x4Skp2, emer.Forward).SetClass("FwdToPulv") // has major effect on cosdiff

	////////////////////
	// to TEO

	// net.ConnectLayers(teo, teo, sameu, emer.Lateral)

	net.ConnectLayers(v1mp, teo, full, emer.Back).SetClass("FmPulv")
	net.ConnectLayers(v1hp, teo, full, emer.Back).SetClass("FmPulv")

	// teoct
	net.ConnectLayers(v1mp, teoct, full, emer.Back).SetClass("FmPulv")
	net.ConnectLayers(v1hp, teoct, full, emer.Back).SetClass("FmPulv")

	teoct.RecvPrjns().SendName(teo.Name()).SetPattern(ss.Prjn3x3Skp1).SetClass("CTFmSuper")
	net.ConnectCtxtToCT(teoct, teoct, pone2one).SetClass("CTSelfHigher") // pone2one similar to 3x3 -- bit better

	net.ConnectLayers(tect, teoct, ss.Prjn4x4Skp2Recip, emer.Back).SetClass("CTBack") // CTBack > not

	net.ConnectLayers(v4ct, teoct, full, emer.Forward).SetClass("CTBack") // instead of direct to v2p

	// todo: test topo on both
	// net.ConnectLayers(v4ct, teop, full, emer.Forward).SetClass("FwdToPulv") // sig effect on TEOP cosdiff, but improves TEP
	// net.ConnectLayers(tect, teop, full, emer.Back).SetClass("BackToPulv") // no effect on cosdiff, but better Cat without

	////////////////////
	// to TE

	// net.ConnectLayers(te, te, sameu, emer.Lateral)

	net.ConnectLayers(v1mp, te, full, emer.Back).SetClass("FmPulv")
	net.ConnectLayers(v1hp, te, full, emer.Back).SetClass("FmPulv")

	tect.RecvPrjns().SendName(te.Name()).SetPattern(ss.Prjn3x3Skp1).SetClass("CTFmSuper")
	net.ConnectCtxtToCT(tect, tect, pone2one).SetClass("CTSelfHigher") // pone2one > full

	net.ConnectLayers(v1mp, tect, full, emer.Back).SetClass("FmPulv")
	net.ConnectLayers(v1hp, tect, full, emer.Back).SetClass("FmPulv")

	net.ConnectLayers(teoct, tect, ss.Prjn4x4Skp2, emer.Forward).SetClass("CTBack") // was FwdWeak

	// net.ConnectLayers(teoct, tep, full, emer.Back).SetClass("FwdToPulv") // sig effect on cosdiff, not much other eff

	////////////////////
	// Latinhib

	// this extra inhibition drives decorrelation, produces significant learning benefits
	net.LateralConnectLayerPrjn(v2, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(v2ct, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(v3, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(v3ct, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(dp, full, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(dpct, full, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(v4, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(v4ct, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(teo, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(teoct, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(te, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)
	net.LateralConnectLayerPrjn(tect, pone2one, &axon.HebbPrjn{}).SetType(emer.Inhib)

	////////////////////
	// Shortcuts

	// V1 shortcuts best for syncing all layers -- like the pulvinar basically
	net.ConnectLayers(v1m, v3, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, dp, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, v4, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, teo, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, te, rndcut, emer.Forward).SetClass("V1SC")

	net.ConnectLayers(v1m, v3ct, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, dpct, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, v4ct, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, teoct, rndcut, emer.Forward).SetClass("V1SC")
	net.ConnectLayers(v1m, tect, rndcut, emer.Forward).SetClass("V1SC")

	////////////////////
	// Position

	v1mp.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: v1m.Name(), XAlign: relpos.Left, Space: 10})
	v1hp.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v1h.Name(), YAlign: relpos.Front, Space: 2})

	v2.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: v1m.Name(), XAlign: relpos.Left, YAlign: relpos.Front})
	lip.SetRelPos(relpos.Rel{Rel: relpos.Above, Other: v2.Name(), XAlign: relpos.Left, YAlign: relpos.Front})
	// v2p.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: v1m.Name(), XAlign: relpos.Left, Space: 10})
	v2ct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: v2.Name(), XAlign: relpos.Left, Space: 10})

	v3.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v2.Name(), YAlign: relpos.Front, Space: 2})
	v3ct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: v3.Name(), XAlign: relpos.Left, Space: 10})
	// v3p.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v3ct.Name(), YAlign: relpos.Front, Space: 2})

	dp.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v3.Name(), YAlign: relpos.Front, Space: 2})
	dpct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: dp.Name(), XAlign: relpos.Left, Space: 10})
	// dpp.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: dpct.Name(), YAlign: relpos.Front, Space: 2})

	v4.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: v3ct.Name(), XAlign: relpos.Left, Space: 10})
	v4ct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: v4.Name(), XAlign: relpos.Left, Space: 10})
	// v4p.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: v4ct.Name(), YAlign: relpos.Back, Space: 2})

	teo.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: eyepos.Name(), YAlign: relpos.Front, Space: 2})
	teoct.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: teo.Name(), XAlign: relpos.Left, Space: 10})
	// teop.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: teoct.Name(), XAlign: relpos.Left, Space: 10})

	te.SetRelPos(relpos.Rel{Rel: relpos.RightOf, Other: teo.Name(), YAlign: relpos.Front, Space: 2})
	tect.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: te.Name(), XAlign: relpos.Left, Space: 10})
	// tep.SetRelPos(relpos.Rel{Rel: relpos.Behind, Other: tect.Name(), XAlign: relpos.Left, Space: 10})

	////////////////////

	// net.LockThreads = true // makes no difference
	runtime.GOMAXPROCS(8) // makes no diff: otherwise gets it from slurm request and it is too small

	/*
		// 4 threads = about 500 msec / trl @8 mpi
		v2.SetThread(1)
		v2ct.SetThread(1)
		v2p.SetThread(1)

		dp.SetThread(1)
		dpct.SetThread(1)
		dpp.SetThread(1)

		v3ct.SetThread(1)

		v3p.SetThread(2)
		v3.SetThread(2)

		v4.SetThread(3)
		v4ct.SetThread(2)
		v4p.SetThread(2)

		teo.SetThread(3) // 23 M -- by far biggest

		teoct.SetThread(0) // 19 M
		teop.SetThread(0)

		te.SetThread(2)

		tect.SetThread(0)
		tep.SetThread(0)
	*/

	//	2 threads = only slight advantage over 1 thread
	v2.SetThread(0)
	v2ct.SetThread(0)
	// v2p.SetThread(0)

	dp.SetThread(0)
	dpct.SetThread(0)
	// dpp.SetThread(0)

	v3ct.SetThread(0)

	// v3p.SetThread(1)
	v3.SetThread(1)

	v4.SetThread(1)
	v4ct.SetThread(1)
	// v4p.SetThread(1)

	teo.SetThread(1) // 23 M -- by far biggest

	teoct.SetThread(0) // 19 M
	// teop.SetThread(0)

	te.SetThread(1)

	tect.SetThread(0)
	// tep.SetThread(0)
}

func (ss *Sim) InitWts(net *deep.Network) {
	if len(net.Layers) == 0 {
		return
	}
	net.InitWts()
	if !ss.LIPOnly {
		mpi.Printf("loading lip_pretrained.wts.gz...\n")
		net.OpenWtsJSON(gi.FileName("lip_pretrained.wts.gz"))
	}
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
		return fmt.Sprintf("Run:\t%d\tEpoch:\t%d\tTrial:\t%d\tCycle:\t%d\tName:\t%v\t\t\t", ss.TrainEnv.Run.Cur, ss.TrainEnv.Epoch.Cur, ss.TrainEnv.Trial.Cur, ss.Time.Cycle, ss.TrainEnv.String())
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

	lays := []string{"V1m", "V1h", "EyePos", "SacPlan", "Saccade", "ObjVel"}
	for _, lnm := range lays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		pats := en.State(ly.Nm)
		if pats != nil {
			ly.ApplyExt(pats)
		}
	}
}

// TrainTrial runs one trial of training using TrainEnv
func (ss *Sim) TrainTrial() {

	if ss.NeedsNewRun {
		ss.NewRun()
	}

	ss.TrainEnv.Step() // the Env encapsulates and manages all counter state

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
	if ss.RepsInterval > 0 && epc%ss.RepsInterval == 0 {
		ss.LogTrnRepTrl(ss.TrnTrlRepLog)
	}
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

		if !ss.LIPOnly {
			fnm := ss.LogFileName("catact")
			fmt.Printf("Saving CatLayActs to: %v\n", fnm)
			ss.CatLayActs.SaveCSV(gi.FileName(fnm), etable.Tab, etable.Headers)
		}
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
	ss.SuperLays = []string{"V1m"}
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

	ss.RSA.Init(ss.SuperLays)
	ss.RSA.SetCats(ss.TrainEnv.Objs)
}

// TrialStats computes the trial-level statistics.
func (ss *Sim) TrialStats(train bool) {
	for li, lnm := range ss.PulvLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		ss.PulvCosDiff[li] = float64(ly.CosDiff.Cos)
		ss.PulvUnitErr[li] = ly.PctUnitErr()
		if (ss.LIPOnly && lnm == "LIPP") || (!ss.LIPOnly && lnm == "V2P") {
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
	dt.SetCellString("Obj", row, ss.TrainEnv.CurCat)
	dt.SetCellString("TrialName", row, ss.TrainEnv.String())

	for li, lnm := range ss.PulvLays {
		dt.SetCellFloat(lnm+"_CosDiff", row, ss.PulvCosDiff[li])
		dt.SetCellFloat(lnm+"_TrlCosDiff", row, ss.PulvTrlCosDiff[li])
		dt.SetCellFloat(lnm+"_UnitErr", row, ss.PulvUnitErr[li])
	}
	for li, lnm := range ss.HidLays {
		dt.SetCellFloat(lnm+"_TrlCosDiff", row, ss.HidTrlCosDiff[li])
	}

	ss.RecCatLayActs(ss.CatLayActs)

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
	plt.Params.Title = "What-Where-Integration 3DObj Train Trial Plot"
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
//  CatLayActs

func (ss *Sim) RecCatLayActs(dt *etable.Table) {
	obj := ss.TrainEnv.CurObj
	rows := dt.RowsByString("Obj", obj, etable.Equals, etable.UseCase)
	if len(rows) != ss.MaxTicks {
		log.Printf("RecCatLayActs: error: object not found: %s\n", obj)
		return
	}
	row := rows[0] + ss.TrainEnv.Tick.Cur
	avgDt := float32(0.1)
	avgDtC := 1 - avgDt
	for _, lnm := range ss.SuperLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		cv := dt.CellTensor(lnm, row).(*etensor.Float32)
		for i := range ly.Neurons {
			cv.Values[i] = avgDtC*cv.Values[i] + avgDt*ly.Neurons[i].ActM
		}
	}
}

// ShareCatLayActs shares CatLayActs table across processors, for MPI mode
func (ss *Sim) ShareCatLayActs() {
	if ss.LIPOnly || !ss.UseMPI {
		return
	}
	np := float32(1) / float32(mpi.WorldSize())
	empi.ReduceTable(ss.CatLayActsDest, ss.CatLayActs, ss.Comm, mpi.OpSum)
	for ci, dcoli := range ss.CatLayActs.Cols {
		if dcoli.DataType() != etensor.FLOAT32 {
			continue
		}
		dcol := dcoli.(*etensor.Float32)
		scol := ss.CatLayActsDest.Cols[ci].(*etensor.Float32)
		for i := range dcol.Values {
			dcol.Values[i] = np * scol.Values[i]
		}
	}
}

func (ss *Sim) ConfigCatLayActs(dt *etable.Table) {
	dt.SetMetaData("name", "CatLayActs")
	dt.SetMetaData("desc", "layer activations for each cat / obj")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Cat", etensor.STRING, nil, nil},
		{"Obj", etensor.STRING, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
	}
	for _, lnm := range ss.SuperLays {
		ly := ss.Net.LayerByName(lnm).(axon.AxonLayer).AsAxon()
		sch = append(sch, etable.Column{lnm, etensor.FLOAT32, ly.Shp.Shp, ly.Shp.Nms})
	}

	nobj := len(ss.TrainEnv.Objs)
	dt.SetFromSchema(sch, nobj*ss.MaxTicks)
	row := 0
	for _, ob := range ss.TrainEnv.Objs {
		co := strings.Split(ob, "/")
		for t := 0; t < ss.MaxTicks; t++ {
			dt.SetCellString("Cat", row, co[0])
			dt.SetCellString("Obj", row, co[1])
			dt.SetCellFloat("Tick", row, float64(t))
			row++
		}
	}
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
	dt.SetCellString("Obj", row, ss.TrainEnv.CurCat)
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
		ss.ShareCatLayActs()
	}

	epc := ss.TrainEnv.Epoch.Prv // this is triggered by increment so use previous value
	nt := float64(trl.Rows)

	if !ss.LIPOnly && mpi.WorldRank() == 0 {
		if (epc % ss.RSA.Interval) == 0 {
			ss.RSA.StatsFmActs(ss.CatLayActs, ss.SuperLays)
			fnm := ss.LogFileName("TEsim")
			fmt.Printf("Saving TEsim to: %v\n", fnm)
			sm := ss.RSA.Sims["TE"]
			etensor.SaveCSV(sm.Mat, gi.FileName(fnm), etable.Tab.Rune())
		}
		for li, lnm := range ss.SuperLays {
			dt.SetCellFloat(lnm+"_V1Sim", row, ss.RSA.V1Sims[li])
			dt.SetCellFloat(lnm+"_CatDst", row, ss.RSA.CatDists[li])
		}
		pr := 0.0
		teidx := len(ss.SuperLays) - 1
		if ss.RSA.PermDists["TE"] > 0 {
			pr = ss.RSA.CatDists[teidx] / ss.RSA.PermDists["TE"]
		}
		dt.SetCellFloat("TE_PermRatio", row, pr)
		dt.SetCellFloat("TE_PermDst", row, ss.RSA.PermDists["TE"])
		dt.SetCellFloat("TE_PermNCat", row, float64(ss.RSA.PermNCats["TE"]))
		dt.SetCellFloat("TE_BasicDst", row, ss.RSA.BasicDists[teidx])
		dt.SetCellFloat("TE_ExptDst", row, ss.RSA.ExptDists[teidx])
	}

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
	for _, lnm := range ss.SuperLays {
		sch = append(sch, etable.Column{lnm + "_V1Sim", etensor.FLOAT64, nil, nil})
	}
	for _, lnm := range ss.SuperLays {
		sch = append(sch, etable.Column{lnm + "_CatDst", etensor.FLOAT64, nil, nil})
	}
	sch = append(sch, etable.Column{"TE_PermRatio", etensor.FLOAT64, nil, nil})
	sch = append(sch, etable.Column{"TE_PermDst", etensor.FLOAT64, nil, nil})
	sch = append(sch, etable.Column{"TE_PermNCat", etensor.FLOAT64, nil, nil})
	sch = append(sch, etable.Column{"TE_BasicDst", etensor.FLOAT64, nil, nil})
	sch = append(sch, etable.Column{"TE_ExptDst", etensor.FLOAT64, nil, nil})

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
	plt.Params.Title = "What-Where-Integration 3DObj Epoch Plot"
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
	for _, lnm := range ss.SuperLays {
		on := lnm == "TE"
		plt.SetColParams(lnm+"_V1Sim", on, eplot.FixMin, 0, eplot.FixMax, 1)
		plt.SetColParams(lnm+"_CatDst", on, eplot.FixMin, 0, eplot.FixMax, 1)
	}
	plt.SetColParams("TE_PermRatio", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("TE_PermDst", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("TE_PermNCat", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("TE_BasicDst", eplot.On, eplot.FixMin, 0, eplot.FixMax, 1)
	plt.SetColParams("TE_ExptDst", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 1)

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

// OpenCatActs Open a catact file with layer activity by category,
// and then run the RSA analysis on it -- see RSA for results
func (ss *Sim) OpenCatActs(fname gi.FileName) {
	ss.CatLayActs.OpenCSV(fname, etable.Tab)
	ss.RSA.StatsFmActs(ss.CatLayActs, ss.SuperLays)
}

// OpenSimMat Open a TEsim TE similarity matrix in standard object order
func (ss *Sim) OpenSimMat(fname gi.FileName) {
	ss.RSA.OpenSimMat("TE", fname)
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
	dt.SetCellString("Obj", row, ss.TestEnv.CurCat)
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
	plt.Params.Title = "What-Where-Integration 3DObj Test Trial Plot"
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
	plt.Params.Title = "What-Where-Integration 3DObj Testing Epoch Plot"
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
	plt.Params.Title = "What-Where-Integration 3DObj Run Plot"
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
	cam.Pose.Pos.Set(0.0, 1.3, 2.56)
	cam.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})
	// cam.Pose.Quat.SetFromAxisAngle(mat32.Vec3{-1, 0, 0}, 0.4077744)
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("wwi3d")
	gi.SetAppAbout(`wwi3d does deep predictive learning of 3D objects tumbling through space, with periodic saccadic eye movements, providing plenty of opportunity for prediction errors. wwi = what, where integration: both pathways combine to predict object -- *where* (dorsal) pathway is trained first and residual prediction error trains *what* pathway. See <a href="https://github.com/ccnlab/deep-obj-cat/blob/master/sims/wwi3d/README.md">README.md on GitHub</a>.</p>`)

	win := gi.NewMainWindow("wwi3d", "WWI 3D", width, height)
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
	tg.SetTensor(&ss.TrainEnv.V1Hi.ImgTsr)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "TstTrlPlot").(*eplot.Plot2D)
	ss.TstTrlPlot = ss.ConfigTstTrlPlot(plt, ss.TstTrlLog)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "TstEpcPlot").(*eplot.Plot2D)
	ss.TstEpcPlot = ss.ConfigTstEpcPlot(plt, ss.TstEpcLog)

	plt = tv.AddNewTab(eplot.KiT_Plot2D, "RunPlot").(*eplot.Plot2D)
	ss.RunPlot = ss.ConfigRunPlot(plt, ss.RunLog)

	ss.ActRFGrids = make(map[string]*etview.TensorGrid)
	for _, nm := range ss.ActRFNms {
		tg := tv.AddNewTab(etview.KiT_TensorGrid, nm).(*etview.TensorGrid)
		tg.SetStretchMax()
		ss.ActRFGrids[nm] = tg
	}

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

	tbar.AddAction(gi.ActOpts{Label: "Open SimMat", Icon: "file-open", Tooltip: "Open a TEsim RSA similarity matrix (in standard object order of rows, not sorted by anything"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(ss, "OpenSimMat", vp)
		})

	tbar.AddAction(gi.ActOpts{Label: "Open CatActs", Icon: "file-open", Tooltip: "Open a catact file with layer activity by category, and then run the RSA analysis on it -- see RSA for results"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(ss, "OpenCatActs", vp)
		})

	tbar.AddSeparator("misc")

	tbar.AddAction(gi.ActOpts{Label: "New Seed", Icon: "new", Tooltip: "Generate a new initial random seed to get different results.  By default, Init re-establishes the same initial seed every time."}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.NewRndSeed()
		})

	tbar.AddAction(gi.ActOpts{Label: "README", Icon: "file-markdown", Tooltip: "Opens your browser on the README file that contains instructions for how to run this model."}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			gi.OpenURL("https://github.com/ccnlab/deep-obj-cat/blob/master/sims/wwi3d/README.md")
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
