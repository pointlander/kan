// Copyright 2024 The KAN Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math/cmplx"
	"os"

	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
)

func main() {
	input, err := os.Open("test.jpg")
	if err != nil {
		panic(err)
	}
	defer input.Close()
	img, _, err := image.Decode(input)
	if err != nil {
		panic(err)
	}
	bounds := img.Bounds()
	bounds = image.Rect(0, 0, bounds.Dx(), bounds.Dx())
	gray := image.NewGray(bounds)
	dx := bounds.Dx()
	dy := bounds.Dy()
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}
	mat := dsputils.MakeMatrix(make([]complex128, dx*dy), []int{dx, dy})
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			g := float64(gray.GrayAt(x, y).Y)
			mat.SetValue(complex(g/255, 0), []int{x, y})
		}
	}
	freq := fft.FFTN(mat)
	for y := 0; y < dy; y++ {
		for x := y + 1; x < dx; x++ {
			v1 := freq.Value([]int{x, y})
			v2 := freq.Value([]int{y, x})
			sum := v1 + v2
			freq.SetValue(sum, []int{x, y})
			freq.SetValue(sum, []int{y, x})
		}
	}
	inverse := fft.IFFTN(freq)
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			v := inverse.Value([]int{x, y})
			gray.SetGray(x, y, color.Gray{Y: byte(255 * cmplx.Abs(v))})
		}
	}
	output, err := os.Create("gray.jpg")
	if err != nil {
		panic(err)
	}
	defer output.Close()
	err = png.Encode(output, gray)
	if err != nil {
		panic(err)
	}
}
