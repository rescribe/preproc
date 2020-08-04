// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"
)

func TestBinarization(t *testing.T) {
	cases := []struct {
		name   string
		orig   string
		golden string
		ksize  float64
		wsize  int
	}{
		{"integralsauvola", "testdata/pg1.png", "testdata/pg1_sauvola_k0.5_w41.png", 0.5, 41},
		{"integralsauvola", "testdata/pg1.png", "testdata/pg1_sauvola_k0.5_w19.png", 0.5, 19},
		{"integralsauvola", "testdata/pg1.png", "testdata/pg1_sauvola_k0.3_w19.png", 0.3, 19},
		{"sauvola", "testdata/pg1.png", "testdata/pg1_sauvola_k0.5_w41.png", 0.5, 41},
		{"sauvola", "testdata/pg1.png", "testdata/pg1_sauvola_k0.5_w19.png", 0.5, 19},
		{"sauvola", "testdata/pg1.png", "testdata/pg1_sauvola_k0.3_w19.png", 0.3, 19},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s_%0.1f_%d", c.name, c.ksize, c.wsize), func(t *testing.T) {
			var actual *image.Gray
			orig, err := decode(c.orig)
			if err != nil {
				t.Fatalf("Could not open file %s: %v\n", c.orig, err)
			}
			switch c.name {
			case "integralsauvola":
				actual = IntegralSauvola(orig, c.ksize, c.wsize)
			case "sauvola":
				if !testing.Short() {
					actual = Sauvola(orig, c.ksize, c.wsize)
				} else {
					t.Skip("Skipping long test due to -short flag.\n")
				}
			default:
				t.Fatalf("No method %s\n", c.name)
			}
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
				t.Errorf("Binarized %s differs to %s\n", c.orig, c.golden)
			}
		})
	}
}
