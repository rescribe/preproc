// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// wipe wipes sections of an image which are outside of an automatically
// determined content area
package main

import (
	"flag"
	"fmt"
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
	hmin := flag.Int("hm", 30, "Minimum percentage of the image width for the content width calculation to be considered valid.")
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

	err := preproc.WipeFile(flag.Arg(0), flag.Arg(1), *wsize, *thresh, *hmin, *vwsize, *vthresh, *vmin)
	if err != nil {
		log.Fatalf("Failed to wipe image: %v\n", err)
	}
}
