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

// Useses the fft to resize an image
func Resize(gray *image.Gray, xscale, yscale int) *image.Gray {
	bounds := gray.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	mat := dsputils.MakeMatrix(make([]complex128, xscale*yscale*((dx/xscale)*(dy/yscale))),
		[]int{xscale, yscale, dx / xscale, dy / yscale})
	for x := 0; x < dx; x += xscale {
		for y := 0; y < dy; y += 8 {
			for a := 0; a < xscale; a++ {
				for b := 0; b < 8; b++ {
					g := float64(gray.GrayAt(x+a, y+b).Y)
					mat.SetValue(complex(g/255, 0), []int{a, b, x / xscale, y / 8})
				}
			}
		}
	}
	freq := fft.FFTN(mat)
	mat2 := dsputils.MakeMatrix(make([]complex128, xscale*yscale*((dx/xscale)*(dy/yscale))),
		[]int{xscale, yscale, dx / xscale, dy / yscale})
	for x := 0; x < dx/xscale; x++ {
		for y := 0; y < dy/yscale; y++ {
			value := freq.Value([]int{0, 0, x, y})
			mat2.SetValue(value, []int{0, 0, x, y})
		}
	}
	inverse := fft.IFFTN(mat2)
	out := image.NewGray(image.Rect(0, 0, dx/xscale, dy/yscale))
	min, max := make([]float64, xscale*yscale), make([]float64, xscale*yscale)
	for i := range min {
		min[i] = 255
	}
	for x := 0; x < dx/xscale; x++ {
		for y := 0; y < dy/yscale; y++ {
			v := inverse.Value([]int{0, 0, x, y})
			value := 255 * cmplx.Abs(v)
			index := 0
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					vv := inverse.Value([]int{a, b, x, y})
					vvalue := 255 * cmplx.Abs(vv)
					if value != vvalue {
						fmt.Println("not the same")
					}
					if vvalue < min[index] {
						min[index] = vvalue
					}
					if vvalue > max[index] {
						max[index] = vvalue
					}
					index++
				}
			}
			out.SetGray(x, y, color.Gray{Y: byte(value)})
		}
	}
	fmt.Println(min)
	fmt.Println(max)
	return out
}

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
	out := Resize(gray, 8, 8)
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
