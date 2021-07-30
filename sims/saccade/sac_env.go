// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/goki/mat32"
)

// SacEnv implements saccading logic for generating visual saccades
// toward one target object out of multiple possible objects.
// V1 = visual blob scene input; S1e = somatosensory eye position
// SCs = superior colliculus superficial "reflexive" motor plan
// SCd = superior colliculus deep layer actual motor output.
// PMD = probability of using MD from Action for actual eye command.
type SacEnv struct {
	Nm        string       `desc:"name of this environment"`
	Dsc       string       `desc:"description of this environment"`
	PMD       float64      `desc:"probability of using MD action"`
	NObjRange minmax.Int   `desc:"range for number of objects"`
	VisSize   int          `desc:"size of the visual input in each axis -- for visualization table too"`
	AngSize   int          `desc:"number of angle units for representing angle of polar coordinates"`
	DistSize  int          `desc:"number of distance units for representing distance of polar coordinates"`
	V1Pop     popcode.TwoD `desc:"2d population code for gaussian bump rendering of v1 obj position"`
	PolarPop  popcode.TwoD `desc:"2d population code for gaussian bump rendering of polar coords"`
	VisPop    popcode.TwoD `desc:"2d population code for visualization gaussian bump rendering of XY"`

	// State below here
	NObjs    int             `inactive:"-" desc:"number of objects"`
	TargObj  int             `inactive:"-" desc:"index of target object"`
	ObjsPos  []mat32.Vec2    `inactive:"+" desc:"object positions, in retinotopic coordinates when generated"`
	TargPos  mat32.Vec2      `inactive:"+" desc:"target position, in XY coordinates"`
	EyePos   mat32.Vec2      `inactive:"+" desc:"eye position, in XY coordinates"`
	SCs      mat32.Vec2      `inactive:"+" desc:"superior colliculus eye movement plan for next Step, XY coords"`
	SCdPolar mat32.Vec2      `inactive:"+" desc:"SCd current step actual eye movement action, polar coords"`
	SCdXY    mat32.Vec2      `inactive:"+" desc:"SCd current step actual eye movement action, XY coords"`
	MDPolar  mat32.Vec2      `inactive:"+" desc:"MD decoded polar coords"`
	MDXY     mat32.Vec2      `inactive:"+" desc:"SCd current step eye movement action, XY coords"`
	Table    etable.Table    `desc:"table showing visualization of state"`
	V1Tsr    etensor.Float32 `desc:"pop-code object blob(s)"`
	S1eTsr   etensor.Float32 `desc:"S1 primary somatosensory eye position polar popcode map"`
	SCsTsr   etensor.Float32 `desc:"SCs saccade plan polar popcode map"`
	SCdTsr   etensor.Float32 `desc:"SCd saccade actual executed action popcode map"`
	MDTsr    etensor.Float32 `desc:"MD corollary discharge Action taken by cortex"`
	Run      env.Ctr         `view:"inline" desc:"current run of model as provided during Init"`
	Epoch    env.Ctr         `view:"inline" desc:"arbitrary aggregation of trials, for stats etc"`
	Trial    env.Ctr         `view:"inline" desc:"each object trajectory is one trial"`
	Tick     env.Ctr         `inactive:"+" desc:"tick counter within trajectory, counts up from 0..TrajLen-1"`
}

func (sc *SacEnv) Name() string { return sc.Nm }
func (sc *SacEnv) Desc() string { return sc.Dsc }

// Defaults sets generic defaults -- use ParamSet to override
func (sc *SacEnv) Defaults() {
	sc.NObjRange.Set(1, 1)
	sc.VisSize = 11
	sc.AngSize = 11
	sc.DistSize = 11

	sc.V1Pop.Defaults()
	sc.V1Pop.Min.Set(-1.1, -1.1)
	sc.V1Pop.Max.Set(1.1, 1.1)
	sc.V1Pop.Sigma.Set(0.2, 0.2)

	sc.VisPop.Defaults()
	sc.VisPop.Min.Set(-1.1, -1.1)
	sc.VisPop.Max.Set(1.1, 1.1)
	sc.VisPop.Sigma.Set(0.1, 0.1)

	sc.PolarPop.Defaults()
	sc.PolarPop.Min.Set(-195, -0.1)
	sc.PolarPop.Max.Set(195, 1.1)
	sc.PolarPop.Sigma.Set(0.2, 0.2)

	sc.ConfigTable(&sc.Table)
	yx := []string{"Y", "X"}
	da := []string{"Dist", "Ang"}
	sc.V1Tsr.SetShape([]int{sc.VisSize, sc.VisSize}, nil, yx)
	sc.S1eTsr.SetShape([]int{sc.DistSize, sc.AngSize}, nil, da)
	sc.SCsTsr.SetShape([]int{sc.DistSize, sc.AngSize}, nil, da)
	sc.SCdTsr.SetShape([]int{sc.DistSize, sc.AngSize}, nil, da)
	sc.MDTsr.SetShape([]int{sc.DistSize, sc.AngSize}, nil, da)
}

// Init must be called at start prior to generating saccades
func (sc *SacEnv) Init(run int) {
	sc.Table.SetNumRows(1)
	sc.Run.Scale = env.Run
	sc.Epoch.Scale = env.Epoch
	sc.Trial.Scale = env.Trial
	sc.Tick.Scale = env.Tick
	sc.Tick.Max = 2
	sc.Run.Init()
	sc.Epoch.Init()
	sc.Trial.Init()
	sc.Tick.Cur = -1 // will increment to 0
	sc.Run.Cur = run
	sc.NewScene()
}

func (sc *SacEnv) Validate() error {
	return nil
}

func (sc *SacEnv) ConfigTable(dt *etable.Table) {
	yx := []string{"Y", "X"}
	da := []string{"Dist", "Ang"}
	sch := etable.Schema{
		{"TrialName", etensor.STRING, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"V1", etensor.FLOAT32, []int{sc.VisSize, sc.VisSize}, yx},
		{"Target", etensor.FLOAT32, []int{sc.VisSize, sc.VisSize}, yx},
		{"S1e", etensor.FLOAT32, []int{sc.DistSize, sc.AngSize}, da},
		{"SCs", etensor.FLOAT32, []int{sc.DistSize, sc.AngSize}, da},
		{"SCd", etensor.FLOAT32, []int{sc.DistSize, sc.AngSize}, da},
		{"MD", etensor.FLOAT32, []int{sc.DistSize, sc.AngSize}, da},
		{"TargPos", etensor.FLOAT32, []int{2}, nil},
	}
	dt.SetFromSchema(sch, 0)
}

func (sc *SacEnv) WriteToTable(dt *etable.Table) {
	row := 0
	dt.SetNumRows(row + 1)

	nm := fmt.Sprintf("t %d, x %+4.2f, y %+4.2f", sc.Tick.Cur, sc.TargPos.X, sc.TargPos.Y)

	dt.SetCellString("TrialName", row, nm)
	dt.SetCellFloat("Tick", row, float64(sc.Tick.Cur))

	dt.SetCellTensor("V1", row, &sc.V1Tsr)
	sc.VisPop.Encode(sc.Table.CellTensor("Target", row).(*etensor.Float32), sc.TargPos, popcode.Set)
	sc.PolarPop.Encode(sc.Table.CellTensor("S1e", row).(*etensor.Float32), sc.EyePos, popcode.Set)
	sc.PolarPop.Encode(sc.Table.CellTensor("SCs", row).(*etensor.Float32), sc.SCs, popcode.Set)
	sc.PolarPop.Encode(sc.Table.CellTensor("SCd", row).(*etensor.Float32), sc.SCdPolar, popcode.Set)
	dt.SetCellTensor("MD", row, &sc.MDTsr)

	dt.SetCellTensorFloat1D("TargPos", row, 0, float64(sc.TargPos.X))
	dt.SetCellTensorFloat1D("TargPos", row, 1, float64(sc.TargPos.Y))
}

// XYToPolar converts XY coordinates to polar (X=angle, Y=dist)
func XYToPolar(xy mat32.Vec2) mat32.Vec2 {
	var plr mat32.Vec2
	plr.X = mat32.Atan2(xy.Y, xy.X)
	plr.Y = xy.Length()
	return plr
}

// PolarToXY converts polar coordinates to XY (X=angle, Y=dist)
func PolarToXY(plr mat32.Vec2) mat32.Vec2 {
	var xy mat32.Vec2
	xy.X = plr.Y * mat32.Cos(plr.X)
	xy.Y = plr.Y * mat32.Sin(plr.X)
	return xy
}

// NewScene generates new scene of object(s) and eye positions
func (sc *SacEnv) NewScene() {
	sc.NObjs = sc.NObjRange.Min + rand.Intn(sc.NObjRange.Range()+1)
	sc.TargObj = rand.Intn(sc.NObjs)
	if cap(sc.ObjsPos) < sc.NObjs {
		sc.ObjsPos = make([]mat32.Vec2, sc.NObjs)
	} else {
		sc.ObjsPos = sc.ObjsPos[0:sc.NObjs]
	}
	for i := 0; i < sc.NObjs; i++ {
		var op mat32.Vec2
		op.X = -1 + rand.Float32()*2
		op.Y = -1 + rand.Float32()*2
		// todo: exclude if too close
		sc.ObjsPos[i] = op
	}
	sc.TargPos = sc.ObjsPos[sc.TargObj]

	// todo: random initial eye position
	sc.EyePos.Set(0, 0)
	sc.SCs = sc.TargPos
	sc.SCdXY.SetZero()
	sc.SCdPolar.SetZero()
}

// DoSaccade updates current eye position, vis targets with actual saccade, resets plan
func (sc *SacEnv) DoSaccade() {
	sc.EyePos.SetAdd(sc.SCdXY)
	sc.SCs.SetZero()
	sc.SCdXY.SetZero()
	sc.SCdPolar.SetZero()
	// eyepos drives render of V1, so obj pos not updated
}

func (sc *SacEnv) String() string {
	return fmt.Sprintf("%v_%d", sc.TargPos, sc.Tick.Cur)
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
	case "V1":
		return &sc.V1Tsr
	case "S1e":
		return &sc.S1eTsr
	case "SCs":
		return &sc.SCsTsr
	case "SCd":
		return &sc.SCdTsr
	}
	return nil
}

// EncodeObjs encodes objects with given offset into V1, omitting any out of range
func (sc *SacEnv) EncodeObjs(off mat32.Vec2) {
	for i := 0; i < sc.NObjs; i++ {
		op := sc.ObjsPos[i]
		op.SetAdd(off)
		if op.X > 1 || op.X < -1 || op.Y > 1 || op.Y < -1 {
			continue
		}
		if i == 0 {
			sc.V1Pop.Encode(&sc.V1Tsr, op, popcode.Set)
		} else {
			sc.V1Pop.Encode(&sc.V1Tsr, op, popcode.Add)
		}
	}
}

// EncodePolar encodes polar coords from given xy value
func (sc *SacEnv) EncodePolar(tsr *etensor.Float32, xy mat32.Vec2) {
	plr := XYToPolar(xy)
	ang := mat32.RadToDeg(plr.X) + 90 // vertical is 0 point
	if xy.X == 0 && xy.Y == 0 {
		ang = 0
	}
	if ang < -180 { // eg. -200 -> 160
		ang += 360
	} else if ang > 180 {
		ang -= 360
	}
	plr.X = ang
	sc.PolarPop.Encode(tsr, plr, popcode.Set)
}

// Encode encodes values into tensors
func (sc *SacEnv) Encode() {
	sc.EncodeObjs(sc.EyePos.Negate())
	sc.EncodePolar(&sc.S1eTsr, sc.EyePos)
	sc.EncodePolar(&sc.SCsTsr, sc.SCs)
	sc.EncodePolar(&sc.SCdTsr, sc.SCdXY)
}

// Step is primary method to call -- generates next state and
// outputs current state to table
func (sc *SacEnv) Step() bool {
	sc.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	sc.Trial.Same()

	sc.Tick.Incr()
	if sc.Tick.Cur == 0 {
		sc.NewScene()
	} else {
		sc.DoSaccade()
		if sc.Trial.Incr() {
			sc.Epoch.Incr()
		}
	}

	// write current state to table
	sc.Encode()
	sc.WriteToTable(&sc.Table)

	return true
}

func (sc *SacEnv) Action(element string, input etensor.Tensor) {
	// only MD accepted
	sc.MDTsr.CopyFrom(input)
	var err error
	sc.MDPolar, err = sc.PolarPop.Decode(&sc.MDTsr)
	if err != nil {
		fmt.Printf("MD Decoding error: %s\n", err)
	} else {
		sc.MDXY = PolarToXY(sc.MDPolar)
	}
	if erand.BoolProb(sc.PMD, -1) {
		sc.SCdPolar = sc.MDPolar
		sc.SCdXY = sc.MDXY
	} else {
		sc.SCdXY = sc.SCs // use SC
		sc.SCdPolar = XYToPolar(sc.SCdXY)
	}
	sc.EncodePolar(&sc.SCdTsr, sc.SCdXY)
	sc.WriteToTable(&sc.Table)
}
