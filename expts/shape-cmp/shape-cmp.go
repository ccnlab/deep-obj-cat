// generate random comparison cases

package main

import (
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/emer/etable/etensor"
)

func PlaceImage(dst, src image.Image, offset image.Point) {
	r := src.Bounds().Add(offset)
	draw.Draw(dst.(*image.RGBA), r, src, image.ZP, draw.Src)
}

func CmpImages(inpath, ima1, ima2, imb1, imb2 string) image.Image {
	// fmt.Printf("inpath: %s ima1: %s  ima2: %s  imb1: %s  imb2: %s\n", inpath, ima1, ima2, imb1, imb2)

	imga1, _ := imgio.Open(filepath.Join(inpath, ima1))
	imga2, _ := imgio.Open(filepath.Join(inpath, ima2))
	imgb1, _ := imgio.Open(filepath.Join(inpath, imb1))
	imgb2, _ := imgio.Open(filepath.Join(inpath, imb2))

	smspc := 10
	pspc := 50

	iszX := imga1.Bounds().Size().X
	iszY := imga1.Bounds().Size().Y

	tszX := iszX*4 + 2*smspc + pspc
	tszY := iszY

	dst := image.NewRGBA(image.Rect(0, 0, tszX, tszY))
	PlaceImage(dst, imga1, image.Point{0, 0})
	PlaceImage(dst, imga2, image.Point{iszX + smspc, 0})
	PlaceImage(dst, imgb1, image.Point{2*iszX + smspc + pspc, 0})
	PlaceImage(dst, imgb2, image.Point{3*iszX + 2*smspc + pspc, 0})
	return dst
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

type CatRec struct {
	Name  string
	Start int
	End   int
}

var CatRecs = []CatRec{
	{"pyramid", 0, 5},
	{"vertical", 5, 8},
	{"round", 8, 12},
	{"box", 12, 16},
	{"horiz", 16, 20},
}

// RndFnm picks a random file name from list at given object index
func RndFnm(objfiles [][]os.FileInfo, objidx int) (string, string) {
	fl := objfiles[objidx]
	onm := Objs[objidx]
	fnm := ""
	for {
		idx := rand.Intn(len(fl))
		fnm = fl[idx].Name()
		if fnm[0] == '.' || !strings.HasSuffix(fnm, ".jpg") {
			continue
		}
		break
	}
	return onm, filepath.Join(onm, fnm)
}

// PurelyRandomPairs generates trials with two pairs of images drawn purely at random
// using permution to ensure that they are at least from different "real" categories.
func PurelyRandomPairs(inpath, outpath string, ntrl int) {
	os.MkdirAll(outpath, 0755)

	fnlist, _ := os.Create(filepath.Join(outpath, "file-list.csv"))
	trllist, _ := os.Create(filepath.Join(outpath, "trial-list.csv"))

	fmt.Fprintf(fnlist, "image_url\n")

	nobj := len(Objs)

	objfiles := make([][]os.FileInfo, nobj)
	for i, ob := range Objs {
		opth := filepath.Join(inpath, ob)
		objfiles[i], _ = ioutil.ReadDir(opth)
	}

	pcnt := etensor.NewInt64([]int{nobj, nobj}, nil, nil)

	robjs := rand.Perm(nobj)
	cidx := 0

	for trl := 0; trl < ntrl; trl++ {
		oa1dx := robjs[cidx]
		oa2dx := robjs[cidx+1]
		ob1dx := robjs[cidx+2]
		ob2dx := robjs[cidx+3]
		cidx += 4
		if cidx >= nobj {
			rand.Shuffle(nobj, func(i, j int) {
				robjs[i], robjs[j] = robjs[j], robjs[i]
			})
			cidx = 0
		}

		pcnt.Set([]int{oa1dx, oa2dx}, pcnt.Value([]int{oa1dx, oa2dx})+1)
		pcnt.Set([]int{oa2dx, oa1dx}, pcnt.Value([]int{oa2dx, oa1dx})+1)
		pcnt.Set([]int{ob1dx, ob2dx}, pcnt.Value([]int{ob1dx, ob2dx})+1)
		pcnt.Set([]int{ob2dx, ob1dx}, pcnt.Value([]int{ob2dx, ob1dx})+1)

		oa1nm, oa1fnm := RndFnm(objfiles, oa1dx)
		oa2nm, oa2fnm := RndFnm(objfiles, oa2dx)
		ob1nm, ob1fnm := RndFnm(objfiles, ob1dx)
		ob2nm, ob2fnm := RndFnm(objfiles, ob2dx)

		dst := CmpImages(inpath, oa1fnm, oa2fnm, ob1fnm, ob2fnm)

		ofn := fmt.Sprintf("trial_%d.jpg", trl)

		fmt.Fprintf(fnlist, "%s\n", ofn)
		fmt.Fprintf(trllist, "%s,%s,%s,%s,%s\n", ofn, oa1nm, oa2nm, ob1nm, ob2nm)

		ofpth := filepath.Join(outpath, ofn)
		if err := imgio.Save(ofpth, dst, imgio.JPEGEncoder(95)); err != nil {
			panic(err)
		}
	}
	fnlist.Close()
	trllist.Close()

	fmt.Println(pcnt.String())
}

// CategPairs generates trials using target category structure, such that one
// pair is always within category and the other is always between categories
// so there should be a "right answer"
func CategPairs(inpath, outpath string, ntrl int) {
	os.MkdirAll(outpath, 0755)

	fnlist, _ := os.Create(filepath.Join(outpath, "file-list.csv"))
	trllist, _ := os.Create(filepath.Join(outpath, "trial-list.csv"))

	fmt.Fprintf(fnlist, "image_url\n")

	nobj := len(Objs)

	objfiles := make([][]os.FileInfo, nobj)
	for i, ob := range Objs {
		opth := filepath.Join(inpath, ob)
		objfiles[i], _ = ioutil.ReadDir(opth)
	}

	pcnt := etensor.NewInt64([]int{nobj, nobj}, nil, nil)

	robjs := rand.Perm(nobj)
	cidx := 0

	ncat := len(CatRecs)

	for trl := 0; trl < ntrl; trl++ {

		var oa1dx, oa2dx, ob1dx, ob2dx int

		lr := rand.Intn(2)
		cati := rand.Intn(ncat)
		cr := CatRecs[cati]
		cno := cr.End - cr.Start
		crnd := rand.Perm(cno)
		c1 := cr.Start + crnd[0]
		c2 := cr.Start + crnd[1]

		var o1, o2 int
		for {
			o1 = robjs[cidx]
			o2 = robjs[cidx+1]
			cidx += 2
			if cidx >= nobj {
				rand.Shuffle(nobj, func(i, j int) {
					robjs[i], robjs[j] = robjs[j], robjs[i]
				})
				cidx = 0
			}
			o1nm := Objs[o1]
			o2nm := Objs[o2]
			o1cat := Cats[o1nm]
			o2cat := Cats[o2nm]
			if o1cat == o2cat {
				continue
			}
			break
		}

		if lr == 0 {
			oa1dx, oa2dx, ob1dx, ob2dx = c1, c2, o1, o2
		} else {
			oa1dx, oa2dx, ob1dx, ob2dx = o1, o2, c1, c2
		}

		pcnt.Set([]int{oa1dx, oa2dx}, pcnt.Value([]int{oa1dx, oa2dx})+1)
		pcnt.Set([]int{oa2dx, oa1dx}, pcnt.Value([]int{oa2dx, oa1dx})+1)
		pcnt.Set([]int{ob1dx, ob2dx}, pcnt.Value([]int{ob1dx, ob2dx})+1)
		pcnt.Set([]int{ob2dx, ob1dx}, pcnt.Value([]int{ob2dx, ob1dx})+1)

		oa1nm, oa1fnm := RndFnm(objfiles, oa1dx)
		oa2nm, oa2fnm := RndFnm(objfiles, oa2dx)
		ob1nm, ob1fnm := RndFnm(objfiles, ob1dx)
		ob2nm, ob2fnm := RndFnm(objfiles, ob2dx)

		dst := CmpImages(inpath, oa1fnm, oa2fnm, ob1fnm, ob2fnm)

		ofn := fmt.Sprintf("trial_%d.jpg", trl)

		fmt.Fprintf(fnlist, "%s\n", ofn)
		fmt.Fprintf(trllist, "%s,%s,%s,%s,%s\n", ofn, oa1nm, oa2nm, ob1nm, ob2nm)

		ofpth := filepath.Join(outpath, ofn)
		if err := imgio.Save(ofpth, dst, imgio.JPEGEncoder(95)); err != nil {
			panic(err)
		}
	}
	fnlist.Close()
	trllist.Close()

	fmt.Println(pcnt.String())
}

// ntrials: 1800 = 60 per minute * 30 minutes; 900 = 30 per minute * 30 minutes
// so 800 seems reasonable

func main() {
	// pureRnd := true
	// if pureRnd {
	PurelyRandomPairs("/Users/oreilly/deep-obj-cat-shape-imgs", "/Users/oreilly/deep-obj-cat-shape-cmp-trls", 800)
	// } else {
	CategPairs("/Users/oreilly/deep-obj-cat-shape-imgs", "/Users/oreilly/deep-obj-cat-shape-cmp-trls-cat", 800)
	// }
}

// PurelyRandomPairs output counts (n = 800)

// Int64: [20, 20]
// [0,0]: 0 10 11 6 10 11 7 7 7 10 9 7 13 7 8 6 8 13 5 5
// [1,0]: 10 0 8 6 6 15 9 9 8 5 6 6 8 11 11 12 7 8 2 13
// [2,0]: 11 8 0 5 11 5 4 8 16 7 7 6 6 8 9 13 12 9 10 5
// [3,0]: 6 6 5 0 9 6 8 13 5 7 12 12 9 9 11 3 5 11 12 11
// [4,0]: 10 6 11 9 0 2 5 17 9 9 8 12 10 7 5 12 3 6 11 8
// [5,0]: 11 15 5 6 2 0 16 11 8 7 10 7 7 7 7 9 11 6 7 8
// [6,0]: 7 9 4 8 5 16 0 5 9 5 6 9 15 4 16 9 7 3 12 11
// [7,0]: 7 9 8 13 17 11 5 0 8 8 6 6 7 10 8 8 10 5 7 7
// [8,0]: 7 8 16 5 9 8 9 8 0 6 8 11 5 13 4 10 10 8 6 9
// [9,0]: 10 5 7 7 9 7 5 8 6 0 8 16 6 10 10 8 12 12 8 6
// [10,0]: 9 6 7 12 8 10 6 6 8 8 0 5 10 8 8 15 9 4 7 14
// [11,0]: 7 6 6 12 12 7 9 6 11 16 5 0 9 11 5 5 6 11 8 8
// [12,0]: 13 8 6 9 10 7 15 7 5 6 10 9 0 7 3 6 11 11 8 9
// [13,0]: 7 11 8 9 7 7 4 10 13 10 8 11 7 0 8 7 6 11 8 8
// [14,0]: 8 11 9 11 5 7 16 8 4 10 8 5 3 8 0 3 10 7 17 10
// [15,0]: 6 12 13 3 12 9 9 8 10 8 15 5 6 7 3 0 10 7 8 9
// [16,0]: 8 7 12 5 3 11 7 10 10 12 9 6 11 6 10 10 0 8 11 4
// [17,0]: 13 8 9 11 6 6 3 5 8 12 4 11 11 11 7 7 8 0 9 11
// [18,0]: 5 2 10 12 11 7 12 7 6 8 7 8 8 8 17 8 11 9 0 4
// [19,0]: 5 13 5 11 8 8 11 7 9 6 14 8 9 8 10 9 4 11 4 0

// CategPairs output counts (n = 800) -- much higher along the diagonal

// Int64: [20, 20]
// [0,0]: 0 10 23 10 29 3 3 5 3 1 7 7 3 6 6 3 8 8 6 5
// [1,0]: 10 0 21 15 22 2 3 6 6 6 7 5 5 5 8 5 1 5 4 6
// [2,0]: 23 21 0 9 9 7 6 11 7 6 6 3 7 6 2 3 4 4 6 2
// [3,0]: 10 15 9 0 17 5 3 9 5 3 6 5 11 2 5 10 1 3 3 4
// [4,0]: 29 22 9 17 0 6 4 3 9 5 5 5 6 1 3 6 4 6 3 6
// [5,0]: 3 2 7 5 6 0 65 43 4 4 5 5 2 6 5 7 7 3 6 8
// [6,0]: 3 3 6 3 4 65 0 52 2 9 6 11 8 7 7 3 3 3 4 5
// [7,0]: 5 6 11 9 3 43 52 0 4 5 3 3 4 8 3 4 4 5 2 5
// [8,0]: 3 6 7 5 9 4 2 4 0 26 20 27 3 5 3 3 5 4 7 6
// [9,0]: 1 6 6 3 5 4 9 5 26 0 28 22 2 6 5 6 6 9 6 0
// [10,0]: 7 7 6 6 5 5 6 3 20 28 0 24 7 4 4 8 5 1 7 5
// [11,0]: 7 5 3 5 5 5 11 3 27 22 24 0 3 6 6 8 6 4 2 7
// [12,0]: 3 5 7 11 6 2 8 4 3 2 7 3 0 28 24 35 7 6 5 3
// [13,0]: 6 5 6 2 1 6 7 8 5 6 4 6 28 0 21 30 6 2 5 6
// [14,0]: 6 8 2 5 3 5 7 3 3 5 4 6 24 21 0 28 8 9 5 5
// [15,0]: 3 5 3 10 6 7 3 4 3 6 8 8 35 30 28 0 4 4 5 6
// [16,0]: 8 1 4 1 4 7 3 4 5 6 5 6 7 6 8 4 0 23 34 35
// [17,0]: 8 5 4 3 6 3 3 5 4 9 1 4 6 2 9 4 23 0 22 26
// [18,0]: 6 4 6 3 3 6 4 2 7 6 7 2 5 5 5 5 34 22 0 22
// [19,0]: 5 6 2 4 6 8 5 5 6 0 5 7 3 6 5 6 35 26 22 0
