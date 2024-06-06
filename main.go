// Copyright 2024 The KAN Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
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
	gray := image.NewGray(bounds)
	dx := bounds.Dx()
	dy := bounds.Dy()
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}
	mat := dsputils.MakeMatrix(make([]complex128, 64*((dx/8)*(dy/8))), []int{64, dx / 8, dy / 8})
	for x := 0; x < dx; x += 8 {
		for y := 0; y < dy; y += 8 {
			for a := 0; a < 8; a++ {
				for b := 0; b < 8; b++ {
					g := float64(gray.GrayAt(x+a, y+b).Y)
					mat.SetValue(complex(g/255, 0), []int{a + 8*b, x / 8, y / 8})
				}
			}
		}
	}
	freq := fft.FFTN(mat)
	mat2 := dsputils.MakeMatrix(make([]complex128, 64*((dx/8)*(dy/8))), []int{64, dx / 8, dy / 8})
	for x := 0; x < dx/8; x++ {
		for y := 0; y < dy/8; y++ {
			value := freq.Value([]int{0, x, y})
			mat2.SetValue(value, []int{0, x, y})
		}
	}
	inverse := fft.IFFTN(mat2)
	out := image.NewGray(image.Rect(0, 0, dx/8, dy/8))
	min, max := make([]float64, 64), make([]float64, 64)
	for i := range min {
		min[i] = 255
	}
	for x := 0; x < dx/8; x++ {
		for y := 0; y < dy/8; y++ {
			v := inverse.Value([]int{0, x, y})
			value := 255 * cmplx.Abs(v)
			for i := 0; i < 64; i++ {
				vv := inverse.Value([]int{i, x, y})
				vvalue := 255 * cmplx.Abs(vv)
				if value != vvalue {
					fmt.Println("not the same")
				}
				if vvalue < min[i] {
					min[i] = vvalue
				}
				if vvalue > max[i] {
					max[i] = vvalue
				}
			}
			out.SetGray(x, y, color.Gray{Y: byte(value)})
		}
	}
	fmt.Println(min)
	fmt.Println(max)
	output, err := os.Create("gray.jpg")
	if err != nil {
		panic(err)
	}
	defer output.Close()
	err = png.Encode(output, out)
	if err != nil {
		panic(err)
	}
}
