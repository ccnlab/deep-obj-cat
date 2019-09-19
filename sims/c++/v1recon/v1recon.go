// remove background, threshold binarize, and box blur images

package main

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/emer/etable/norm"
	"github.com/emer/leabra/fffb"
	"github.com/emer/vision/gabor"
	"github.com/emer/vision/kwta"
	"github.com/emer/vision/vfilter"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/kit"
)

// this is the stub main for gogi that calls our actual
// mainrun function, at end of file
func main() {
	gimain.Main(func() {
		mainrun()
	})
}

// Vis encapsulates specific visual processing pipeline in
// use in a given case -- can add / modify this as needed
type Vis struct {
	V1sGabor      gabor.Filter    `desc:"V1 simple gabor filter parameters"`
	V1sGeom       vfilter.Geom    `inactive:"+" view:"inline" desc:"geometry of input, output for V1 simple-cell processing"`
	V1sNeighInhib kwta.NeighInhib `desc:"neighborhood inhibition for V1s -- each unit gets inhibition from same feature in nearest orthogonal neighbors -- reduces redundancy of feature code"`
	V1sKWTA       kwta.KWTA       `desc:"kwta parameters for V1s"`
	ImgSize       image.Point     `desc:"target image size to use -- images will be rescaled to this size"`
	V1sGaborTsr   etensor.Float32 `view:"no-inline" desc:"V1 simple gabor filter tensor"`
	Img           image.Image     `view:"-" desc:"current input image"`
	ImgTsr        etensor.Float32 `view:"no-inline" desc:"input image as tensor"`
	ImgFmV1sTsr   etensor.Float32 `view:"no-inline" desc:"input image reconstructed from V1s tensor"`
	V1sTsr        etensor.Float32 `view:"no-inline" desc:"V1 simple gabor filter output tensor"`
	V1sExtGiTsr   etensor.Float32 `view:"no-inline" desc:"V1 simple extra Gi from neighbor inhibition tensor"`
	V1sKwtaTsr    etensor.Float32 `view:"no-inline" desc:"V1 simple gabor filter output, kwta output tensor"`
	V1sPoolTsr    etensor.Float32 `view:"no-inline" desc:"V1 simple gabor filter output, max-pooled 2x2 of V1sKwta tensor"`
	V1sUnPoolTsr  etensor.Float32 `view:"no-inline" desc:"V1 simple gabor filter output, un-max-pooled 2x2 of V1sPool tensor"`
	V1cTsr        etensor.Float32 `view:"no-inline" desc:"V1 complex all tensor"`
	V1sInhibs     []fffb.Inhib    `view:"no-inline" desc:"inhibition values for V1s KWTA"`
}

var KiT_Vis = kit.Types.AddType(&Vis{}, nil)

func (vi *Vis) Defaults() {
	vi.V1sGabor.Defaults()
	sz := 12 // V1mF16 typically = 12, no border
	spc := 4
	vi.V1sGabor.SetSize(sz, spc)
	// note: first arg is border -- we are relying on Geom
	// to set border to .5 * filter size
	// any further border sizes on same image need to add Geom.FiltRt!
	vi.V1sGeom.Set(image.Point{0, 0}, image.Point{spc, spc}, image.Point{sz, sz})
	vi.V1sNeighInhib.Defaults()
	vi.V1sKWTA.Defaults()
	// vi.V1sNeighInhib.On = false
	// vi.V1sKWTA.On = false
	vi.ImgSize = image.Point{128, 128}
	// vi.ImgSize = image.Point{64, 64}
	vi.V1sGabor.ToTensor(&vi.V1sGaborTsr)
}

// PrepImage prepares image for v1 processing:
// converts to a float32 tensor for processing
func (vi *Vis) PrepImage(img image.Image) error {
	isz := img.Bounds().Size()
	if isz != vi.ImgSize {
		vi.Img = transform.Resize(img, vi.ImgSize.X, vi.ImgSize.Y, transform.Linear)
	} else {
		vi.Img = img
	}
	vfilter.RGBToGrey(vi.Img, &vi.ImgTsr, vi.V1sGeom.FiltRt.X, false) // pad for filt, bot zero
	vfilter.WrapPad(&vi.ImgTsr, vi.V1sGeom.FiltRt.X)
	vi.ImgTsr.SetMetaData("image", "+")
	return nil
}

// V1Simple runs V1Simple Gabor filtering on input image
// must have valid Img in place to start.
// Runs kwta and pool steps after gabor filter.
func (vi *Vis) V1Simple() {
	vfilter.Conv(&vi.V1sGeom, &vi.V1sGaborTsr, &vi.ImgTsr, &vi.V1sTsr, vi.V1sGabor.Gain)
	if vi.V1sNeighInhib.On {
		vi.V1sNeighInhib.Inhib4(&vi.V1sTsr, &vi.V1sExtGiTsr)
	} else {
		vi.V1sExtGiTsr.SetZeros()
	}
	if vi.V1sKWTA.On {
		vi.V1sKWTA.KWTAPool(&vi.V1sTsr, &vi.V1sKwtaTsr, &vi.V1sInhibs, &vi.V1sExtGiTsr)
	} else {
		vi.V1sKwtaTsr.CopyFrom(&vi.V1sTsr)
	}
	vfilter.MaxPool(image.Point{2, 2}, image.Point{2, 2}, &vi.V1sKwtaTsr, &vi.V1sPoolTsr)
}

// ImgFmV1Simple reverses V1Simple Gabor filtering from V1s back to input image
func (vi *Vis) ImgFmV1Simple(v1data string) {
	psz := 16
	vi.V1cTsr.SetShape([]int{psz, psz, 5, 4}, nil, []string{"PY", "PX", "Feat", "Ang"})
	etensor.OpenCSV(&vi.V1cTsr, gi.FileName(v1data), etable.Tab)

	vi.V1sTsr.SetShape([]int{psz, psz, 2, 4}, nil, []string{"PY", "PX", "Feat", "Ang"})
	for py := 0; py < psz; py++ {
		for px := 0; px < psz; px++ {
			for fy := 0; fy < 2; fy++ {
				for ang := 0; ang < 4; ang++ {
					cv := vi.V1cTsr.Value([]int{py, px, 3 + fy, ang}) // v1s at end
					vi.V1sTsr.Set([]int{py, px, fy, ang}, cv)
				}
			}
		}
	}

	// vi.V1sUnPoolTsr.CopyShapeFrom(&vi.V1sTsr)
	// vi.V1sUnPoolTsr.SetZeros()
	isz := 64
	ipd := isz + 2*vi.V1sGeom.FiltRt.X // padded
	vi.ImgFmV1sTsr.SetShape([]int{ipd, ipd}, nil, nil)
	vi.ImgFmV1sTsr.SetZeros()
	// vfilter.UnPool(image.Point{2, 2}, image.Point{2, 2}, &vi.V1sUnPoolTsr, &vi.V1sPoolTsr, true) // random max
	vfilter.Deconv(&vi.V1sGeom, &vi.V1sGaborTsr, &vi.ImgFmV1sTsr, &vi.V1sTsr, vi.V1sGabor.Gain)
	// this goes straight from kwta and skips un-pooling:
	// vfilter.Deconv(&vi.V1sGeom, &vi.V1sGaborTsr, &vi.ImgFmV1sTsr, &vi.V1sKwtaTsr, vi.V1sGabor.Gain)
	vi.ImgFmV1sTsr.SetMetaData("image", "+")
	norm.Unit32(vi.ImgFmV1sTsr.Values)
	img := vfilter.GreyTensorToImage(nil, &vi.ImgFmV1sTsr, vi.V1sGeom.FiltRt.X, false)
	ofn := filepath.Base(v1data) + ".jpg"
	if err := imgio.Save(ofn, img, imgio.JPEGEncoder(95)); err != nil {
		panic(err)
	}
}

// RecSeq reconstructs sequence starting with given object name
func (vi *Vis) RecSeq(objnm string) {
	for tick := 0; tick < 8; tick++ {
		fn := fmt.Sprintf("%s_tick_%d_sac_%d", objnm, tick, tick%2)
		vi.ImgFmV1Simple(fn + "_actm.tsv")
		vi.ImgFmV1Simple(fn + "_actp.tsv")
	}
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

// ConfigGui configures the GoGi gui interface for this Vis
func (vi *Vis) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("v1recon")
	gi.SetAppAbout(`This reconstructs input images from v1 activity patterns.  See <a href="https://github.com/emer/vision/v1">V1 on GitHub</a>.</p>`)

	win := gi.NewWindow2D("v1recon", "V1 Image Reconstruction", width, height, true)
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
	sv.SetStruct(vi)

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

var TheVis Vis

func mainrun() {
	TheVis.Defaults()
	// TheVis.RecSeq("car_sedan_002")
	TheVis.RecSeq("slrcamera_004")
	win := TheVis.ConfigGui()
	win.StartEventLoop()
}
