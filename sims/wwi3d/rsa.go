// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
	"github.com/emer/etable/simat"
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
	Sims     map[string]*simat.SimMat `desc:"similarity matricies for each layer"`
	V1Sims   []float64                `desc:"similarity for each layer relative to V1"`
	CatDists []float64                `desc:"AvgContrastDist for each layer under LbaCats5 centroid meta categories"`
}

// StatsFmActs computes RSA stats from given acts table, for given columns (layer names)
func (rs *RSA) StatsFmActs(acts *etable.Table, cols []string) {
	nc := len(cols)
	if rs.Sims == nil {
		rs.Sims = make(map[string]*simat.SimMat, nc)
		rs.V1Sims = make([]float64, nc)
		rs.CatDists = make([]float64, nc)
	}
	tick := 2 // use this tick for analyses..
	tix := etable.NewIdxView(acts)
	tix.Filter(func(et *etable.Table, row int) bool {
		tck := int(et.CellFloat("Tick", row))
		return tck == tick
	})

	cats := make([]string, tix.Len())
	for i, row := range tix.Idxs {
		cats[i] = acts.CellString("Cat", row)
	}

	for _, cn := range cols {
		sm, ok := rs.Sims[cn]
		if !ok || sm == nil {
			sm = &simat.SimMat{}
		}
		rs.SimMatFmActs(sm, tix, cn)
		rs.Sims[cn] = sm
	}

	v1sm := rs.Sims["V1m"]
	v1sm64 := v1sm.Mat.(*etensor.Float64)
	for i, cn := range cols {
		osm := rs.Sims[cn]

		rs.CatDists[i] = rs.AvgContrastDist(osm, cats, LbaCats5)

		if v1sm == osm {
			rs.V1Sims[i] = 1
			continue
		}
		osm64 := osm.Mat.(*etensor.Float64)
		rs.V1Sims[i] = metric.Correlation64(osm64.Values, v1sm64.Values)
	}
}

// SimMatFmActs computes the given SimMat from given acts table (IdxView),
// for given column name.
func (rs *RSA) SimMatFmActs(sm *simat.SimMat, acts *etable.IdxView, colnm string) {
	sm.Init()
	smat := sm.Mat.(*etensor.Float64)
	smat.SetMetaData("max", "1.1")
	smat.SetMetaData("min", "0")
	smat.SetMetaData("colormap", "Viridis")
	smat.SetMetaData("grid-fill", "1")
	smat.SetMetaData("dim-extra", "0.5")

	sm.TableCol(acts, colnm, "Cat", true, metric.Correlation64)
}

// CatSortSimMat takes an input sim matrix and categorizes the items according to given cats
// and then sorts items within that according to their average within - between cat similarity
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
