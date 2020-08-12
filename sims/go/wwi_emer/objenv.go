// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"log"
	"path/filepath"

	"github.com/anthonynsimon/bild/transform"
	"github.com/emer/emergent/env"
	"github.com/emer/emergent/popcode"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/mat32"
)

// Obj3DSacEnv provides the rendered results of the Obj3D + Saccade generator.
type Obj3DSacEnv struct {
	Nm        string          `desc:"name of this environment"`
	Dsc       string          `desc:"description of this environment"`
	Path      string          `desc:"path to data.tsv file as rendered, e.g., images/train"`
	Table     *etable.Table   `desc:"loaded table of generated trial / tick data"`
	IdxView   *etable.IdxView `desc:"indexed view of the table -- so you can do some additional filtering as needed -- sequential view created automatically if not otherwise set"`
	EyePop    popcode.TwoD    `desc:"2d population code for gaussian bump rendering of eye position"`
	SacPop    popcode.TwoD    `desc:"2d population code for gaussian bump rendering of saccade plan / execution"`
	ObjVelPop popcode.TwoD    `desc:"2d population code for gaussian bump rendering of object velocity"`
	V1Med     Vis             `desc:"v1 medium resolution filtering of image -- V1AllTsr has result"`
	V1Hi      Vis             `desc:"v1 higher resolution filtering of image -- V1AllTsr has result"`
	Run       env.Ctr         `view:"inline" desc:"current run of model as provided during Init"`
	Epoch     env.Ctr         `view:"inline" desc:"arbitrary aggregation of trials, for stats etc"`
	Trial     env.Ctr         `view:"inline" desc:"each object trajectory is one trial"`
	Tick      env.Ctr         `view:"inline" desc:"step along the trajectory"`
	Row       env.Ctr         `view:"inline" desc:"row of table -- this is actual counter driving everything"`
	CurCat    string          `desc:"current category"`
	CurObj    string          `desc:"current object"`

	// user can set the 2D shapes of these tensors -- Defaults sets default shapes
	EyePos  etensor.Float32 `view:"eye position popcode"`
	SacPlan etensor.Float32 `view:"saccade plan popcode"`
	Saccade etensor.Float32 `view:"saccade popcode "`
	ObjVel  etensor.Float32 `view:"object velocity"`

	Image image.Image `view:"-" desc:"rendered image as loaded"`
}

func (ev *Obj3DSacEnv) Name() string { return ev.Nm }
func (ev *Obj3DSacEnv) Desc() string { return ev.Dsc }

func (ev *Obj3DSacEnv) Validate() error {
	if ev.Table == nil {
		return fmt.Errorf("env.Obj3DSacEnv: %v has no Table set", ev.Nm)
	}
	if ev.Table.NumCols() == 0 {
		return fmt.Errorf("env.Obj3DSacEnv: %v Table has no columns -- Outputs will be invalid", ev.Nm)
	}
	return nil
}

func (ev *Obj3DSacEnv) Defaults() {
	ev.Path = "images/train"

	ev.EyePop.Defaults()
	ev.EyePop.Min.Set(-1.1, -1.1)
	ev.EyePop.Max.Set(1.1, 1.1)
	ev.EyePop.Sigma.Set(0.1, 0.1)

	ev.EyePos.SetShape([]int{21, 21}, nil, nil)

	ev.SacPop.Defaults()
	ev.SacPop.Min.Set(-0.45, -0.45)
	ev.SacPop.Max.Set(0.45, 0.45)

	ev.SacPlan.SetShape([]int{11, 11}, nil, nil)
	ev.Saccade.SetShape([]int{11, 11}, nil, nil)

	ev.ObjVelPop.Defaults()
	ev.ObjVelPop.Min.Set(-0.45, -0.45)
	ev.ObjVelPop.Max.Set(0.45, 0.45)

	ev.ObjVel.SetShape([]int{11, 11}, nil, nil)

	ev.V1Med.Defaults(24, 8)
	ev.V1Hi.Defaults(12, 4)

}

func (ev *Obj3DSacEnv) Init(run int) {
	ev.Run.Scale = env.Run
	ev.Epoch.Scale = env.Epoch
	ev.Trial.Scale = env.Trial
	ev.Tick.Scale = env.Tick
	ev.Row.Scale = env.Tick
	ev.Run.Init()
	ev.Epoch.Init()
	ev.Trial.Init()
	ev.Tick.Init()
	ev.Run.Cur = run
	ev.Row.Cur = -1 // init state -- key so that first Step() = 0
	ev.OpenTable()
}

// OpenTable loads data.tsv file at Path
func (ev *Obj3DSacEnv) OpenTable() error {
	if ev.Table == nil {
		ev.Table = etable.NewTable("obj3dsac_data")
	}
	fnm := filepath.Join(ev.Path, "data.tsv")
	err := ev.Table.OpenCSV(gi.FileName(fnm), etable.Tab)
	if err != nil {
		log.Println(err)
	} else {
		ev.Row.Max = ev.Table.Rows
	}
	return err
}

// CurRow returns current row in table, filtered through indexes
func (ev *Obj3DSacEnv) CurRow() int {
	if ev.IdxView == nil {
		ev.IdxView = etable.NewIdxView(ev.Table)
		ev.IdxView.Sequential()
	}
	return ev.IdxView.Idxs[ev.Row.Cur]
}

// OpenImage opens current image
func (ev *Obj3DSacEnv) OpenImage() error {
	row := ev.CurRow()
	ifnm := ev.Table.CellString("ImgFile", row)
	fnm := filepath.Join(ev.Path, ifnm)
	var err error
	ev.Image, err = gi.OpenImage(fnm)
	if err != nil {
		log.Println(err)
	}
	return err
}

// FilterImage opens and filters current image
func (ev *Obj3DSacEnv) FilterImage() error {
	err := ev.OpenImage()
	if err != nil {
		return err
	}
	// resize once for both..
	tsz := ev.V1Med.ImgSize
	isz := ev.Image.Bounds().Size()
	if isz != tsz {
		ev.Image = transform.Resize(ev.Image, tsz.X, tsz.Y, transform.Linear)
	}
	ev.V1Med.Filter(ev.Image)
	ev.V1Hi.Filter(ev.Image)
	return nil
}

// EncodePops encodes population codes from current row data
func (ev *Obj3DSacEnv) EncodePops() {
	row := ev.CurRow()
	val := mat32.Vec2{}
	val.X = float32(ev.Table.CellTensorFloat1D("EyePos", row, 0))
	val.Y = float32(ev.Table.CellTensorFloat1D("EyePos", row, 1))
	ev.EyePop.Encode(&ev.EyePos, val, popcode.Set)

	val.X = float32(ev.Table.CellTensorFloat1D("SacPlan", row, 0))
	val.Y = float32(ev.Table.CellTensorFloat1D("SacPlan", row, 1))
	ev.SacPop.Encode(&ev.SacPlan, val, popcode.Set)

	val.X = float32(ev.Table.CellTensorFloat1D("Saccade", row, 0))
	val.Y = float32(ev.Table.CellTensorFloat1D("Saccade", row, 1))
	ev.SacPop.Encode(&ev.Saccade, val, popcode.Set)

	val.X = float32(ev.Table.CellTensorFloat1D("ObjVel", row, 0))
	val.Y = float32(ev.Table.CellTensorFloat1D("ObjVel", row, 1))
	ev.ObjVelPop.Encode(&ev.ObjVel, val, popcode.Set)
}

// SetCtrs sets ctrs from current row data
func (ev *Obj3DSacEnv) SetCtrs() {
	row := ev.CurRow()
	epc := int(ev.Table.CellFloat("Epoch", row))
	ev.Epoch.Set(epc)
	trial := int(ev.Table.CellFloat("Trial", row))
	ev.Trial.Set(trial)
	tick := int(ev.Table.CellFloat("Tick", row))
	ev.Tick.Set(tick)

	ev.CurCat = ev.Table.CellString("Cat", row)
	ev.CurObj = ev.Table.CellString("Obj", row)
}

func (ev *Obj3DSacEnv) String() string {
	return fmt.Sprintf("%s:%s_%d", ev.CurCat, ev.CurObj, ev.Tick.Cur)
}

func (ev *Obj3DSacEnv) Step() bool {
	ev.Epoch.Same() // good idea to just reset all non-inner-most counters at start
	ev.Trial.Same()

	ev.Row.Incr() // auto-rotates

	ev.SetCtrs()
	ev.EncodePops()
	ev.FilterImage()

	return true
}

func (ev *Obj3DSacEnv) Counter(scale env.TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case env.Run:
		return ev.Run.Query()
	case env.Epoch:
		return ev.Epoch.Query()
	case env.Trial:
		return ev.Trial.Query()
	case env.Tick:
		return ev.Tick.Query()
	}
	return -1, -1, false
}

func (ev *Obj3DSacEnv) State(element string) etensor.Tensor {
	switch element {
	case "EyePos":
		return &ev.EyePos
	case "SacPlan":
		return &ev.SacPlan
	case "Saccade":
		return &ev.Saccade
	case "ObjVel":
		return &ev.ObjVel
	case "V1Med":
		return &ev.V1Med.V1AllTsr
	case "V1Hi":
		return &ev.V1Hi.V1AllTsr
	}
	et, err := ev.IdxView.Table.CellTensorTry(element, ev.CurRow())
	if err != nil {
		log.Println(err)
	}
	return et
}

func (ev *Obj3DSacEnv) Action(element string, input etensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ env.Env = (*Obj3DSacEnv)(nil)
