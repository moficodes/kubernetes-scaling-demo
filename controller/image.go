package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

func makeSquare(img image.Image) image.Image {
	// Get the dimensions of the image.
	width, height := img.Bounds().Size().X, img.Bounds().Size().Y

	// Determine the longer dimension.
	longerDim := width
	if height > width {
		longerDim = height
	}

	// Calculate the padding needed on each side.
	padding := longerDim/2 - width/2

	// Create a new image with the desired dimensions.
	squareImage := image.NewRGBA(image.Rect(0, 0, longerDim, longerDim))

	// Copy the original image into the new image, padding with black pixels as needed.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			squareImage.Set(x+padding, y+padding, img.At(x, y))
		}
	}

	// Return the new image.
	return squareImage
}

func ConvertToSquare(img image.Image) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// Determine the longer side
	maxSide := width
	if height > width {
		maxSide = height
	}

	// Create a new square image of size maxSide x maxSide
	square := image.NewRGBA(image.Rect(0, 0, maxSide, maxSide))
	black := color.RGBA{0, 0, 0, 255}

	// Fill the square image with black color
	draw.Draw(square, square.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)

	// Compute offset to draw the original image at the center
	offsetX := (maxSide - width) / 2
	offsetY := (maxSide - height) / 2

	// Draw the original image onto the square image at the calculated offset
	draw.Draw(square, image.Rect(offsetX, offsetY, width+offsetX, height+offsetY), img, img.Bounds().Min, draw.Over)

	return square
}

func gammaCorrect(value int, gamma float64) int {
	return int(math.Pow(float64(value)/255.0, gamma)*255.0 + 0.5)
}
