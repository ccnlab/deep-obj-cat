// Copyright (c) 2020, The CCNLab Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objenv

import (
	"testing"

	"github.com/emer/etable/etable"
	"github.com/goki/gi/gi"
)

func TestSaccade(t *testing.T) {
	sac := Saccade{}
	sac.Defaults()
	sac.Init()

	sac.AddRows = true

	for s := 0; s < 16; s++ {
		sac.Step()
	}
	sac.Table.SaveCSV(gi.FileName("sac_test.tsv"), etable.Tab, true)
}
