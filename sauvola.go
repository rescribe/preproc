// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

import (
	"image"
	"image/color"

	"rescribe.xyz/integralimg"
)

// Implements Sauvola's algorithm for text binarization, see paper
// "Adaptive document image binarization" (2000)
func Sauvola(img *image.Gray, ksize float64, windowsize int) *image.Gray {
	b := img.Bounds()
	new := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			window := surrounding(img, x, y, windowsize)
			m, dev := meanstddev(window)
			threshold := m * (1 + ksize*((dev/128)-1))
			if img.GrayAt(x, y).Y < uint8(threshold) {
				new.SetGray(x, y, color.Gray{0})
			} else {
				new.SetGray(x, y, color.Gray{255})
			}
		}
	}

	return new
}

// Implements Sauvola's algorithm using Integral Images, see paper
// "Efficient Implementation of Local Adaptive Thresholding Techniques Using Integral Images"
// and
// https://stackoverflow.com/questions/13110733/computing-image-integral
func IntegralSauvola(img *image.Gray, ksize float64, windowsize int) *image.Gray {
	b := img.Bounds()
	new := image.NewGray(b)

	integrals := integralimg.ToAllIntegralImg(img)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			m, dev := integrals.MeanStdDevWindow(x, y, windowsize)
			threshold := m * (1 + ksize*((dev/128)-1))
			if img.GrayAt(x, y).Y < uint8(threshold) {
				new.SetGray(x, y, color.Gray{0})
			} else {
				new.SetGray(x, y, color.Gray{255})
			}
		}
	}

	return new
}

// PreCalcedSauvola Implements Sauvola's algorithm using precalculated Integral Images
func PreCalcedSauvola(integrals integralimg.WithSq, img *image.Gray, ksize float64, windowsize int) *image.Gray {
	// TODO: have this be the root function that the other two reference
	b := img.Bounds()
	new := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			m, dev := integrals.MeanStdDevWindow(x, y, windowsize)
			threshold := m * (1 + ksize*((dev/128)-1))
			if img.GrayAt(x, y).Y < uint8(threshold) {
				new.SetGray(x, y, color.Gray{0})
			} else {
				new.SetGray(x, y, color.Gray{255})
			}
		}
	}

	return new
}
