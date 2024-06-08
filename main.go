// Copyright 2024 The KAN Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math/cmplx"
	"math/rand"
	"os"

	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
	"github.com/pointlander/gradient/sc128"
	"github.com/pointlander/matrix"
)

// Useses the fft to resize an image
func Resize(gray *image.Gray, xscale, yscale, padx, pady int) *image.Gray {
	bounds := gray.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	fmt.Println(dx/xscale, dy/yscale)
	mat := dsputils.MakeMatrix(make([]complex128, xscale*yscale*((dx/xscale)*(dy/yscale))),
		[]int{xscale, yscale, dx / xscale, dy / yscale})
	for x := 0; x < dx; x += xscale {
		for y := 0; y < dy; y += yscale {
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					g := float64(gray.GrayAt(x+a, y+b).Y)
					mat.SetValue(complex(g/255, 0), []int{a, b, x / xscale, y / yscale})
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
	out := image.NewGray(image.Rect(0, 0, dx/xscale+padx, dy/yscale+pady))
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

// Transform uses fft and self attention to transform an image
func Transform(gray *image.Gray, xscale, yscale, padx, pady int) *image.Gray {
	bounds := gray.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	fmt.Println(dx/xscale, dy/yscale)
	mat := dsputils.MakeMatrix(make([]complex128, xscale*yscale*((dx/xscale)*(dy/yscale))),
		[]int{xscale, yscale, dx / xscale, dy / yscale})
	for x := 0; x < dx; x += xscale {
		for y := 0; y < dy; y += yscale {
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					g := float64(gray.GrayAt(x+a, y+b).Y)
					mat.SetValue(complex(g/255, 0), []int{a, b, x / xscale, y / yscale})
				}
			}
		}
	}
	freq := fft.FFTN(mat)
	mat3 := matrix.NewComplexMatrix(xscale*yscale, ((dx / xscale) * (dy / yscale)))
	for x := 0; x < dx/xscale; x++ {
		for y := 0; y < dy/yscale; y++ {
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					value := freq.Value([]int{a, b, x, y})
					mat3.Data = append(mat3.Data, value)
				}
			}
		}
	}
	mat4 := matrix.ComplexSelfAttention(mat3, mat3, mat3)
	mat2 := dsputils.MakeMatrix(make([]complex128, xscale*yscale*((dx/xscale)*(dy/yscale))),
		[]int{xscale, yscale, dx / xscale, dy / yscale})
	for x := 0; x < dx/xscale; x++ {
		for y := 0; y < dy/yscale; y++ {
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					value := mat4.Data[(y*(dy/yscale)+x)*yscale*xscale+b*xscale+a]
					mat2.SetValue(value, []int{a, b, x, y})
				}
			}
		}
	}
	inverse := fft.IFFTN(mat2)
	out := image.NewGray(image.Rect(0, 0, dx+padx, dy+pady))
	min, max := make([]float64, xscale*yscale), make([]float64, xscale*yscale)
	for i := range min {
		min[i] = 255
	}
	for x := 0; x < dx/xscale; x++ {
		for y := 0; y < dy/yscale; y++ {
			index := 0
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					vv := inverse.Value([]int{a, b, x, y})
					vvalue := 255 * cmplx.Abs(vv)
					if vvalue < min[index] {
						min[index] = vvalue
					}
					if vvalue > max[index] {
						max[index] = vvalue
					}
					index++
				}
			}
		}
	}
	fmt.Println(min)
	fmt.Println(max)

	for x := 0; x < dx; x += xscale {
		for y := 0; y < dy; y += yscale {
			index := 0
			for a := 0; a < xscale; a++ {
				for b := 0; b < yscale; b++ {
					v := inverse.Value([]int{a, b, x / xscale, y / yscale})
					value := 255 * cmplx.Abs(v)
					out.SetGray(x+a, y+b, color.Gray{Y: byte(255 * value / max[index])})
					index++
				}
			}
		}
	}
	return out
}

// FFTSA is fft with self attention
func FFTSA() {
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
	out := Resize(gray, 8, 8, 2, 0)
	output, err := os.Create("gray.jpg")
	if err != nil {
		panic(err)
	}
	defer output.Close()
	err = png.Encode(output, out)
	if err != nil {
		panic(err)
	}
	out2 := Transform(out, 16, 16, 0, 0)
	output2, err := os.Create("transformed.jpg")
	if err != nil {
		panic(err)
	}
	defer output.Close()
	err = png.Encode(output2, out2)
	if err != nil {
		panic(err)
	}
}

var (
	// FlatFFTSA is fft with self attention mode
	FlagFFTSA = flag.Bool("fftsa", false, "fftsa mode")
)

func main() {
	flag.Parse()

	if *FlagFFTSA {
		FFTSA()
		return
	}

	xor := [][]bool{
		{false, false, false},
		{true, false, true},
		{false, true, true},
		{true, true, false},
	}

	rng := rand.New(rand.NewSource(1))
	input := make([]sc128.V, 2)
	layer1 := make([]sc128.V, 10)
	for i := range layer1 {
		layer1[i].X = complex(rng.Float64()*2-1, rng.Float64()*2-1)
	}
	layer2 := make([]sc128.V, 5)
	for i := range layer2 {
		layer2[i].X = complex(rng.Float64()*2-1, rng.Float64()*2-1)
	}
	output := sc128.V{}
	x0 := sc128.Mul(input[0].Meta(), layer1[0].Meta())
	x1 := sc128.Mul(input[0].Meta(), layer1[1].Meta())
	x2 := sc128.Mul(input[0].Meta(), layer1[2].Meta())
	x3 := sc128.Mul(input[0].Meta(), layer1[3].Meta())
	x4 := sc128.Mul(input[0].Meta(), layer1[4].Meta())
	x5 := sc128.Mul(input[1].Meta(), layer1[5].Meta())
	x6 := sc128.Mul(input[1].Meta(), layer1[6].Meta())
	x7 := sc128.Mul(input[1].Meta(), layer1[7].Meta())
	x8 := sc128.Mul(input[1].Meta(), layer1[8].Meta())
	x9 := sc128.Mul(input[1].Meta(), layer1[9].Meta())
	y0 := sc128.Add(x0, x5)
	y1 := sc128.Add(x1, x6)
	y2 := sc128.Add(x2, x7)
	y3 := sc128.Add(x3, x8)
	y4 := sc128.Add(x4, x9)
	z0 := sc128.Mul(y0, layer2[0].Meta())
	z1 := sc128.Mul(y1, layer2[1].Meta())
	z2 := sc128.Mul(y2, layer2[2].Meta())
	z3 := sc128.Mul(y3, layer2[3].Meta())
	z4 := sc128.Mul(y4, layer2[4].Meta())
	grand := sc128.Add(sc128.Add(sc128.Add(z0, z1), sc128.Add(z2, z3)), z4)
	loss := sc128.Sub(output.Meta(), grand)
	loss = sc128.Mul(loss, loss)
	for i := 0; i < 1024; i++ {
		for i := range layer1 {
			layer1[i].D = 0
		}
		for i := range layer2 {
			layer2[i].D = 0
		}
		t := xor[rand.Intn(len(xor))]
		if t[0] {
			input[0].X = 1 + 1i
		} else {
			input[0].X = -1 - 1i
		}
		if t[1] {
			input[1].X = 1 + 1i
		} else {
			input[1].X = -1 - 1i
		}
		if t[2] {
			output.X = 1 + 1i
		} else {
			output.X = -1 - 1i
		}
		cost := sc128.Gradient(loss)
		for i := range layer1 {
			layer1[i].X -= (.001 + .001i) * layer1[i].D
		}
		for i := range layer2 {
			layer2[i].X -= (.001 + .001i) * layer2[i].D
		}
		fmt.Println(cmplx.Abs(cost.X))
	}
	for _, t := range xor {
		if t[0] {
			input[0].X = 1 + 1i
		} else {
			input[0].X = -1 - 1i
		}
		if t[1] {
			input[1].X = 1 + 1i
		} else {
			input[1].X = -1 - 1i
		}
		if t[2] {
			output.X = 1 + 1i
		} else {
			output.X = -1 - 1i
		}
		grand(func(a *sc128.V) bool {
			fmt.Println(output.X, a.X)
			return true
		})
	}
}
