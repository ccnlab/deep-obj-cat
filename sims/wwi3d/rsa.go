// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
	"github.com/emer/etable/simat"
	"github.com/goki/gi/gi"
)

// LbaCats5 is best-fitting 5-category leabra ("Centroid")
var LbaCats5 = map[string]string{
	"banana":      "1-pyramid",
	"layercake":   "1-pyramid",
	"trafficcone": "1-pyramid",
	"sailboat":    "1-pyramid",
	"trex":        "1-pyramid",
	"person":      "2-vertical",
	"guitar":      "2-vertical",
	"tablelamp":   "2-vertical",
	"doorknob":    "3-round",
	"donut":       "3-round",
	"handgun":     "3-round",
	"chair":       "3-round",
	"slrcamera":   "4-box",
	"elephant":    "4-box",
	"piano":       "4-box",
	"fish":        "4-box",
	"car":         "5-horiz",
	"heavycannon": "5-horiz",
	"stapler":     "5-horiz",
	"motorcycle":  "5-horiz",
}

// RSA handles representational similarity analysis
type RSA struct {
	Interval int                      `desc:"how often to run RSA analyses over epochs"`
	Cats     []string                 `desc:"category names for each row of simmat / activation table -- call SetCats"`
	Sims     map[string]*simat.SimMat `desc:"similarity matricies for each layer"`
	V1Sims   []float64                `desc:"similarity for each layer relative to V1"`
	CatDists []float64                `desc:"AvgContrastDist for each layer under LbaCats5 centroid meta categories"`
	Cat5Sims map[string]*simat.SimMat `desc:"similarity matricies for each layer, organized into LbaCats5 and sorted"`
}

// Init initializes maps etc if not done yet
func (rs *RSA) Init(lays []string) {
	if rs.Sims != nil {
		return
	}
	nc := len(lays)
	rs.Sims = make(map[string]*simat.SimMat, nc)
	rs.Cat5Sims = make(map[string]*simat.SimMat, nc)
	rs.V1Sims = make([]float64, nc)
	rs.CatDists = make([]float64, nc)
}

// SetCats sets the categories from given list of category/object_file names
func (rs *RSA) SetCats(objs []string) {
	rs.Cats = make([]string, 0, 20*20)
	for _, ob := range objs {
		cat := strings.Split(ob, "/")[0]
		rs.Cats = append(rs.Cats, cat)
	}
}

func (rs *RSA) SimByName(cn string) *simat.SimMat {
	sm, ok := rs.Sims[cn]
	if !ok || sm == nil {
		sm = &simat.SimMat{}
		rs.Sims[cn] = sm
	}
	return sm
}

func (rs *RSA) Cat5SimByName(cn string) *simat.SimMat {
	sm, ok := rs.Cat5Sims[cn]
	if !ok || sm == nil {
		sm = &simat.SimMat{}
		rs.Cat5Sims[cn] = sm
	}
	return sm
}

// StatsFmActs computes RSA stats from given acts table, for given columns (layer names)
func (rs *RSA) StatsFmActs(acts *etable.Table, lays []string) {
	tick := 2 // use this tick for analyses..
	tix := etable.NewIdxView(acts)
	tix.Filter(func(et *etable.Table, row int) bool {
		tck := int(et.CellFloat("Tick", row))
		return tck == tick
	})

	for _, cn := range lays {
		sm := rs.SimByName(cn)
		rs.SimMatFmActs(sm, tix, cn)
	}

	v1sm := rs.Sims["V1m"]
	v1sm64 := v1sm.Mat.(*etensor.Float64)
	for i, cn := range lays {
		osm := rs.SimByName(cn)

		rs.CatDists[i] = -rs.AvgContrastDist(osm, rs.Cats, LbaCats5)

		if v1sm == osm {
			rs.V1Sims[i] = 1
			continue
		}
		osm64 := osm.Mat.(*etensor.Float64)
		rs.V1Sims[i] = metric.Correlation64(osm64.Values, v1sm64.Values)
	}
	cat5s := []string{"TE"}
	for _, cn := range cat5s {
		sm := rs.SimByName(cn)
		sm5 := rs.Cat5SimByName(cn)
		rs.CatSortSimMat(sm, sm5, rs.Cats, LbaCats5, true, cn+"_LbaCat")
	}
}

// ConfigSimMat sets meta data
func (rs *RSA) ConfigSimMat(sm *simat.SimMat) {
	smat := sm.Mat.(*etensor.Float64)
	smat.SetMetaData("max", "2")
	smat.SetMetaData("min", "0")
	smat.SetMetaData("colormap", "Viridis")
	smat.SetMetaData("grid-fill", "1")
	smat.SetMetaData("dim-extra", "0.5")
}

// SimMatFmActs computes the given SimMat from given acts table (IdxView),
// for given column name.
func (rs *RSA) SimMatFmActs(sm *simat.SimMat, acts *etable.IdxView, colnm string) {
	sm.Init()
	rs.ConfigSimMat(sm)

	sm.TableCol(acts, colnm, "Cat", true, metric.InvCorrelation64)
}

// OpenSimMat opens a saved sim mat for given layer name,
// using given cat strings per row of sim mat
func (rs *RSA) OpenSimMat(laynm string, fname gi.FileName) {
	sm := rs.SimByName(laynm)
	no := len(rs.Cats)
	sm.Init()
	rs.ConfigSimMat(sm)
	smat := sm.Mat.(*etensor.Float64)
	smat.SetShape([]int{no, no}, nil, nil)
	err := etensor.OpenCSV(smat, fname, etable.Tab.Rune())
	if err != nil {
		log.Println(err)
		return
	}
	sm.Rows = simat.BlankRepeat(rs.Cats)
	sm.Cols = sm.Rows
	sm5 := rs.Cat5SimByName(laynm)
	rs.CatSortSimMat(sm, sm5, rs.Cats, LbaCats5, true, laynm+"_LbaCat")
}

// CatSortSimMat takes an input sim matrix and categorizes the items according to given cats
// and then sorts items within that according to their average within - between cat similarity.
// contrast = use within - between metric, otherwise just within
func (rs *RSA) CatSortSimMat(insm *simat.SimMat, osm *simat.SimMat, nms []string, catmap map[string]string, contrast bool, name string) {
	no := len(insm.Rows)
	sch := etable.Schema{
		{"Cat", etensor.STRING, nil, nil},
		{"Dist", etensor.FLOAT64, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, no)
	cats := dt.Cols[0].(*etensor.String).Values
	dists := dt.Cols[1].(*etensor.Float64).Values
	for i, nm := range nms {
		cats[i] = catmap[nm]
	}
	smatv := insm.Mat.(*etensor.Float64).Values
	avgCtrstDist := 0.0
	for ri := 0; ri < no; ri++ {
		roff := ri * no
		aid := 0.0
		ain := 0
		abd := 0.0
		abn := 0
		rc := cats[ri]
		for ci := 0; ci < no; ci++ {
			if ri == ci {
				continue
			}
			cc := cats[ci]
			d := smatv[roff+ci]
			if cc == rc {
				aid += d
				ain++
			} else {
				abd += d
				abn++
			}
		}
		if ain > 0 {
			aid /= float64(ain)
		}
		if abn > 0 {
			abd /= float64(abn)
		}
		dval := aid
		if contrast {
			dval -= abd
		}
		dists[ri] = dval
		avgCtrstDist += (1 - aid) - (1 - abd)
	}
	avgCtrstDist /= float64(no)
	ix := etable.NewIdxView(dt)
	ix.SortColNames([]string{"Cat", "Dist"}, true) // ascending
	osm.Init()
	osm.Mat.CopyShapeFrom(insm.Mat)
	osm.Mat.CopyMetaData(insm.Mat)
	rs.ConfigSimMat(osm)
	omatv := osm.Mat.(*etensor.Float64).Values
	bcols := make([]string, no)
	last := ""
	for sri := 0; sri < no; sri++ {
		sroff := sri * no
		ri := ix.Idxs[sri]
		roff := ri * no
		cat := cats[ri]
		if cat != last {
			bcols[sri] = cat
			last = cat
		}
		// bcols[sri] = nms[ri] // uncomment this to see all the names
		for sci := 0; sci < no; sci++ {
			ci := ix.Idxs[sci]
			d := smatv[roff+ci]
			omatv[sroff+sci] = d
		}
	}
	osm.Rows = bcols
	osm.Cols = bcols
	fmt.Printf("%v  avg contrast dist: %.4f\n", name, avgCtrstDist)
}

// AvgContrastDist computes average contrast dist over given cat map
// nms gives the base category names for each row in the simat, which is
// then used to lookup the meta category in the catmap, which is used
// for determining the within vs. between category status.
func (rs *RSA) AvgContrastDist(insm *simat.SimMat, nms []string, catmap map[string]string) float64 {
	no := len(insm.Rows)
	smatv := insm.Mat.(*etensor.Float64).Values
	avgd := 0.0
	for ri := 0; ri < no; ri++ {
		roff := ri * no
		aid := 0.0
		ain := 0
		abd := 0.0
		abn := 0
		rnm := nms[ri]
		rc := catmap[rnm]
		for ci := 0; ci < no; ci++ {
			if ri == ci {
				continue
			}
			cnm := nms[ci]
			cc := catmap[cnm]
			d := smatv[roff+ci]
			if cc == rc {
				aid += d
				ain++
			} else {
				abd += d
				abn++
			}
		}
		if ain > 0 {
			aid /= float64(ain)
		}
		if abn > 0 {
			abd /= float64(abn)
		}
		avgd += aid - abd
	}
	avgd /= float64(no)
	return avgd
}
