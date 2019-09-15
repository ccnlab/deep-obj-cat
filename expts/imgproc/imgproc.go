// remove background, threshold binarize, and box blur images

package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthonynsimon/bild/clone"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/chewxy/math32"
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
func (vi *Vis) ImgFmV1Simple() {
	vi.V1sUnPoolTsr.CopyShapeFrom(&vi.V1sTsr)
	vi.V1sUnPoolTsr.SetZeros()
	vi.ImgFmV1sTsr.CopyShapeFrom(&vi.ImgTsr)
	vi.ImgFmV1sTsr.SetZeros()
	vfilter.UnPool(image.Point{2, 2}, image.Point{2, 2}, &vi.V1sUnPoolTsr, &vi.V1sPoolTsr, true) // random max
	vfilter.Deconv(&vi.V1sGeom, &vi.V1sGaborTsr, &vi.ImgFmV1sTsr, &vi.V1sUnPoolTsr, vi.V1sGabor.Gain)
	// this goes straight from kwta and skips un-pooling:
	// vfilter.Deconv(&vi.V1sGeom, &vi.V1sGaborTsr, &vi.ImgFmV1sTsr, &vi.V1sKwtaTsr, vi.V1sGabor.Gain)
	vi.ImgFmV1sTsr.SetMetaData("image", "+")
	norm.Unit32(vi.ImgFmV1sTsr.Values)
}

// V1Invert does V1s filtering and inverts that back out to resulting image
func (vi *Vis) V1Invert(img *image.Gray) *image.Gray {
	vi.PrepImage(img)
	vi.V1Simple()
	vi.ImgFmV1Simple()
	return vfilter.GreyTensorToImage(img, &vi.ImgFmV1sTsr, vi.V1sGeom.FiltRt.X, false)
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

// ConfigGui configures the GoGi gui interface for this Vis
func (vi *Vis) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("v1gabor")
	gi.SetAppAbout(`This demonstrates basic V1 Gabor Filtering.  See <a href="https://github.com/emer/vision/v1">V1 on GitHub</a>.</p>`)

	win := gi.NewWindow2D("v1gabor", "V1 Gabor Filtering", width, height, true)
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

// AbsDif8 computes the absolute value of differnce
func AbsDif8(a uint8, b uint8) int {
	df := int(a) - int(b)
	if df < 0 {
		return -df
	}
	return df
}

func ColorMatchTol(c1, c2 color.RGBA, tol int) bool {
	if AbsDif8(c1.R, c2.R) <= tol && AbsDif8(c1.G, c2.G) <= tol && AbsDif8(c1.R, c2.R) <= tol {
		return true
	} else {
		return false
	}
}

// FilterBgAndThresh turns an image into a BW image by changing all bg pixels (within
// tolerance around each rgb val) to white, and everything else to black.
func FilterBgAndThresh(img image.Image, bg color.RGBA, tol int) *image.Gray {
	src := clone.AsRGBA(img)
	bounds := src.Bounds()

	dst := image.NewGray(bounds)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			srcPos := y*src.Stride + x*4
			dstPos := y*dst.Stride + x

			c := src.Pix[srcPos : srcPos+4]

			if ColorMatchTol(bg, color.RGBA{c[0], c[1], c[2], c[3]}, tol) {
				dst.Pix[dstPos] = 0xFF
			} else {
				dst.Pix[dstPos] = 0x00
			}
		}
	}

	return dst
}

// FilterBg turns an image into a Grey image by changing all bg pixels (within
// tolerance around each rgb val) to white, and everything else as it is.
func FilterBg(img image.Image, bg color.RGBA, tol int) *image.Gray {
	src := clone.AsRGBA(img)
	bounds := src.Bounds()

	dst := image.NewGray(bounds)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			srcPos := y*src.Stride + x*4
			dstPos := y*dst.Stride + x

			c := src.Pix[srcPos : srcPos+4]

			if ColorMatchTol(bg, color.RGBA{c[0], c[1], c[2], c[3]}, tol) {
				dst.Pix[dstPos] = 0xFF
			} else {
				dst.Pix[dstPos] = uint8(int(c[0]+c[1]+c[2]) / int(3))
			}
		}
	}
	return dst
}

// ToGrey turns an image into a Grey image
func ToGrey(img image.Image) *image.Gray {
	src := clone.AsRGBA(img)
	bounds := src.Bounds()
	dst := image.NewGray(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			srcPos := y*src.Stride + x*4
			dstPos := y*dst.Stride + x
			c := src.Pix[srcPos : srcPos+4]
			dst.Pix[dstPos] = uint8(int(c[0]+c[1]+c[2]) / int(3))
		}
	}
	return dst
}

// ImgStdDev returns the standard deviation of pixels in image and
// around the image border which is used to determine if there is anything
// there and should be skipped
func ImgStdDev(img *image.Gray, bord int) (float32, float32) {
	sz := img.Bounds().Size()
	npix := sz.Y * sz.X
	bsz := sz
	bsz.Y -= bord
	bsz.X -= bord
	bmean := float32(0)
	tmean := float32(0)
	bn := 0
	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			idx := y*img.Stride + x
			vl := float32(img.Pix[idx]) / 256
			tmean += vl
			if x >= bord && x < bsz.X && y >= bord && y < bsz.Y {
				continue
			}
			bmean += vl
			bn++
		}
	}
	tmean /= float32(npix)
	bmean /= float32(bn)
	tvr := float32(0)
	bvr := float32(0)
	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			idx := y*img.Stride + x
			vl := float32(img.Pix[idx]) / 256
			dv := vl - tmean
			tvr += dv * dv
			if x >= bord && x < bsz.X && y >= bord && y < bsz.Y {
				continue
			}
			dv = vl - bmean
			bvr += dv * dv
		}
	}
	tvr /= float32(npix)
	bvr /= float32(bn)
	return math32.Sqrt(tvr), math32.Sqrt(bvr)
}

func GetNmSegment(fn string, seg int) string {
	idx := 0
	for i := 0; i < seg; i++ {
		ni := strings.Index(fn[idx+1:], "_")
		idx += ni + 1
	}
	ni := strings.Index(fn[idx+1:], "_")
	return fn[idx+1 : idx+1+ni]
}

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

var TheVis Vis

func FilterDir(inpath, outpath string, bg color.RGBA, tol int, blurRad float64) {
	os.MkdirAll(outpath, 0755)

	files, err := ioutil.ReadDir(inpath)
	if err != nil {
		return
	}

	objNs := make(map[string]int, len(Cats))
	for k, _ := range Cats {
		objNs[k] = 0
	}

	// ntick := 8
	// tick := rand.Intn(ntick)
	tick := 4
	// epcCut := 100
	totPerCat := 25
	border := 8
	bordCut := float32(.0001) // std dev around border -- not much!
	imgCut := float32(.05)    // std dev within image -- need a lot! .1 excluded all.
	// .05 leaves just guitar at the end.

	vi := &TheVis
	vi.Defaults()

	fmt.Printf("starting\n")

	for _, fi := range files {
		// if i > 10 {
		// 	break
		// }
		fn := fi.Name()

		if strings.Index(fn, fmt.Sprintf("tick_%d_", tick)) < 0 {
			continue
		}

		// epcs := strings.TrimPrefix(GetNmSegment(fn, 1), "0")
		// epc, _ := strconv.Atoi(epcs)
		// if epc > epcCut {
		// 	fmt.Printf("fn: %v  epc: %v\n", fn, epc)
		// 	continue
		// }

		objnm := GetNmSegment(fn, 4)
		n := objNs[objnm]
		if n >= totPerCat {
			continue
		}
		// cat := Cats[objnm]

		fpath := filepath.Join(inpath, fn)
		img, err := imgio.Open(fpath)
		if err != nil {
			panic(err)
		}
		// flt := FilterBgAndThresh(img, bg, tol)
		// flt := FilterBg(img, bg, tol)
		flt := ToGrey(img)
		// blr := blur.Box(flt, blurRad) // Box is much blurrier than Gaussian for a given radius
		blr := vi.V1Invert(flt)

		tsd, bsd := ImgStdDev(blr, border)
		if bsd > bordCut { // too close to border
			// fmt.Printf("fn: %v  bord stdev: %v\n", fn, bsd)
			continue
		}
		if tsd < imgCut { // not enough signal in image -- too faint
			fmt.Printf("fn: %v  img stdev: %v\n", fn, tsd)
			continue
		}

		objNs[objnm] = n + 1

		ofpth := filepath.Join(outpath, objnm)
		os.MkdirAll(ofpth, 0755)

		// fmt.Printf("fn: %v obj: %v n: %v epc: %v  ofpth: %v\n", fn, objnm, n, epc, ofpth)

		ofn := filepath.Join(ofpth, fmt.Sprintf("%s_%d.jpg", objnm, n))

		if err := imgio.Save(ofn, blr, imgio.JPEGEncoder(95)); err != nil {
			panic(err)
		}
	}

	// win := vi.ConfigGui()
	// win.StartEventLoop()
}

func mainrun() {
	bg := color.RGBA{107, 184, 254, 255}
	tol := 20
	blurRad := 25.0
	tst := false
	if tst {
		FilterDir("/Users/oreilly/deep-obj-cat-shape-imgs-test-in", "/Users/oreilly/deep-obj-cat-shape-imgs-tst", bg, tol, blurRad)
	} else {
		FilterDir("/Users/oreilly/wwi_emer_imgs_20fg_8tick_rot1", "/Users/oreilly/deep-obj-cat-shape-imgs", bg, tol, blurRad)
	}

}
