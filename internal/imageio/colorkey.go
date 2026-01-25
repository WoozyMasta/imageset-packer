package imageio

import (
	"image"
	"image/draw"
)

// ApplyColorKey makes all pixels matching the RGB key fully transparent.
func ApplyColorKey(img image.Image, key RGB) image.Image {
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, img, b.Min, draw.Src)

	p := rgba.Pix
	for i := 0; i+3 < len(p); i += 4 {
		if p[i] == key.R && p[i+1] == key.G && p[i+2] == key.B {
			p[i+3] = 0
		}
	}

	return rgba
}
