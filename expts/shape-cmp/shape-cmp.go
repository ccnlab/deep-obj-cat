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

	"github.com/anthonynsimon/bild/imgio"
)

func PlaceImage(dst, src image.Image, offset image.Point) {
	r := src.Bounds().Add(offset)
	draw.Draw(dst.(*image.RGBA), r, src, image.ZP, draw.Src)
}

func CmpImages(inpath, ima1, ima2, imb1, imb2 string) image.Image {
	fmt.Printf("inpath: %s ima1: %s  ima2: %s  imb1: %s  imb2: %s\n", inpath, ima1, ima2, imb1, imb2)
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

func RndFnm(objfiles [][]os.FileInfo, objidx int) (string, string) {
	fl := objfiles[objidx]
	idx := rand.Intn(len(fl))
	onm := Objs[objidx]
	return onm, filepath.Join(onm, fl[idx].Name())
}

func DoDir(inpath, outpath string, ntrl int) {
	os.MkdirAll(outpath, 0755)

	fnlist, _ := os.Create(filepath.Join(outpath, "file-list.csv"))
	trllist, _ := os.Create(filepath.Join(outpath, "trial-list.csv"))

	fmt.Fprintf(fnlist, "image_url\n", ofn)

	nobj := len(Objs)

	objfiles := make([][]os.FileInfo, nobj)
	for i, ob := range Objs {
		opth := filepath.Join(inpath, ob)
		objfiles[i], _ = ioutil.ReadDir(opth)
	}

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
}

func main() {
	DoDir("/Users/oreilly/deep-obj-cat-shape-imgs", "/Users/oreilly/deep-obj-cat-shape-cmp-trls", 800)
}
