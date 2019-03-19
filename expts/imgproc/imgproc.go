// remove background, threshold binarize, and box blur images

package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/clone"
	"github.com/anthonynsimon/bild/imgio"
)

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

	ntick := 8
	epcCut := 100

	for _, fi := range files {
		// if i > 10 {
		// 	break
		// }
		fn := fi.Name()

		tick := rand.Intn(ntick)

		if strings.Index(fn, fmt.Sprintf("tick_%d_", tick)) < 0 {
			continue
		}

		epcs := strings.TrimPrefix(GetNmSegment(fn, 1), "0")
		epc, _ := strconv.Atoi(epcs)
		if epc > epcCut {
			continue
		}

		objnm := GetNmSegment(fn, 4)
		n := objNs[objnm]
		objNs[objnm] = n + 1
		// cat := Cats[objnm]

		fpath := filepath.Join(inpath, fn)
		img, err := imgio.Open(fpath)
		if err != nil {
			panic(err)
		}

		flt := FilterBgAndThresh(img, bg, tol)
		blr := blur.Box(flt, blurRad) // Box is much blurrier than Gaussian for a given radius

		ofpth := filepath.Join(outpath, objnm)
		os.MkdirAll(ofpth, 0755)

		// fmt.Printf("fn: %v obj: %v n: %v epc: %v  ofpth: %v\n", fn, objnm, n, epc, ofpth)

		ofn := filepath.Join(ofpth, fmt.Sprintf("%s_%d.jpg", objnm, n))

		if err := imgio.Save(ofn, blr, imgio.JPEGEncoder(95)); err != nil {
			panic(err)
		}
	}
}

func main() {
	bg := color.RGBA{107, 184, 254, 255}
	tol := 20
	blurRad := 25.0
	tst := false
	if tst {
		FilterDir("/Users/oreilly/wwi_filter_tst", "/Users/oreilly/wwi_filter_out", bg, tol, blurRad)
	} else {
		FilterDir("/Users/oreilly/wwi_emer_imgs_20fg_8tick_rot1", "/Users/oreilly/wwi_filter_out", bg, tol, blurRad)
	}

}
