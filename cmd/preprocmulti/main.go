// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// preprocmulti runs binarisation with a variety of different binarisation
// levels, preprocessing and saving each version
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
	"rescribe.xyz/integral"
)

// TODO: do more testing to see how good this assumption is
func autowsize(bounds image.Rectangle) int {
	return bounds.Dx() / 60
}

func main() {
	ksizes := []float64{0.1, 0.2, 0.4, 0.5}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: preprocmulti [-bt bintype] [-bw winsize] [-m minperc] [-nowipe] [-ws wipesize] inimg outbase\n")
		fmt.Fprintf(os.Stderr, "Binarize and preprocess an image, with multiple binarisation levels,\n")
		fmt.Fprintf(os.Stderr, "saving images to outbase_bin{k}.png.\n")
		fmt.Fprintf(os.Stderr, "Binarises with these levels for k: %v.\n", ksizes)
		flag.PrintDefaults()
	}
	binwsize := flag.Int("bw", 0, "Window size for sauvola binarization algorithm. Set automatically based on resolution if not set.")
	btype := flag.String("bt", "binary", "Type of binarization threshold. binary or zeroinv are currently implemented.")
	min := flag.Int("m", 30, "Minimum percentage of the image width for the content width calculation to be considered valid.")
	nowipe := flag.Bool("nowipe", false, "Disable wiping completely.")
	wipewsize := flag.Int("ws", 5, "Window size for wiping algorithm.")
	vmin := flag.Int("vm", 30, "Minimum percentage of the image height for the content width calculation to be considered valid.")
	vthresh := flag.Float64("vt", 0.005, "Threshold for the proportion of black pixels below which a vertical wipe window is determined to be the edge. Higher means more aggressive wiping.")
	vwsize := flag.Int("vw", 120, "Window size for vertical mask finding algorithm. Should be set to approximately line height + largest expected gap.")
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Opening %s\n", flag.Arg(0))
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

	if *binwsize == 0 {
		*binwsize = autowsize(b)
	}

	if *binwsize%2 == 0 {
		*binwsize++
	}

	var clean, threshimg, vclean image.Image
	log.Print("Precalculating integral images")
	intImg := integral.NewImage(b)
	draw.Draw(intImg, b, img, b.Min, draw.Src)
	intSqImg := integral.NewSqImage(b)
	draw.Draw(intSqImg, b, img, b.Min, draw.Src)

	for _, k := range ksizes {
		log.Print("Binarising")
		threshimg = preproc.PreCalcedSauvola(*intImg, *intSqImg, img, k, *binwsize)

		if *btype == "zeroinv" {
			threshimg, err = preproc.BinToZeroInv(threshimg.(*image.Gray), img.(*image.RGBA))
			if err != nil {
				log.Fatal(err)
			}
		}

		if !*nowipe {
			log.Print("Wiping sides")
			vclean = preproc.VWipe(threshimg.(*image.Gray), *vwsize, *vthresh, *vmin)
			clean = preproc.Wipe(vclean.(*image.Gray), *wipewsize, k*0.02, *min)
		} else {
			clean = threshimg
		}

		savefn := fmt.Sprintf("%s_bin%0.1f.png", flag.Arg(1), k)
		log.Printf("Saving %s\n", savefn)
		f, err = os.Create(savefn)
		if err != nil {
			log.Fatalf("Could not create file %s: %v\n", savefn, err)
		}
		defer f.Close()
		err = png.Encode(f, clean)
		if err != nil {
			log.Fatalf("Could not encode image: %v\n", err)
		}
	}
}
