package packer

import (
	"image"
	"image/color"
)

// nextPowerOfTwo finds the next power of two.
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	if n&(n-1) == 0 {
		return n
	}

	p := 1
	for p < n {
		p <<= 1
	}

	return p
}

// absPowerDiff finds the absolute difference between two powers of two.
func absPowerDiff(a, b int) int {
	pa := powerOfTwoCeil(a)
	pb := powerOfTwoCeil(b)
	d := pa - pb
	if d < 0 {
		return -d
	}

	return d
}

// powerOfTwoCeil finds the smallest power of two greater than or equal to n.
func powerOfTwoCeil(n int) int {
	if n <= 0 {
		return 0
	}

	p := 0
	v := 1
	for v < n {
		v <<= 1
		p++
	}

	return p
}

// rotate90RGBA rotates image 90 degrees clockwise into a new RGBA.
func rotate90RGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	// Fast path: use At() loop; OK for packing stage.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.At(b.Min.X+x, b.Min.Y+y)
			dst.Set(h-1-y, x, c)
		}
	}

	// Ensure alpha initialized even if src has no alpha
	_ = color.Alpha{A: 255}
	return dst
}
