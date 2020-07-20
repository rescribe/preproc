// Copyright 2020 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// pggraph creates a graph showing the proportion of black pixels
// for slices through an image. This is useful to determine
// appropriate parameters to pass to the wipe functions in this
// module.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	_ "image/jpeg"
	"io"
	"log"
	"os"
	"sort"

	"rescribe.xyz/integralimg"
	chart "github.com/wcharczuk/go-chart"
)

const usage = `Usage: pggraph [-vertical] [-width] inimg graphname

Creates a graph showing the proportion of black pixels
for slices through an image. This is useful to determine
appropriate parameters to pass to the wipe functions in
this module.
`

const middlePercent = 20
const tickEvery = 30

// sideways flips an image sideways
func sideways(img *image.Gray) *image.Gray {
	b := img.Bounds()
	newb := image.Rect(b.Min.Y, b.Min.X, b.Max.Y, b.Max.X)
	new := image.NewGray(newb)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(y, x, img.GrayAt(x, y))
		}
	}
	return new
}

func graph(title string, points map[int]float64, w io.Writer) error {
	var xvals, yvals []float64
	var xs []int
	var midxvals, midyvals []float64
	var midxs []int

	if len(points) < 2 {
		return fmt.Errorf("Not enough points to graph, only %d\n", len(points))
	}

	for x, _ := range points {
		xs = append(xs, x)
	}
	sort.Ints(xs)

	for _, x := range xs {
		xvals = append(xvals, float64(x))
		yvals = append(yvals, points[x])
	}

	mainSeries := chart.ContinuousSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorBlue,
			FillColor:   chart.ColorAlternateBlue,
		},
		XValues: xvals,
		YValues: yvals,
	}

	numAroundMiddle := int(middlePercent / 2 * float64(len(xs)) / float64(100))
	if numAroundMiddle > 1 {
		for i := (len(xs) / 2) - numAroundMiddle; i < (len(xs) / 2) + numAroundMiddle; i++ {
			midxs = append(midxs, xs[i])
		}
	} else {
		midxs = xs
	}

	for _, x := range midxs {
		midxvals = append(midxvals, float64(x))
		midyvals = append(midyvals, points[x])
	}

	middleSeries := chart.ContinuousSeries{
		XValues: midxvals,
		YValues: midyvals,
	}

	minSeries := &chart.MinSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: middleSeries,
	}

	maxSeries := &chart.MaxSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: middleSeries,
	}

	width := xs[1]
	var ticks []chart.Tick
	// if width is larger than tickEvery, just have a tick for each point,
	// otherwise have a tick every tickEvery period
	if width > tickEvery {
		for _, v := range xs {
			ticks = append(ticks, chart.Tick{float64(v), fmt.Sprintf("%d", v)})
		}
	} else {
		for i := 0; i < len(xs) - 1; i += tickEvery {
			ticks = append(ticks, chart.Tick{float64(xs[i]), fmt.Sprintf("%d", xs[i])})
		}
		if len(ticks) > 1 {
			lastx := xs[len(xs) - 1]
			ticks[len(ticks) - 1] = chart.Tick{float64(lastx), fmt.Sprintf("%d", lastx)}
		}
	}

	graph := chart.Chart{
		Title: title,
		Width: 1920,
		Height: 800,
		XAxis: chart.XAxis{
			Name: "Pixel number",
			Ticks: ticks,
		},
		YAxis: chart.YAxis{
			Name: "Proportion of black pixels",
		},
		Series: []chart.Series{
			mainSeries,
			minSeries,
			maxSeries,
			chart.LastValueAnnotationSeries(minSeries),
			chart.LastValueAnnotationSeries(maxSeries),
		},
	}

	return graph.Render(chart.PNG, w)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	vertical := flag.Bool("vertical", false, "Slice image vertically (from top to bottom) rather than horizontally")
	width := flag.Int("width", 5, "Width of slice in pixels (height if in vertical mode)")
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
	if *vertical {
		gray = sideways(gray)
	}
	integral := integralimg.ToIntegralImg(gray)

	points := make(map[int]float64)
	maxx := len(integral[0]) -1
	for x := 0; x + *width < maxx; x += *width {
		w := integral.GetVerticalWindow(x, *width)
		points[x] = w.Proportion()
	}

	f, err = os.Create(flag.Arg(1))
	defer f.Close()
	if err != nil {
		log.Fatalf("Could not create file %s: %v\n", flag.Arg(1), err)
	}

	title := fmt.Sprintf("Proportion of black pixels over %s (width %d)", flag.Arg(0), *width)
	if *vertical {
		title += " (vertical)"
	}
	err = graph(title, points, f)
	if err != nil {
		log.Fatalf("Could not create graph: %v\n", err)
	}
}
