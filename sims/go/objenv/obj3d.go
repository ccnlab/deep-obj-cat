// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"

	"github.com/emer/etable/etable"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/dirs"
)

// Obj3D implements management of lists of 3D objects, organized in directories
// that provide the category names.
type Obj3D struct {
	Path          string        `desc:"path to object files -- this should point to a directory that has subdirectories that then have .obj files in them"`
	NTestPerCat   int           `desc:"number of testing items per category"`
	Cats          []string      `desc:"list of object categories (directory name)"`
	ObjFilesAll   [][]string    `desc:"full list of objects, organized by category (directory) and then filename"`
	ObjFilesTrain [][]string    `desc:"list of training objects, organized by category (directory) and then filename"`
	ObjFilesTest  [][]string    `desc:"list of testing objects, organized by category (directory) and then filename"`
	ObjCatProps   *etable.Table `desc:"properties for each object category, used for rendering: z_offset is extra offset to add to z-dimension (depth), for objects that are larger than others in total size, despite normalizing along individual axes.  y_rot_mirror is 1 (true) if the object has rough mirror symmetry when rotated around the Y (vertical) axis 180 deg."`
	FlatAll       []string      `desc:"flat list of all objects, as cat/filename.obj -- Flats() makes from above"`
	FlatTrain     []string      `desc:"flat list of all training objects, as cat/filename.obj -- Flats() makes from above"`
	FlatTest      []string      `desc:"flat list of all testing objects, as cat/filename.obj -- Flats() makes from above"`
}

// OpenPath opens list of objs at given path
func (ob *Obj3D) OpenPath(path string) error {
	ob.Path = path
	return ob.Open()
}

// Open opens at Path
func (ob *Obj3D) Open() error {
	ob.Cats = dirs.Dirs(ob.Path)
	nc := len(ob.Cats)
	if nc == 0 {
		err := fmt.Errorf("Obj3D.Open() -- no directories for categories in: %s", ob.Path)
		log.Println(err)
		return err
	}
	ob.ObjFilesAll = make([][]string, nc)
	for ci := nc - 1; ci >= 0; ci-- {
		cat := ob.Cats[ci]
		cp := filepath.Join(ob.Path, cat)
		fls := dirs.ExtFileNames(cp, []string{".obj"})
		if len(fls) == 0 {
			ob.Cats = append(ob.Cats[:ci], ob.Cats[ci+1:]...)
			ob.ObjFilesAll = append(ob.ObjFilesAll[:ci], ob.ObjFilesAll[ci+1:]...)
			continue
		}
		ob.ObjFilesAll[ci] = fls
	}
	ob.Split()
	return nil
}

// Split does the train / test split
func (ob *Obj3D) Split() {
	nc := len(ob.ObjFilesAll)
	ob.ObjFilesTrain = make([][]string, nc)
	ob.ObjFilesTest = make([][]string, nc)
	for ci, fls := range ob.ObjFilesAll {
		nitm := len(fls)
		ntst := ob.NTestPerCat
		if ntst >= nitm {
			ntst = nitm / 2
		}
		ntrn := nitm - ntst
		slist := rand.Perm(nitm)
		for i := 0; i < ntrn; i++ {
			ob.ObjFilesTrain[ci] = append(ob.ObjFilesTrain[ci], fls[slist[i]])
		}
		for i := ntrn; i < nitm; i++ {
			ob.ObjFilesTest[ci] = append(ob.ObjFilesTest[ci], fls[slist[i]])
		}
	}
	ob.Flats()
}

// OpenCatProps opens a table of category properties used in rendering
// typically expects z_offset and y_rot_mirror columns, but others can be added
// for ad-hoc uses beyond (or instead) of those.
func (ob *Obj3D) OpenCatProps(fname string) error {
	ob.ObjCatProps = &etable.Table{}
	return ob.ObjCatProps.OpenCSV(gi.FileName(fname), etable.Comma)
}

// SelectCats filters the list of objs to those within given list of categories.
func (ob *Obj3D) SelectCats(cats []string) {
	nc := len(ob.Cats)
	for ci := nc - 1; ci >= 0; ci-- {
		cat := ob.Cats[ci]

		sel := false
		for _, cs := range cats {
			if cat == cs {
				sel = true
				break
			}
		}
		if !sel {
			ob.Cats = append(ob.Cats[:ci], ob.Cats[ci+1:]...)
			ob.ObjFilesAll = append(ob.ObjFilesAll[:ci], ob.ObjFilesAll[ci+1:]...)
			ob.ObjFilesTrain = append(ob.ObjFilesTrain[:ci], ob.ObjFilesTrain[ci+1:]...)
			ob.ObjFilesTest = append(ob.ObjFilesTest[:ci], ob.ObjFilesTest[ci+1:]...)
		}
	}
	ob.Flats()
}

// DeleteCats filters the list of objs to exclude those within given list of categories.
func (ob *Obj3D) DeleteCats(cats []string) {
	nc := len(ob.Cats)
	for ci := nc - 1; ci >= 0; ci-- {
		cat := ob.Cats[ci]

		del := false
		for _, cs := range cats {
			if cat == cs {
				del = true
				break
			}
		}
		if del {
			ob.Cats = append(ob.Cats[:ci], ob.Cats[ci+1:]...)
			ob.ObjFilesAll = append(ob.ObjFilesAll[:ci], ob.ObjFilesAll[ci+1:]...)
			ob.ObjFilesTrain = append(ob.ObjFilesTrain[:ci], ob.ObjFilesTrain[ci+1:]...)
			ob.ObjFilesTest = append(ob.ObjFilesTest[:ci], ob.ObjFilesTest[ci+1:]...)
		}
	}
	ob.Flats()
}

// Flats generates flat lists from categorized lists, in form categ/fname.obj
func (ob *Obj3D) Flats() {
	ob.FlatAll = ob.FlatImpl(ob.ObjFilesAll)
	ob.FlatTrain = ob.FlatImpl(ob.ObjFilesTrain)
	ob.FlatTest = ob.FlatImpl(ob.ObjFilesTest)
}

// FlatImpl generates flat lists from categorized lists, in form categ/fname.obj
func (ob *Obj3D) FlatImpl(objs [][]string) []string {
	var flat []string
	for ci, fls := range objs {
		cat := ob.Cats[ci]
		for _, fn := range fls {
			flat = append(flat, cat+"/"+fn)
		}
	}
	return flat
}
