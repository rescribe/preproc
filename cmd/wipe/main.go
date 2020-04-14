// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// wipe wipes sections of an image which are outside of an automatically
// determined content area
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

	"rescribe.xyz/preproc"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: wipe inimg outimg\n")
		fmt.Fprintf(os.Stderr, "Wipes the sections of an image which are outside the content area.\n")
		flag.PrintDefaults()
	}
	min := flag.Int("hm", 30, "Minimum percentage of the image width for the content width calculation to be considered valid.")
	thresh := flag.Float64("ht", 0.05, "Threshold for the proportion of black pixels below which a window is determined to be the edge. Higher means more aggressive wiping.")
	wsize := flag.Int("hw", 5, "Window size for mask finding algorithm.")
	vmin := flag.Int("vm", 30, "Minimum percentage of the image height for the content width calculation to be considered valid.")
	vthresh := flag.Float64("vt", 0.005, "Threshold for the proportion of black pixels below which a vertical wipe window is determined to be the edge. Higher means more aggressive wiping.")
	vwsize := flag.Int("vw", 120, "Window size for vertical mask finding algorithm. Should be set to approximately line height + largest expected gap.")
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

	sidesdone := preproc.Wipe(gray, *wsize, *thresh, *min)
	clean := preproc.VWipe(sidesdone, *vwsize, *vthresh, *vmin)

	f, err = os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("Could not create file %s: %v\n", flag.Arg(1), err)
	}
	defer f.Close()
	err = png.Encode(f, clean)
	if err != nil {
		log.Fatalf("Could not encode image: %v\n", err)
	}
}
