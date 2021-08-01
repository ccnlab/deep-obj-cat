// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/goki/mat32"
)

func TestXYPolar(t *testing.T) {
	tol := float32(1.0e-4)
	var ang float32
	for ang = -179.99; ang <= 180; ang += 10 {
		plr := mat32.Vec2{ang, .5}
		xy := PolarToXY(plr)
		plr2 := XYToPolar(xy)
		// fmt.Printf("plr: %v  xy: %v  plr2: %v\n", plr, xy, plr2)
		if mat32.Abs(plr2.X-plr.X) > tol {
			t.Errorf("plr ang err: %g != %g\n", plr2.X, plr.X)
		}
		if mat32.Abs(plr2.Y-plr.Y) > tol {
			t.Errorf("plr dist err: %g != %g\n", plr2.Y, plr.Y)
		}
	}
}
