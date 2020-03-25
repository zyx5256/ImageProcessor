// Package pngimg allows for loading png images and applying
// image flitering effects on them
package imgProcessor

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

var kernelS = [3][3]float64 {
		[3]float64{0, -1, 0},
		[3]float64{-1, 5, -1},
		[3]float64{0, -1, 0},
	}

var kernelE = [3][3]float64 {
		[3]float64{-1, -1, -1},
		[3]float64{-1, 8, -1},
		[3]float64{-1, -1, -1},
	}

var kernelB = [3][3]float64 {
		[3]float64{float64(1)/9, float64(1)/9, float64(1)/9},
		[3]float64{float64(1)/9, float64(1)/9, float64(1)/9},
		[3]float64{float64(1)/9, float64(1)/9, float64(1)/9},
	}

func (img *PNGImage) process(knl [3][3]float64, x int, y int) {

	resR := float64(0)
	resG := float64(0)
	resB := float64(0)

	_, _, _, a := img.in.At(x + 1, y + 1).RGBA()

	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {

			r, g, b, _ := img.in.At(x + 1 + i, y + 1 + j).RGBA()

			resR += float64(r) * knl[1 - i][1 - j]
			resG += float64(g) * knl[1 - i][1 - j]
			resB += float64(b) * knl[1 - i][1 - j]
		}
	}
	img.out.Set(x, y, color.RGBA64{clamp(resR), clamp(resG), clamp(resB), uint16(a)})
}

func (img *PNGImage) padding() {

	bounds := img.in.Bounds()
	bounds.Max.X += 2
	bounds.Max.Y += 2
	padded := image.NewRGBA64(bounds)

	for y := bounds.Min.Y + 1; y < bounds.Max.Y - 1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X - 1; x++ {

			r, g, b, a := img.in.At(x - 1, y - 1).RGBA()
			padded.Set(x, y, color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)})
		}
	}
	img.in = padded
}

// The PNGImage represents a structure for working with PNG images.
type PNGImage struct {
	in  image.Image
	out *image.RGBA64
}

//
// Public functions
//
func (img *PNGImage) Bounds() image.Rectangle {
	return img.in.Bounds()
}

// Save saves the image to the given file
func (img *PNGImage) Save(filePath string) error {

	outWriter, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outWriter.Close()

	err = png.Encode(outWriter, img.out)
	if err != nil {
		return err
	}
	return nil
}

//clamp will clamp the comp parameter to zero if it is less than zero or to 65535 if the comp parameter
// is greater than 65535.
func clamp(comp float64) uint16 {
	return uint16(math.Min(65535, math.Max(0, comp)))
}

// G applies a grayscale filtering effect to the image
func (img *PNGImage) G(bounds image.Rectangle) *PNGImage {

	// Bounds returns defines the dimensions of the image.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			//Returns the pixel (i.e., RGBA) value at a (x,y) position
			r, g, b, a := img.in.At(x, y).RGBA()

			greyC := clamp(float64(r+g+b) / 3)
			
			img.out.Set(x, y, color.RGBA64{greyC, greyC, greyC, uint16(a)})
		}
	}
	return img
}

// KernelEffect applies a kernel effect to the image
func (img *PNGImage) KernelEffect(bounds image.Rectangle, knl [3][3]float64) *PNGImage {

	img.padding()
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			img.process(knl, x, y)
		}
	}
	return img
}

// Apply applies specified effect to the image
func (img *PNGImage) Apply(effect string, bounds image.Rectangle) {

	switch effect {
	case "G":
		img.G(bounds)
	case "S":
		img.KernelEffect(bounds, kernelS)
	case "B":
		img.KernelEffect(bounds, kernelB)
	case "E":
		img.KernelEffect(bounds, kernelE)
	}
}
