// Copyright 2020 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// splittable is an experimental program to split a table into
// individual cells suitable for OCR
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"

	"rescribe.xyz/integralimg"
)

const usage = `Usage: splittable [-t thresh] [-w winsize] inimg outbase

splittable is an experimental program to split a table into individual
cells suitable for OCR. It does this by detecting lines. At present it
just detects vertical lines and outputs images for each section
between those lines.

`

// returns the proportion of the given window that is black pixels
func proportion(i integralimg.Image, x int, size int) float64 {
	w := i.GetVerticalWindow(x, size)
	return w.Proportion()
}

// findbestvline goes through every vertical line from x to x+w to
// find the one with the lowest proportion of black pixels.
func findbestvline(img integralimg.Image, x int, w int) int {
	var bestx int
	var best float64

	if w == 1 {
		return x
	}

	right := x + w
	for ; x < right; x++ {
		prop := proportion(img, x, 1)
		if prop > best {
			best = prop
			bestx = x
		}
	}

	return bestx
}

// findvlines finds vertical lines, returning an array of x coordinates
// for each line. It works by moving a window of wsize across the image,
// marking each place where there is a higher proportion of black pixels
// than thresh.
func findvlines(img integralimg.Image, wsize int, thresh float64) []int {
	maxx := img.Bounds().Dx() - 1
	var lines []int

	for x := 0; x < maxx-wsize; x+=wsize {
		if proportion(img, x, wsize) >= thresh {
			l := findbestvline(img, x, wsize)
			lines = append(lines, l)
		}
	}

	return lines
}

func drawsection(img *image.Gray, x1 int, x2 int) *image.Gray {
	b := img.Bounds()
	width := x2-x1
	new := image.NewGray(image.Rect(0, b.Min.Y, width, b.Max.Y))

	for x := 0; x < width; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, img.GrayAt(x1 + x, y))
		}
	}

	return new
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	thresh := flag.Float64("t", 0.85, "Threshold for the proportion of black pixels below which a window is determined to be a line. Higher means fewer lines will be found.")
	wsize := flag.Int("w", 1, "Window size for mask finding algorithm.")
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	defer f.Close()
	if err != nil {
		log.Fatalf("Could not open file %s: %v\n", flag.Arg(0), err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("Could not decode image: %v\n", err)
	}
	b := img.Bounds()
	gray := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(gray, b, img, b.Min, draw.Src)

	integral := integralimg.NewImage(b)
	draw.Draw(integral, b, gray, b.Min, draw.Src)
	vlines := findvlines(*integral, *wsize, *thresh)

	for i, v := range vlines {
		fmt.Printf("line detected at x=%d\n", v)

		if i+1 >= len(vlines) {
			break
		}
		section := drawsection(gray, v, vlines[i+1])

		fn := fmt.Sprintf("%s-%d.png", flag.Arg(1), v)
		f, err = os.Create(fn)
		if err != nil {
			log.Fatalf("Could not create file %s: %v\n", fn, err)
		}
		defer f.Close()
		err := png.Encode(f, section)
		if err != nil {
			log.Fatalf("Could not encode image %s: %v\n", fn, err)
		}
	}


	// TODO: find horizontal lines too
	// TODO: do rotation
	// TODO: output table cells
	// TODO: potentially send cells straight to tesseract
}
