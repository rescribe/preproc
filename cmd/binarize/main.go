// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// binarize does fast Integral Image sauvola binarisation on an image
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

// TODO: do more testing to see how good this assumption is
func autowsize(bounds image.Rectangle) int {
	return bounds.Dx() / 60
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: binarize [-k num] [-t type] [-w num] inimg outimg\n")
		flag.PrintDefaults()
	}
	wsize := flag.Int("w", 0, "Window size for sauvola algorithm. Set automatically based on resolution if not set.")
	ksize := flag.Float64("k", 0.5, "K for sauvola algorithm. This controls the overall threshold level. Set it lower for very light text (try 0.1 or 0.2).")
	btype := flag.String("t", "binary", "Type of threshold. binary or zeroinv are currently implemented.")
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

	if *wsize == 0 {
		*wsize = autowsize(b)
		log.Printf("Set window size to %d\n", *wsize)
	}

	if *wsize%2 == 0 {
		*wsize++
	}

	// TODO: come up with a way to set a good ksize automatically

	var thresh image.Image
	thresh = preproc.IntegralSauvola(gray, *ksize, *wsize)

	if *btype == "zeroinv" {
		thresh, err = preproc.BinToZeroInv(thresh.(*image.Gray), img.(*image.RGBA))
		if err != nil {
			log.Fatal(err)
		}
	}

	f, err = os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("Could not create file %s: %v\n", flag.Arg(1), err)
	}
	defer f.Close()
	err = png.Encode(f, thresh)
	if err != nil {
		log.Fatalf("Could not encode image: %v\n", err)
	}
}
