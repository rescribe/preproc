// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

// TODO: optionally return the edges chosen

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"

	"rescribe.xyz/integralimg"
)

// returns the proportion of the given window that is black pixels
func proportion(i ImageWindower, x int, size int) float64 {
	w := i.GetVerticalWindow(x, size)
	return w.Proportion()
}

// findbestedge goes through every vertical line from x to x+w to
// find the one with the lowest proportion of black pixels.
// if there are multiple lines with the same proportion (e.g. zero),
// choose the middle one.
func findbestedge(img ImageWindower, x int, w int) int {
	var best float64
	var bestxs []int

	if w == 1 {
		return x
	}

	best = 100

	right := x + w
	for ; x < right; x++ {
		prop := proportion(img, x, 1)
		if prop < best {
			bestxs = make([]int, 0)
			best = prop
		}
		if prop == best {
			bestxs = append(bestxs, x)
		}
	}
	middlex := bestxs[len(bestxs)/2]

	return middlex
}

// findedges finds the edges of the main content, by moving a window of wsize
// from near the middle of the image to the left and right, stopping when it reaches
// a point at which there is a lower proportion of black pixels than thresh.
func findedges(img ImageWindower, wsize int, thresh float64) (int, int) {
	maxx := img.Bounds().Dx() - 1
	var lowedge, highedge int = 0, maxx

	// don't start at the middle, as this will fail for 2 column layouts,
	// start 10% left or right of the middle
	notcentre := maxx / 10

	for x := maxx/2 + notcentre; x < maxx-wsize; x++ {
		if proportion(img, x, wsize) <= thresh {
			highedge = findbestedge(img, x, wsize)
			break
		}
	}

	for x := maxx/2 - notcentre; x > 0; x-- {
		if proportion(img, x, wsize) <= thresh {
			lowedge = findbestedge(img, x, wsize)
			break
		}
	}

	return lowedge, highedge
}

// findedgesOutin finds the edges of the main content as findedges does,
// but working from the outside of the image inwards, rather than from the
// middle outwards.
// TODO: test what difference this makes
func findedgesOutin(img ImageWindower, wsize int, thresh float64) (int, int) {
	maxx := img.Bounds().Dx() - 1
	var lowedge, highedge int = 0, maxx

	for x := maxx-wsize; x > 0; x-- {
		if proportion(img, x, wsize) > thresh {
			highedge = findbestedge(img, x, wsize)
			break
		}
	}

	for x := 0; x < maxx-wsize; x++ {
		if proportion(img, x, wsize) > thresh {
			lowedge = findbestedge(img, x, wsize)
			break
		}
	}

	return lowedge, highedge
}

// wipesides fills the sections of image not within the boundaries
// of lowedge and highedge with white
func wipesides(img *image.Gray, lowedge int, highedge int) *image.Gray {
	b := img.Bounds()
	new := image.NewGray(b)

	// set left edge white
	for x := b.Min.X; x < lowedge; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, color.Gray{255})
		}
	}
	// copy middle
	for x := lowedge; x < highedge; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, img.GrayAt(x, y))
		}
	}
	// set right edge white
	for x := highedge; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, color.Gray{255})
		}
	}

	return new
}

// toonarrow checks whether the area between lowedge and highedge is
// less than min % of the total image width
func toonarrow(img *image.Gray, lowedge int, highedge int, min int) bool {
	b := img.Bounds()
	imgw := b.Max.X - b.Min.X
	wipew := highedge - lowedge
	if float64(wipew)/float64(imgw)*100 < float64(min) {
		return true
	}
	return false
}

// sideways flips an image sideways
func sideways(img image.Image) *image.Gray {
	b := img.Bounds()
	newb := image.Rect(b.Min.Y, b.Min.X, b.Max.Y, b.Max.X)
	new := image.NewGray(newb)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			c := img.At(x, y)
			new.SetGray(y, x, color.GrayModel.Convert(c).(color.Gray))
		}
	}
	return new
}

// Wipe fills the sections of image which fall outside the content
// area with white, providing the content area is above min %
func Wipe(img *image.Gray, wsize int, thresh float64, min int) *image.Gray {
	b := img.Bounds()
	intImg := integralimg.NewImage(b)
	draw.Draw(intImg, b, img, b.Min, draw.Src)
	lowedge, highedge := findedges(*intImg, wsize, thresh)
	if toonarrow(img, lowedge, highedge, min) {
		return img
	}
	return wipesides(img, lowedge, highedge)
}

// VWipe fills the sections of image which fall outside the vertical
// content area with white, providing the content area is above min %
func VWipe(img *image.Gray, wsize int, thresh float64, min int) *image.Gray {
	rotimg := sideways(img)
	b := rotimg.Bounds()
	intImg := integralimg.NewImage(b)
	draw.Draw(intImg, b, rotimg, b.Min, draw.Src)
	// TODO: test whether there are any places where Outin makes a real difference
	lowedge, highedge:= findedgesOutin(*intImg, wsize, thresh)
	if toonarrow(img, lowedge, highedge, min) {
		return img
	}
	wiped := wipesides(rotimg, lowedge, highedge)
	return sideways(wiped)
}

// WipeFile wipes an image file, filling the sections of the image
// which fall outside the content area with white, providing the
// content area is above min %.
// inPath: path of the input image.
// outPath: path to save the output image.
// hwsize: window size for horizontal wipe algorithm.
// hthresh: threshold for horizontal wipe algorithm.
// hmin: minimum % of content area width to consider valid.
// vwsize: window size for vertical wipe algorithm.
// vthresh: threshold for vertical wipe algorithm.
// vmin: minimum % of content area height to consider valid.
func WipeFile(inPath string, outPath string, hwsize int, hthresh float64, hmin int, vwsize int, vthresh float64, vmin int) error {
	f, err := os.Open(inPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not open file %s: %v", inPath, err))
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not decode image: %v", err))
	}
	b := img.Bounds()
	gray := image.NewGray(b)
	draw.Draw(gray, b, img, b.Min, draw.Src)

	vclean := VWipe(gray, vwsize, vthresh, vmin)
	clean := Wipe(vclean, hwsize, hthresh, hmin)

	f, err = os.Create(outPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create file %s: %v", outPath, err))
	}
	defer f.Close()
	err = png.Encode(f, clean)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not encode image: %v", err))
	}
	return nil
}
