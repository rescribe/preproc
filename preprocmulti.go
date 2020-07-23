// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"

	"rescribe.xyz/integralimg"
)

func autowsize(bounds image.Rectangle) int {
	// TODO: do more testing to see how good this assumption is
	return bounds.Dx() / 60
}

// PreProcMulti binarizes and preprocesses an image with multiple binarisation levels.
// inPath: Path of input image.
// ksizes: Slice of k values to pass to Sauvola algorithm
// binType: Type of binarization threshold. binary or zeroinv are currently implemented.
// binWsize: Window size for sauvola binarization algorithm. Set automatically based on resolution if 0.
// wipe: Whether to wipe (clear sides) the image
// wipeWsize: Window size for wiping algorithm
// wipeMinWidthPerc: Minimum percentage of the image width for the content width calculation to be considered valid
func PreProcMulti(inPath string, ksizes []float64, binType string, binWsize int, wipe bool, wipeWsize int, wipeMinWidthPerc int) ([]string, error) {
	// Note: copied from cmd/preprocmulti/main.go, should think about the best way to organise this code later.
	// TODO: return errors that encapsulate the err describing where it was encountered
	// TODO: come up with a way to set a good ksize automatically

	// Make outBase inPath up to final .
	s := strings.Split(inPath, ".")
	outBase := strings.Join(s[:len(s)-1], "")

	var donePaths []string

	f, err := os.Open(inPath)
	if err != nil {
		return donePaths, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return donePaths, err
	}

	b := img.Bounds()
	if binWsize == 0 {
		binWsize = autowsize(b)
	}

	if binWsize%2 == 0 {
		binWsize++
	}

	intImg := integralimg.NewImage(b)
	draw.Draw(intImg, b, img, b.Min, draw.Src)
	intSqImg := integralimg.NewSqImage(b)
	draw.Draw(intSqImg, b, img, b.Min, draw.Src)

	var clean, threshimg image.Image

	for _, k := range ksizes {
		threshimg = PreCalcedSauvola(*intImg, *intSqImg, img, k, binWsize)

		if binType == "zeroinv" {
			threshimg, err = BinToZeroInv(threshimg.(*image.Gray), img.(*image.RGBA))
			if err != nil {
				return donePaths, err
			}
		}

		if wipe {
			clean = Wipe(threshimg.(*image.Gray), wipeWsize, k*0.02, wipeMinWidthPerc)
		} else {
			clean = threshimg
		}

		savefn := fmt.Sprintf("%s_bin%0.1f.png", outBase, k)
		f, err = os.Create(savefn)
		if err != nil {
			return donePaths, err
		}
		defer f.Close()
		err = png.Encode(f, clean)
		if err != nil {
			return donePaths, err
		}
		donePaths = append(donePaths, savefn)
	}
	return donePaths, nil
}
