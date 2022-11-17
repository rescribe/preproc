// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"rescribe.xyz/integral"
)

// Implements Sauvola's algorithm for text binarization, see paper
// "Adaptive document image binarization" (2000)
func Sauvola(img image.Image, ksize float64, windowsize int) *image.Gray {
	b := img.Bounds()
	gray := image.NewGray(b)
	draw.Draw(gray, b, img, b.Min, draw.Src)
	new := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			window := surrounding(gray, x, y, windowsize)
			m, dev := meanstddev(window)
			threshold := m * (1 + ksize*((dev/128)-1))
			if gray.GrayAt(x, y).Y < uint8(math.Round(threshold)) {
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
func IntegralSauvola(img image.Image, ksize float64, windowsize int) *image.Gray {
	b := img.Bounds()

	intImg := integral.NewImage(b)
	draw.Draw(intImg, b, img, b.Min, draw.Src)
	intSqImg := integral.NewSqImage(b)
	draw.Draw(intSqImg, b, img, b.Min, draw.Src)

	return PreCalcedSauvola(*intImg, *intSqImg, img, ksize, windowsize)
}

// PreCalcedSauvola Implements Sauvola's algorithm using precalculated Integral Images
func PreCalcedSauvola(intImg integral.Image, intSqImg integral.SqImage, img image.Image, ksize float64, windowsize int) *image.Gray {
	b := img.Bounds()
	gray := image.NewGray(b)
	draw.Draw(gray, b, img, b.Min, draw.Src)
	new := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r := centeredRectangle(x, y, windowsize)
			m, dev := integral.MeanStdDev(intImg, intSqImg, r)
			// Divide by 255 to adjust from Gray16 used by integral.Image to 8 bit Gray
			m8 := m / 255
			dev8 := dev / 255
			threshold := m8 * (1 + ksize*((dev8/128)-1))
			if gray.GrayAt(x, y).Y < uint8(math.Round(threshold)) {
				new.SetGray(x, y, color.Gray{0})
			} else {
				new.SetGray(x, y, color.Gray{255})
			}
		}
	}

	return new
}

func centeredRectangle(x, y, size int) image.Rectangle {
	step := size / 2
	return image.Rect(x-step-1, y-step-1, x+step, y+step)
}
