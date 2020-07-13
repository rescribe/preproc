// Copyright 2019-2020 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

// TODO: Integrate all test cases into one struct that has
//       several different tests (horiz only, vert only,
//       both) run on it.

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"

	"rescribe.xyz/integralimg"
)

func TestWipeSides(t *testing.T) {
	cases := []struct {
		orig   string
		golden string
		thresh float64
		wsize  int
	}{
		{"testdata/pg2.png", "testdata/pg2_integralwipesides_t0.02_w5.png", 0.02, 5},
		{"testdata/pg2.png", "testdata/pg2_integralwipesides_t0.05_w5.png", 0.05, 5},
		{"testdata/pg2.png", "testdata/pg2_integralwipesides_t0.05_w25.png", 0.05, 25},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Exact/%s_%0.2f_%d", c.orig, c.thresh, c.wsize), func(t *testing.T) {
			var actual *image.Gray
			orig, err := decode(c.orig)
			if err != nil {
				t.Fatalf("Could not open file %s: %v\n", c.orig, err)
			}
			actual = Wipe(orig, c.wsize, c.thresh, 30)
			if *update {
				f, err := os.Create(c.golden)
				defer f.Close()
				if err != nil {
					t.Fatalf("Could not open file %s to update: %v\n", c.golden, err)
				}
				err = png.Encode(f, actual)
				if err != nil {
					t.Fatalf("Could not encode update of %s: %v\n", c.golden, err)
				}
			}
			golden, err := decode(c.golden)
			if err != nil {
				t.Fatalf("Could not open file %s: %v\n", c.golden, err)
			}
			if !imgsequal(golden, actual) {
				t.Errorf("Processed %s differs to %s\n", c.orig, c.golden)
			}
		})
	}
	leftrightedgecases := []struct {
		filename string
		minleft  int
		maxleft  int
		minright int
		maxright int
		thresh   float64
		wsize    int
	}{
		{"testdata/0002.png", 36, 250, 967, 998, 0.02, 5},
		{"testdata/1727_GREENE_0048.png", 40, 150, 1270, 1281, 0.02, 5},
	}

	for _, c := range leftrightedgecases {
		t.Run(fmt.Sprintf("LeftRightEdge/%s_%0.2f_%d", c.filename, c.thresh, c.wsize), func(t *testing.T) {
			img, err := decode(c.filename)
			if err != nil {
				t.Fatalf("Could not open file %s: %v\n", c.filename, err)
			}
			integral := integralimg.ToIntegralImg(img)
			leftedge, rightedge := findedges(integral, c.wsize, c.thresh)
			if leftedge < c.minleft {
				t.Errorf("Left edge %d < minimum %d", leftedge, c.minleft)
			}
			if leftedge > c.maxleft {
				t.Errorf("Left edge %d > maximum %d", leftedge, c.maxleft)
			}
			if rightedge < c.minright {
				t.Errorf("Right edge %d < minimum %d", rightedge, c.minright)
			}
			if rightedge > c.maxright {
				t.Errorf("Right edge %d > maximum %d", rightedge, c.maxright)
			}
		})
	}

	topbottomedgecases := []struct {
		filename  string
		mintop    int
		maxtop    int
		minbottom int
		maxbottom int
		thresh    float64
		wsize     int
	}{
		{"testdata/1727_GREENE_0048.png", 70, 237, 2204, 2450, 0.005, 120},
		{"testdata/1687_SCHWEITZER_0030.png", 70, 231, 2595, 2770, 0.005, 90},
	}

	for _, c := range topbottomedgecases {
		t.Run(fmt.Sprintf("TopBottomEdge/%s_%0.3f_%d", c.filename, c.thresh, c.wsize), func(t *testing.T) {
			img, err := decode(c.filename)
			if err != nil {
				t.Fatalf("Could not open file %s: %v\n", c.filename, err)
			}
			integral := integralimg.ToIntegralImg(sideways(img))
			topedge, bottomedge := findedges(integral, c.wsize, c.thresh)
			if topedge < c.mintop {
				t.Errorf("Top edge %d < minimum %d", topedge, c.mintop)
			}
			if topedge > c.maxtop {
				t.Errorf("Top edge %d > maximum %d", topedge, c.maxtop)
			}
			if bottomedge < c.minbottom {
				t.Errorf("Bottom edge %d < minimum %d", bottomedge, c.minbottom)
			}
			if bottomedge > c.maxbottom {
				t.Errorf("Bottom edge %d > maximum %d", bottomedge, c.maxbottom)
			}
		})
	}
}
