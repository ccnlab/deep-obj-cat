// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/evec"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/goki/mat32"
)

// SacEnv implements saccading logic for generating visual saccades
// around a 2D World plane, with a moving object that must remain
// in view.  Generates the track of the object.
// World size is defined as -1..1 in normalized units.
type SacEnv struct {
	Nm           string       `desc:"name of this environment"`
	Dsc          string       `desc:"description of this environment"`
	TrajLenRange minmax.Int   `desc:"range of trajectory lengths (time steps)"`
	FixDurRange  minmax.Int   `desc:"range of fixation durations"`
	SacGenMax    float32      `desc:"maximum saccade size"`
	VelGenMax    float32      `desc:"maximum object velocity"`
	ZeroVelP     float64      `desc:"probability of zero velocity object motion as a discrete option prior to computing random velocity"`
	Margin       float32      `desc:"edge around World to not look past"`
	ViewPct      float32      `desc:"size of view as proportion of -1..1 world size"`
	WorldVisSz   evec.Vec2i   `desc:"visualization size of world -- for debug visualization"`
	ViewVisSz    evec.Vec2i   `desc:"visualization size of view -- for debug visualization"`
	AddRows      bool         `desc:"add rows to Table for each step (for debugging) -- else re-use 0"`
	V1Pop        popcode.TwoD `desc:"2d population code for gaussian bump rendering of v1 obj position"`
	EyePop       popcode.TwoD `desc:"2d population code for gaussian bump rendering of eye position"`
	SacPop       popcode.TwoD `desc:"2d population code for gaussian bump rendering of saccade plan / execution"`
	ObjVelPop    popcode.TwoD `desc:"2d population code for gaussian bump rendering of object velocity"`

	// State below here
	Table       etable.Table    `desc:"table showing visualization of state"`
	WorldTsr    etensor.Float32 `desc:"tensor state showing world position of obj"`
	ViewTsr     etensor.Float32 `desc:"tensor state showing view position of obj"`
	V1Tsr       etensor.Float32 `desc:"pop-code rendered view position"`
	EyePosTsr   etensor.Float32 `desc:"eye position popcode"`
	SacPlanTsr  etensor.Float32 `desc:"saccade plan popcode"`
	SaccadeTsr  etensor.Float32 `desc:"saccade popcode "`
	ObjVelTsr   etensor.Float32 `desc:"object velocity"`
	TrajLen     int             `inactive:"+" desc:"current trajectory length"`
	FixDur      int             `inactive:"+" desc:"current fixation duration"`
	Run         env.Ctr         `view:"inline" desc:"current run of model as provided during Init"`
	Epoch       env.Ctr         `view:"inline" desc:"arbitrary aggregation of trials, for stats etc"`
	Trial       env.Ctr         `view:"inline" desc:"each object trajectory is one trial"`
	Tick        env.Ctr         `inactive:"+" desc:"tick counter within trajectory, counts up from 0..TrajLen-1"`
	SacTick     env.Ctr         `inactive:"+" desc:"tick counter within current fixation"`
	World       minmax.F32      `inactive:"+" desc:"World minus margin"`
	View        minmax.F32      `inactive:"+" desc:"View minus margin"`
	ObjPos      mat32.Vec2      `inactive:"+" desc:"object position, in world coordinates"`
	ObjViewPos  mat32.Vec2      `inactive:"+" desc:"object position, in view coordinates"`
	ObjVel      mat32.Vec2      `inactive:"+" desc:"object velocity, in world coordinates"`
	ObjPosNext  mat32.Vec2      `inactive:"+" desc:"next object position, in world coordinates"`
	ObjVelNext  mat32.Vec2      `inactive:"+" desc:"next object velocity, in world coordinates"`
	EyePos      mat32.Vec2      `inactive:"+" desc:"eye position, in world coordinates"`
	SacPlan     mat32.Vec2      `inactive:"+" desc:"eye movement plan, in world coordinates"`
	Saccade     mat32.Vec2      `inactive:"+" desc:"current trial eye movement, in world coordinates"`
	NewTraj     bool            `inactive:"+" desc:"true if new trajectory started on this trial"`
	NewSac      bool            `inactive:"+" desc:"true if new saccade was made on this trial"`
	NewTrajNext bool            `inactive:"+" desc:"true if next trial will be a new trajectory"`
}

func (sc *SacEnv) Name() string { return sc.Nm }
func (sc *SacEnv) Desc() string { return sc.Dsc }

// Defaults sets generic defaults -- use ParamSet to override
func (sc *SacEnv) Defaults() {
	sc.TrajLenRange.Set(8, 8)
	sc.FixDurRange.Set(2, 2)
	sc.SacGenMax = 0.4
	sc.VelGenMax = 0 // 0.4
	sc.ZeroVelP = 0
	sc.Margin = 0.05
	sc.ViewPct = 0.5
	sc.WorldVisSz.Set(24, 24)
	sc.ViewVisSz.Set(16, 16)

	sc.ConfigTable(&sc.Table)
	yx := []string{"Y", "X"}
	sc.WorldTsr.SetShape([]int{sc.WorldVisSz.Y, sc.WorldVisSz.X}, nil, yx)
	sc.ViewTsr.SetShape([]int{sc.ViewVisSz.Y, sc.ViewVisSz.X}, nil, yx)

	sc.V1Pop.Defaults()
	sc.V1Pop.Min.Set(-0.9, -0.9)
	sc.V1Pop.Max.Set(0.9, 0.9)
	sc.V1Pop.Sigma.Set(0.2, 0.2) // 0.1

	sc.V1Tsr.SetShape([]int{11, 11}, nil, yx)

	sc.EyePop.Defaults()
	sc.EyePop.Min.Set(-1.1, -1.1)
	sc.EyePop.Max.Set(1.1, 1.1)
	sc.EyePop.Sigma.Set(0.2, 0.2) // 0.1 orig

	sc.EyePosTsr.SetShape([]int{11, 11}, nil, yx)

	sc.SacPop.Defaults()
	sc.SacPop.Min.Set(-0.45, -0.45)
	sc.SacPop.Max.Set(0.45, 0.45)

	sc.SacPlanTsr.SetShape([]int{11, 11}, nil, yx)
	sc.SaccadeTsr.SetShape([]int{11, 11}, nil, yx)

	sc.ObjVelPop.Defaults()
	sc.ObjVelPop.Min.Set(-0.45, -0.45)
	sc.ObjVelPop.Max.Set(0.45, 0.45)

	sc.ObjVelTsr.SetShape([]int{11, 11}, nil, yx)
}

// Init must be called at start prior to generating saccades
func (sc *SacEnv) Init(run int) {
	sc.World.Max = 1 - sc.Margin
	sc.World.Min = -1 + sc.Margin
	sc.View.Max = sc.ViewPct - sc.Margin
	sc.View.Min = -sc.ViewPct + sc.Margin
	sc.Table.SetNumRows(1)
	sc.Run.Scale = env.Run
	sc.Epoch.Scale = env.Epoch
	sc.Trial.Scale = env.Trial
	sc.Tick.Scale = env.Tick
	sc.Tick.Max = sc.TrajLen
	sc.SacTick.Scale = env.Tick
	sc.SacTick.Max = sc.FixDur
	sc.SacTick.Cur = sc.SacTick.Max - 1 // ensure that we saccade next time
	sc.Run.Init()
	sc.Epoch.Init()
	sc.Trial.Init()
	sc.Tick.Cur = -1 // will increment to 0
	sc.Run.Cur = run
	sc.NextTraj() // start with a trajectory ready
}

func (sc *SacEnv) Validate() error {
	return nil
}

func (sc *SacEnv) ConfigTable(dt *etable.Table) {
	yx := []string{"Y", "X"}
	sch := etable.Schema{
		{"TrialName", etensor.STRING, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"SacTick", etensor.INT64, nil, nil},
		{"World", etensor.FLOAT32, []int{sc.WorldVisSz.Y, sc.WorldVisSz.X}, yx},
		{"View", etensor.FLOAT32, []int{sc.ViewVisSz.Y, sc.ViewVisSz.X}, yx},
		{"ObjPos", etensor.FLOAT32, []int{2}, nil},
		{"ObjViewPos", etensor.FLOAT32, []int{2}, nil},
		{"ObjVel", etensor.FLOAT32, []int{2}, nil},
		{"ObjPosNext", etensor.FLOAT32, []int{2}, nil},
		{"EyePos", etensor.FLOAT32, []int{2}, nil},
		{"SacPlan", etensor.FLOAT32, []int{2}, nil},
		{"Saccade", etensor.FLOAT32, []int{2}, nil},
	}
	dt.SetFromSchema(sch, 0)
}

func (sc *SacEnv) WriteToTable(dt *etable.Table) {
	row := 0
	if sc.AddRows {
		row = dt.Rows
	}
	dt.SetNumRows(row + 1)

	nm := fmt.Sprintf("t %d, s %d, x %+4.2f, y %+4.2f", sc.Tick.Cur, sc.SacTick.Cur, sc.ObjPos.X, sc.ObjPos.Y)

	dt.SetCellString("TrialName", row, nm)
	dt.SetCellFloat("Tick", row, float64(sc.Tick.Cur))
	dt.SetCellFloat("SacTick", row, float64(sc.SacTick.Cur))

	sc.WorldTsr.SetZeros()
	opx := int(math.Floor(float64(0.5 * (sc.ObjPos.X + 1) * float32(sc.WorldVisSz.X))))
	opy := int(math.Floor(float64(0.5 * (sc.ObjPos.Y + 1) * float32(sc.WorldVisSz.Y))))
	idx := []int{opy, opx}
	if sc.WorldTsr.IdxIsValid(idx) {
		sc.WorldTsr.SetFloat(idx, 1)
	} else {
		log.Printf("SacEnv: World index invalid: %v\n", idx)
	}

	sc.ViewTsr.SetZeros()
	opx = int(math.Floor(float64((0.5 * (sc.ObjViewPos.X + sc.ViewPct) / sc.ViewPct) * float32(sc.ViewVisSz.X))))
	opy = int(math.Floor(float64((0.5 * (sc.ObjViewPos.Y + sc.ViewPct) / sc.ViewPct) * float32(sc.ViewVisSz.Y))))
	idx = []int{opy, opx}
	if sc.ViewTsr.IdxIsValid(idx) {
		sc.ViewTsr.SetFloat(idx, 1)
	} else {
		log.Printf("SacEnv: View index invalid: %v\n", idx)
	}

	dt.SetCellTensor("World", row, &sc.WorldTsr)
	dt.SetCellTensor("View", row, &sc.ViewTsr)

	dt.SetCellTensorFloat1D("ObjPos", row, 0, float64(sc.ObjPos.X))
	dt.SetCellTensorFloat1D("ObjPos", row, 1, float64(sc.ObjPos.Y))
	dt.SetCellTensorFloat1D("ObjViewPos", row, 0, float64(sc.ObjViewPos.X))
	dt.SetCellTensorFloat1D("ObjViewPos", row, 1, float64(sc.ObjViewPos.Y))
	dt.SetCellTensorFloat1D("ObjVel", row, 0, float64(sc.ObjVel.X))
	dt.SetCellTensorFloat1D("ObjVel", row, 1, float64(sc.ObjVel.Y))
	dt.SetCellTensorFloat1D("ObjPosNext", row, 0, float64(sc.ObjPosNext.X))
	dt.SetCellTensorFloat1D("ObjPosNext", row, 1, float64(sc.ObjPosNext.Y))
	dt.SetCellTensorFloat1D("EyePos", row, 0, float64(sc.EyePos.X))
	dt.SetCellTensorFloat1D("EyePos", row, 1, float64(sc.EyePos.Y))
	dt.SetCellTensorFloat1D("SacPlan", row, 0, float64(sc.SacPlan.X))
	dt.SetCellTensorFloat1D("SacPlan", row, 1, float64(sc.SacPlan.Y))
	dt.SetCellTensorFloat1D("Saccade", row, 0, float64(sc.Saccade.X))
	dt.SetCellTensorFloat1D("Saccade", row, 1, float64(sc.Saccade.Y))
}

func (sc *SacEnv) LimitVel(vel, start, trials float32) float32 {
	if trials <= 0 {
		return vel
	}
	end := start + vel*trials
	if end > sc.World.Max {
		vel = (sc.World.Max - start) / trials
	} else if end < sc.World.Min {
		vel = (sc.World.Min - start) / trials
	}
	return vel
}

func (sc *SacEnv) LimitPos(pos, max float32) float32 {
	if pos > max {
		pos = max
	}
	if pos < -max {
		pos = -max
	}
	return pos
}

func (sc *SacEnv) LimitSac(sacDev, start, objPos, objVel, trials float32) float32 {
	objEnd := objPos + objVel*trials
	eyep := start + sacDev
	lowView := eyep + sc.View.Min
	highView := eyep + sc.View.Max
	// do obj_end first then pos so it has stronger constraint
	if objEnd < lowView {
		sacDev += (objEnd - lowView)
	} else if objEnd > highView {
		sacDev += objEnd - highView
	}
	eyep = start + sacDev
	if eyep < sc.World.Min {
		sacDev += sc.World.Min - eyep
	} else if eyep > sc.World.Max {
		sacDev += sc.World.Max - eyep
	}
	eyep = start + sacDev
	lowView = eyep + sc.View.Min
	highView = eyep + sc.View.Max
	if objPos < lowView {
		sacDev += objPos - lowView
	} else if objPos > highView {
		sacDev += objPos - highView
	}
	eyep = start + sacDev
	if eyep < sc.World.Min {
		sacDev += sc.World.Min - eyep
	} else if eyep > sc.World.Max {
		sacDev += sc.World.Max - eyep
	}
	return sacDev
}

// NextTraj computes the next object position and trajectory, at start of a
func (sc *SacEnv) NextTraj() {
	sc.TrajLen = sc.TrajLenRange.Min + rand.Intn(sc.TrajLenRange.Range()+1)
	zeroVel := erand.BoolProb(sc.ZeroVelP, -1)
	// keep same position
	// sc.ObjPosNext.X = sc.World.Min + rand.Float32()*sc.World.Range()
	// sc.ObjPosNext.Y = sc.World.Min + rand.Float32()*sc.World.Range()
	// sc.ObjPosNext.X = sc.World.Min + .5*sc.World.Range()
	// sc.ObjPosNext.Y = sc.World.Min + .5*sc.World.Range()
	if zeroVel {
		sc.ObjVelNext.SetZero()
	} else {
		sc.ObjVelNext.X = -sc.VelGenMax + 2*rand.Float32()*sc.VelGenMax
		sc.ObjVelNext.Y = -sc.VelGenMax + 2*rand.Float32()*sc.VelGenMax
		sc.ObjVelNext.X = sc.LimitVel(sc.ObjVelNext.X, sc.ObjPosNext.X, float32(sc.TrajLen))
		sc.ObjVelNext.Y = sc.LimitVel(sc.ObjVelNext.Y, sc.ObjPosNext.Y, float32(sc.TrajLen))
	}
	// saccade directly to position of new object at start -- set duration too
	sc.FixDur = sc.FixDurRange.Min + rand.Intn(sc.FixDurRange.Range()+1)
	// sc.SacPlan.X = sc.ObjPosNext.X - sc.EyePos.X
	// sc.SacPlan.Y = sc.ObjPosNext.Y - sc.EyePos.Y
	sc.NextSaccade()
	sc.SacTick.Cur = sc.SacTick.Max - 1 // ensure that we saccade next time
	sc.NewTrajNext = true
}

// NextSaccade generates next saccade plan
func (sc *SacEnv) NextSaccade() {
	sc.FixDur = sc.FixDurRange.Min + rand.Intn(sc.FixDurRange.Range()+1)
	sc.SacPlan.X = -sc.SacGenMax + 2*rand.Float32()*sc.SacGenMax
	sc.SacPlan.Y = -sc.SacGenMax + 2*rand.Float32()*sc.SacGenMax
	sc.SacPlan.X = sc.LimitSac(sc.SacPlan.X, sc.EyePos.X, sc.ObjPosNext.X, sc.ObjVelNext.X, float32(sc.FixDur))
	sc.SacPlan.Y = sc.LimitSac(sc.SacPlan.Y, sc.EyePos.Y, sc.ObjPosNext.Y, sc.ObjVelNext.Y, float32(sc.FixDur))
}

// DoSaccade updates current eye position with planned saccade, resets plan
func (sc *SacEnv) DoSaccade() {
	sc.EyePos.X = sc.EyePos.X + sc.SacPlan.X
	sc.EyePos.Y = sc.EyePos.Y + sc.SacPlan.Y
	sc.Saccade.X = sc.SacPlan.X
	sc.Saccade.Y = sc.SacPlan.Y
	sc.SacPlan.X = 0
	sc.SacPlan.Y = 0
}

// DoneSaccade clears saccade state
func (sc *SacEnv) DoneSaccade() {
	sc.Saccade.X = 0
	sc.Saccade.Y = 0
}

func (sc *SacEnv) String() string {
	return fmt.Sprintf("%s_%d", sc.ObjViewPos, sc.Tick.Cur)
}

func (sc *SacEnv) Counter(scale env.TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case env.Run:
		return sc.Run.Query()
	case env.Epoch:
		return sc.Epoch.Query()
	case env.Trial:
		return sc.Trial.Query()
	case env.Tick:
		return sc.Tick.Query()
	}
	return -1, -1, false
}

func (sc *SacEnv) State(element string) etensor.Tensor {
	switch element {
	case "EyePos":
		return &sc.EyePosTsr
	case "SacPlan":
		return &sc.SacPlanTsr
	case "Saccade":
		return &sc.SaccadeTsr
	case "ObjVel":
		return &sc.ObjVelTsr
	case "V1":
		return &sc.V1Tsr
	}
	return nil
}

func (sc *SacEnv) Action(element string, input etensor.Tensor) {
	// nop
}

// EncodePops encodes population codes from current row data
func (sc *SacEnv) EncodePops() {
	sc.V1Pop.Encode(&sc.V1Tsr, sc.ObjViewPos, popcode.Set)
	sc.EyePop.Encode(&sc.EyePosTsr, sc.EyePos, popcode.Set)
	sc.SacPop.Encode(&sc.SacPlanTsr, sc.SacPlan, popcode.Set)
	sc.SacPop.Encode(&sc.SaccadeTsr, sc.Saccade, popcode.Set)
	sc.ObjVelPop.Encode(&sc.ObjVelTsr, sc.ObjVel, popcode.Set)
}

// Step is primary method to call -- generates next state and
// outputs currents tate to table
func (sc *SacEnv) Step() bool {
	sc.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	sc.Trial.Same()

	sc.NewTraj = sc.Tick.Incr()
	sc.NewSac = sc.SacTick.Incr()

	if sc.NewTrajNext {
		sc.NewTrajNext = false
	}
	if sc.NewTraj {
		sc.Tick.Max = sc.TrajLen // was computed last time
		sc.ObjVel = sc.ObjVelNext
	}

	if sc.NewSac { // actually move eyes according to plan
		sc.DoSaccade()
		sc.SacTick.Max = sc.FixDur // was computed last time
	} else {
		sc.DoneSaccade()
	}
	// increment state -- next has already been computed
	sc.ObjPos = sc.ObjPosNext
	sc.ObjViewPos = sc.ObjPos.Sub(sc.EyePos)

	// now make new plans

	// if we will exceed traj next time, prepare new trajectory
	if sc.Tick.Cur+1 >= sc.Tick.Max {
		sc.NextTraj()
		if sc.Trial.Incr() {
			sc.Epoch.Incr()
		}
	} else { // otherwise, move object along and see if we need to plan saccade
		sc.ObjPosNext = sc.ObjPos.Add(sc.ObjVel)
		if sc.SacTick.Cur+1 >= sc.SacTick.Max {
			sc.NextSaccade()
		}
	}

	// write current state to table
	sc.WriteToTable(&sc.Table)
	sc.EncodePops()

	return true
}
