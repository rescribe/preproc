// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// preproc runs binarisation and wipe preprocessing on an image
package main

// TODO: come up with a way to set a good ksize automatically

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
		fmt.Fprintf(os.Stderr, "Usage: preproc [-bt bintype] [-bw winsize] [-k num] [-m minperc] [-nowipe] [-wt wipethresh] [-ws wipesize] inimg outimg\n")
		fmt.Fprintf(os.Stderr, "Binarize and preprocess an image\n")
		flag.PrintDefaults()
	}
	binwsize := flag.Int("bw", 0, "Window size for sauvola binarization algorithm. Set automatically based on resolution if not set.")
	ksize := flag.Float64("k", 0.5, "K for sauvola binarization algorithm. This controls the overall threshold level. Set it lower for very light text (try 0.1 or 0.2).")
	btype := flag.String("bt", "binary", "Type of binarization threshold. binary or zeroinv are currently implemented.")
	min := flag.Int("m", 30, "Minimum percentage of the image width for the content width calculation to be considered valid.")
	nowipe := flag.Bool("nowipe", false, "Disable wiping completely.")
	wipewsize := flag.Int("ws", 5, "Window size for wiping algorithm.")
	thresh := flag.Float64("wt", 0.05, "Threshold for the wiping algorithm to determine the proportion of black pixels below which a window is determined to be the edge.")
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

	if *binwsize == 0 {
		*binwsize = autowsize(b)
	}

	if *binwsize%2 == 0 {
		*binwsize++
	}

	log.Print("Binarising")
	var clean, threshimg, vclean image.Image
	threshimg = preproc.IntegralSauvola(gray, *ksize, *binwsize)

	if *btype == "zeroinv" {
		threshimg, err = preproc.BinToZeroInv(threshimg.(*image.Gray), img.(*image.RGBA))
		if err != nil {
			log.Fatal(err)
		}
	}

	if !*nowipe {
		log.Print("Wiping sides")
		vclean = preproc.VWipe(threshimg.(*image.Gray), *vwsize, *vthresh, *vmin)
		clean = preproc.Wipe(vclean.(*image.Gray), *wipewsize, *thresh, *min)
	} else {
		clean = threshimg
	}

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
