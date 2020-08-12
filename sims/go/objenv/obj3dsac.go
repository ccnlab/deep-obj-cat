// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/gpu"
	"github.com/goki/mat32"
)

// Obj3DSac generates renderings of 3D objects with saccadic eye movements.
// Object is moved around using Sac positions relative to 0,0,0 origin
// and Camera is at 0,0,Z with rotations based on saccade movements.
// Can save a table of saccade data plus image file names to use for
// offline training of models on a cluster using this code, or
// incorporate into an env for more dynamic uses.
type Obj3DSac struct {
	Objs       Obj3D         `desc:"list of 3D objects"`
	Sac        Saccade       `desc:"saccade control"`
	Env        Obj3DSacEnv   `desc:"environment that loads rendered images"`
	SaveFiles  bool          `desc:"if true, save images (in epoch-wise subdirs) and data.tsv file with saccade position data and image name, to images dir"`
	NTrials    int           `desc:"number of trials per epoch, for saving"`
	NEpcs      int           `desc:"number of epochs"`
	Table      *etable.Table `desc:"if saving, this is the trial-by-trial data"`
	Train      bool          `desc:"if true, use training set of objects, else test"`
	Sequential bool          `desc:"if true, present in sequential order -- else permuted"`
	FOV        float32       `desc:"field of view for camera"`
	ImgSize    image.Point   `desc:"size of image to render"`
	ViewScale  int           `desc:"scale factor for viewing the image"`
	CamPos     mat32.Vec3    `desc:"camera position -- object is positioned around 0,0,0"`
	ZOffScale  float32       `desc:"multiplies the object XYtrg - XYsz to set Zoff to keep general size of objects about the same"`
	ZOffXYtrg  float32       `desc:"target XYsz for z offset scaling"`
	XYPosScale float32       `desc:"multiplier for X,Y positions from saccade"`
	Rot3D      mat32.Vec3    `desc:"how much to rotate along each axis, in degrees per step"`
	ObjIdx     int           `desc:"index in objs list of current object"`
	Order      []int         `desc:"order to present items (permuted or sequential)"`

	CurObj  string     `inactive:"+" desc:"current object to show"`
	CurCat  string     `inactive:"+" desc:"current category to show"`
	CurXYsz float32    `inactive:"+" desc:"current object XY size"`
	CurZoff float32    `inactive:"+" desc:"current Z offset based on XYsz and ZOff*"`
	InitRot mat32.Vec3 `inactive:"+" desc:"initial euler 3D rotation, in degrees"`
	RotVel  mat32.Vec3 `inactive:"+" desc:"3D rotational velocity (degrees per step) for current object"`
	CurRot  mat32.Vec3 `inactive:"+" desc:"current euler rotation"`
	Trial   env.Ctr    `inactive:"+" desc:"current trial, for saving"`
	Epoch   env.Ctr    `inactive:"+" desc:"current epoch, for saving"`
	SaveDir string     `inactive:"+" desc:"name of current directory to save files into (images/train or images/test)"`
	ImgFile string     `inactive:"+" desc:"name of image file"`

	Image     image.Image     `view:"-" desc:"current rendered image in specified size"`
	ViewImage *gi.Bitmap      `view:"-" desc:"View of image (scaled up) as a bitmap"`
	Scene     *gi3d.Scene     `view:"-" desc:"3D scene"`
	Group     *gi3d.Group     `view:"-" desc:"group holding loaded object"`
	Frame     gpu.Framebuffer `view:"-" desc:"offscreen render buffer"`
	File      *os.File        `view:"-" desc:"save file"`
	StopNow   bool            `view:"-" desc:"flag to stop running"`
	IsRunning bool            `view:"-" desc:"true when running"`
}

func (ob *Obj3DSac) Defaults() {
	hdir, _ := os.UserHomeDir()
	path := filepath.Join(hdir, "ccn_images/CU3D_100_plus_models_obj") // downloadable from TODO
	ob.Objs.Path = path
	ob.Objs.NTestPerCat = 2
	ob.Objs.OpenCatProps("cu3d_obj_cat_props.csv")
	ob.Sac.Defaults()
	ob.Sac.TrajLenRange.Set(8, 8)
	ob.NTrials = 64
	ob.NEpcs = 2
	ob.Train = true
	ob.FOV = 50
	ob.ImgSize = image.Point{256, 256}
	ob.ViewScale = 2
	ob.CamPos.Z = 3 // set to have object take about 1/2 of width of display overall
	ob.ZOffScale = 2
	ob.ZOffXYtrg = 0.6
	ob.XYPosScale = 1.5
	ob.Rot3D.Set(0, 5, 0.5)
	ob.Trial.Scale = env.Trial
	ob.Epoch.Scale = env.Epoch

	ob.Env.Defaults()
}

func (ob *Obj3DSac) Config() {
	ob.Sac.Init()
	ob.Objs.Open()
	ob.Objs.DeleteCats(ObjsBigSlow) // avoid!
	// ob.Objs.SelectCats() // todo: add sub-lists here
	ob.Init()
}

// ConfigScene must be called with pointer to Scene that is created
// in some form in GUI -- Scene must have access to a Window
func (ob *Obj3DSac) ConfigScene(sc *gi3d.Scene) {
	ob.Scene = sc
	sc.SetStretchMax()
	sc.Defaults()
	sc.BgColor.SetUInt8(103, 176, 255, 255) // sky blue
	sc.Camera.FOV = ob.FOV
	sc.Camera.Pose.Pos = ob.CamPos
	sc.Camera.LookAt(mat32.Vec3Zero, mat32.Vec3Y) // defaults to looking at origin
	dir := gi3d.AddNewDirLight(sc, "dir", 1, gi3d.DirectSun)
	dir.Pos.Set(0, 0, 1) // default: 0,1,1 = above and behind us (we are at 0,0,X)
	ob.Group = gi3d.AddNewGroup(sc, sc, "obj-gp")
}

// ObjList returns the object list to use (Train or Test)
func (ob *Obj3DSac) ObjList() []string {
	if ob.Train {
		return ob.Objs.FlatTrain
	} else {
		return ob.Objs.FlatTest
	}
}

// Init restarts counters
func (ob *Obj3DSac) Init() {
	ob.ObjIdx = -1
	nobj := len(ob.ObjList())
	ob.Order = rand.Perm(nobj)
	ob.Sac.Init()
}

// OpenObj opens object from file path -- relative to Objs.Path
func (ob *Obj3DSac) OpenObj(obj string) error {
	fn := filepath.Join(ob.Objs.Path, obj)
	sc := ob.Scene
	updt := sc.UpdateStart()
	ob.Group.DeleteChildren(true)
	sc.DeleteMeshes()
	sc.DeleteTextures()
	_, err := sc.OpenNewObj(fn, ob.Group)
	if err != nil {
		log.Println(err)
	}
	sc.UpdateEnd(updt)
	sc.UpdateMeshBBox()
	ob.CurXYsz = 0.5 * (ob.Group.MeshBBox.BBox.Max.X + ob.Group.MeshBBox.BBox.Max.Y)
	ob.CurZoff = ob.ZOffScale * (ob.ZOffXYtrg - ob.CurXYsz)
	crows := ob.Objs.ObjCatProps.RowsByString("category", ob.CurCat, etable.Equals, etable.UseCase)
	crow := crows[0]
	// zoff := float32(ob.Objs.ObjCatProps.CellFloat("z_offset", crow))
	ymirv := ob.Objs.ObjCatProps.CellFloat("y_rot_mirror", crow)
	ymir := ymirv != 0
	yflip := erand.BoolProb(.5, -1)
	if ymir && yflip {
		ob.InitRot.Y = 180
	} else {
		ob.InitRot.Y = 0
	}
	ob.RotVel.Z = -ob.Rot3D.Z + 2*ob.Rot3D.Z*rand.Float32()
	ob.RotVel.Y = -ob.Rot3D.Y + 2*ob.Rot3D.Y*rand.Float32()
	return err
}

// Render generates image from current object, saving to Image
func (ob *Obj3DSac) Render() error {
	frame := &ob.Frame
	sc := ob.Scene
	if !sc.ActivateOffFrame(frame, "objrend", ob.ImgSize, 4) { // 4 samples
		err := fmt.Errorf("could not activate offscreen framebuffer")
		log.Println(err)
		return err
	}
	if !sc.RenderOffFrame() {
		err := fmt.Errorf("could not render to offscreen framebuffer")
		log.Println(err)
		return err
	}
	(*frame).Rendered()
	// ob.Scene.Render2D()
	// ob.Scene.DirectWinUpload()

	oswin.TheApp.RunOnMain(func() {
		tex := (*frame).Texture()
		tex.SetBotZero(true)
		ob.Image = tex.GrabImage()
	})
	if ob.ViewImage != nil {
		vwsz := ob.ImgSize.Mul(ob.ViewScale)
		ob.ViewImage.SetImage(ob.Image, float32(vwsz.X), float32(vwsz.Y))
	}
	return nil
}

// Position puts object into position according to saccade table
func (ob *Obj3DSac) Position() {
	op := mat32.Vec3{}
	op.Z += ob.CurZoff
	op.X += ob.Sac.ObjPos.X * ob.XYPosScale
	op.Y += ob.Sac.ObjPos.Y * ob.XYPosScale
	ob.Group.Pose.Pos = op
	ob.CurRot = ob.InitRot.Add(ob.RotVel.MulScalar(float32(ob.Sac.Tick.Cur)))
	ob.Group.Pose.SetEulerRotation(ob.CurRot.X, ob.CurRot.Y, ob.CurRot.Z)
}

// Fixate moves eyes to fixate on eye position
func (ob *Obj3DSac) Fixate() {
	sc := ob.Scene
	trg := mat32.Vec3{}
	trg.X = ob.Sac.EyePos.X * ob.XYPosScale
	trg.Y = ob.Sac.EyePos.Y * ob.XYPosScale
	sc.Camera.LookAt(trg, mat32.Vec3Y)
}

// SetObj sets the current obj info based on flat list of objects
func (ob *Obj3DSac) SetObj(list []string) {
	if ob.ObjIdx >= len(list) {
		ob.ObjIdx = 0
		erand.PermuteInts(ob.Order)
	}
	idx := ob.ObjIdx
	if !ob.Sequential {
		idx = ob.Order[ob.ObjIdx]
	}
	ob.CurObj = list[idx]
	ob.CurCat = strings.Split(ob.CurObj, "/")[0]
	ob.OpenObj(ob.CurObj)
}

// Step iterates to next item
func (ob *Obj3DSac) Step() {
	ob.Sac.Step()
	if ob.Sac.NewTraj || ob.ObjIdx < 0 {
		ob.ObjIdx++
		ob.SetObj(ob.ObjList()) // wraps objidx
		if ob.Trial.Incr() {
			if ob.Epoch.Incr() {
				ob.Stop()
				return

			}
		}
	}
	ob.Position()
	ob.Fixate()
	ob.Render()
	ob.SaveTick()
}

// Run runs full set of Save trials / epochs
func (ob *Obj3DSac) Run() {

	ob.Trial.Max = ob.NTrials
	ob.Epoch.Max = ob.NEpcs
	ob.Trial.Init()
	ob.Trial.Cur = -1 // gets inc to 0 at start
	ob.Epoch.Init()

	ob.Table = &etable.Table{}
	ob.ConfigTable(ob.Table)
	ob.Table.SetNumRows(1) // just re-use same row.. fine..

	if ob.Train {
		ob.SaveDir = "images/train"
	} else {
		ob.SaveDir = "images/test"
	}

	var err error
	os.MkdirAll(ob.SaveDir, 0755)

	ob.File, err = os.Create(filepath.Join(ob.SaveDir, "data.tsv"))
	if err != nil {
		log.Println(err)
		return
	}

	ob.SaveFiles = true
	ob.IsRunning = true

	vp := ob.Scene.Win.WinViewport2D()

	for {
		ob.Step()
		vp.FullRender2DTree()
		if ob.StopNow {
			ob.StopNow = false
			break
		}
	}
	ob.IsRunning = false
	vp.FullRender2DTree()
}

// Stop tells the sim to stop running
func (ob *Obj3DSac) Stop() {
	ob.StopNow = true
}

func (ob *Obj3DSac) ConfigTable(dt *etable.Table) {
	dt.SetMetaData("name", "Obj3DSacTable")
	dt.SetMetaData("desc", "table of obj3d data")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Epoch", etensor.INT64, nil, nil},
		{"Trial", etensor.INT64, nil, nil},
		{"Tick", etensor.INT64, nil, nil},
		{"SacTick", etensor.INT64, nil, nil},
		{"Cat", etensor.STRING, nil, nil},
		{"Obj", etensor.STRING, nil, nil},
		{"ImgFile", etensor.STRING, nil, nil},
		{"ObjPos", etensor.FLOAT32, []int{2}, nil},
		{"ObjViewPos", etensor.FLOAT32, []int{2}, nil},
		{"ObjVel", etensor.FLOAT32, []int{2}, nil},
		{"ObjPosNext", etensor.FLOAT32, []int{2}, nil},
		{"ObjRot", etensor.FLOAT32, []int{3}, nil},
		{"EyePos", etensor.FLOAT32, []int{2}, nil},
		{"SacPlan", etensor.FLOAT32, []int{2}, nil},
		{"Saccade", etensor.FLOAT32, []int{2}, nil},
	}
	dt.SetFromSchema(sch, 0)
}

// SaveTick saves the current tick, if saving
func (ob *Obj3DSac) SaveTick() {
	if !ob.SaveFiles || ob.Table == nil {
		return
	}
	sc := &ob.Sac

	epc := ob.Epoch.Cur
	trl := ob.Trial.Cur
	tick := sc.Tick.Cur

	obj := strings.Split(ob.CurObj, "/")[1]

	epcdir := fmt.Sprintf("epc_%04d", epc)
	imgdir := filepath.Join(ob.SaveDir, epcdir)
	os.MkdirAll(imgdir, 0755)

	ob.ImgFile = fmt.Sprintf("%s/img_%04d_%03d_%d.jpg", epcdir, epc, trl, tick)

	gi.SaveImage(filepath.Join(ob.SaveDir, ob.ImgFile), ob.Image)

	dt := ob.Table
	row := 0

	dt.SetCellFloat("Epoch", row, float64(epc))
	dt.SetCellFloat("Trial", row, float64(trl))
	dt.SetCellFloat("Tick", row, float64(tick))
	dt.SetCellFloat("SacTick", row, float64(sc.SacTick.Cur))
	dt.SetCellString("Cat", row, ob.CurCat)
	dt.SetCellString("Obj", row, obj)
	dt.SetCellString("ImgFile", row, ob.ImgFile)

	// this is from saccade.go:
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

	dt.SetCellTensorFloat1D("ObjRot", row, 0, float64(ob.CurRot.X))
	dt.SetCellTensorFloat1D("ObjRot", row, 1, float64(ob.CurRot.Y))
	dt.SetCellTensorFloat1D("ObjRot", row, 2, float64(ob.CurRot.Z))

	if ob.File != nil {
		if trl == 0 && epc == 0 && tick == 0 {
			dt.WriteCSVHeaders(ob.File, etable.Tab)
		}
		dt.WriteCSVRow(ob.File, row, etable.Tab)
	}
}
