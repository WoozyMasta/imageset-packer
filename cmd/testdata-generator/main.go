package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/jessevdk/go-flags"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type Options struct {
	Args struct {
		OutputDir string `positional-arg-name:"output" description:"Output directory for generated PNG files" required:"yes"`
	} `positional-args:"yes" required:"yes"`

	MinSize      int  `short:"m" long:"min-size" description:"Minimum image size" default:"16"`
	MaxSize      int  `short:"M" long:"max-size" description:"Maximum image size" default:"256"`
	Count        int  `short:"c" long:"count" description:"Number of images to generate" default:"10"`
	MaxRatio     int  `short:"r" long:"max-ratio" description:"Maximum side ratio (1=squares only, 4=one side can be 4x larger)" default:"1"`
	AllowNonPow2 bool `short:"n" long:"allow-non-pow2" description:"Allow non-power-of-2 sizes"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "testdata-generator"
	parser.Usage = "[OPTIONS] <output>"

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := run(&opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(opts *Options) error {
	// Validate arguments.
	if opts.MinSize <= 0 || opts.MaxSize <= 0 {
		return fmt.Errorf("min-size and max-size must be positive")
	}
	if opts.MinSize > opts.MaxSize {
		return fmt.Errorf("min-size must be <= max-size")
	}
	if opts.Count <= 0 {
		return fmt.Errorf("count must be positive")
	}
	if opts.MaxRatio < 1 {
		return fmt.Errorf("max-ratio must be >= 1")
	}

	// Create output directory.
	if err := os.MkdirAll(opts.Args.OutputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	//nolint:gosec // Non-crypto randomness is fine for test data.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate images.
	for i := 0; i < opts.Count; i++ {
		width, height := generateSize(rng, opts)
		if err := generateImage(opts.Args.OutputDir, i, width, height, rng); err != nil {
			return fmt.Errorf("failed to generate image %d: %w", i, err)
		}
	}

	fmt.Printf("Successfully generated %d images in %s\n", opts.Count, opts.Args.OutputDir)
	return nil
}

// generateSize produces image dimensions based on options.
func generateSize(rng *rand.Rand, opts *Options) (width, height int) {
	// Pick a base size.
	size := opts.MinSize + rng.Intn(opts.MaxSize-opts.MinSize+1)

	// For power-of-two mode, round the base size.
	if !opts.AllowNonPow2 {
		size = nextPowerOfTwo(size)
		// Clamp to MaxSize.
		if size > opts.MaxSize {
			size = prevPowerOfTwo(opts.MaxSize)
		}
	}

	// If max-ratio = 1, we only generate squares.
	if opts.MaxRatio == 1 {
		return size, size
	}

	// If non-power-of-two sizes are allowed, pick more varied dimensions.
	if opts.AllowNonPow2 {
		for i := 0; i < 24; i++ {
			width = opts.MinSize + rng.Intn(opts.MaxSize-opts.MinSize+1)
			height = opts.MinSize + rng.Intn(opts.MaxSize-opts.MinSize+1)
			if width == 0 || height == 0 {
				continue
			}
			ratio := float64(max(width, height)) / float64(min(width, height))
			if ratio <= float64(opts.MaxRatio) {
				return width, height
			}
		}

		// Fallback: clamp to max ratio while keeping within bounds.
		if width == 0 || height == 0 {
			width = size
			height = size
		}
		if width >= height {
			width = min(opts.MaxSize, max(opts.MinSize, height*opts.MaxRatio))
		} else {
			height = min(opts.MaxSize, max(opts.MinSize, width*opts.MaxRatio))
		}
		return width, height
	}

	// Pick an aspect ratio in [1..MaxRatio].
	ratio := 1 + rng.Intn(opts.MaxRatio)

	// Decide which side is larger.
	if rng.Intn(2) == 0 {
		// width >= height
		width = size * ratio
		height = size
		// Clamp to MaxSize.
		if width > opts.MaxSize {
			width = opts.MaxSize
			if !opts.AllowNonPow2 {
				width = prevPowerOfTwo(opts.MaxSize)
			}
		}
	} else {
		// height >= width
		width = size
		height = size * ratio
		// Clamp to MaxSize.
		if height > opts.MaxSize {
			height = opts.MaxSize
			if !opts.AllowNonPow2 {
				height = prevPowerOfTwo(opts.MaxSize)
			}
		}
	}

	// For power-of-two mode, round both sides.
	if !opts.AllowNonPow2 {
		width = nextPowerOfTwo(width)
		height = nextPowerOfTwo(height)
		// Clamp again.
		if width > opts.MaxSize {
			width = prevPowerOfTwo(opts.MaxSize)
		}
		if height > opts.MaxSize {
			height = prevPowerOfTwo(opts.MaxSize)
		}
	}

	return width, height
}

// generateImage creates a PNG image with simple visual markers.
func generateImage(outputDir string, index, width, height int, rng *rand.Rand) error {
	// Create image.
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Random background color.
	bgColor := color.RGBA{
		R: randByte(rng),
		G: randByte(rng),
		B: randByte(rng),
		A: 255,
	}

	// Fill background.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Add a simple pattern for visual distinction.
	patternColor := color.RGBA{
		R: randByte(rng),
		G: randByte(rng),
		B: randByte(rng),
		A: 255,
	}

	// Draw a border.
	for y := 0; y < height; y++ {
		img.Set(0, y, patternColor)
		img.Set(width-1, y, patternColor)
	}
	for x := 0; x < width; x++ {
		img.Set(x, 0, patternColor)
		img.Set(x, height-1, patternColor)
	}

	// Draw a diagonal.
	drawDiagonal(img, patternColor)

	// Draw index label in the center.
	labelColor := color.RGBA{R: 0, G: 0, B: 0, A: 128}
	labelSize := float64(min(width, height)) * 0.5
	drawCenteredLabel(img, fmt.Sprintf("%d", index+1), labelSize, labelColor)

	// Save the file.
	filename := filepath.Join(outputDir, fmt.Sprintf("test_%03d_%dx%d.png", index, width, height))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

func drawDiagonal(img *image.RGBA, c color.RGBA) {
	b := img.Bounds()
	x0, y0 := b.Min.X, b.Min.Y
	x1, y1 := b.Max.X-1, b.Max.Y-1

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func drawCenteredLabel(img *image.RGBA, label string, size float64, c color.RGBA) {
	if size < 6 {
		return
	}
	tt, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return
	}
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return
	}
	defer func() { _ = face.Close() }()

	bounds, _ := font.BoundString(face, label)
	textW := (bounds.Max.X - bounds.Min.X).Ceil()
	textH := (bounds.Max.Y - bounds.Min.Y).Ceil()

	b := img.Bounds()
	x := b.Min.X + (b.Dx()-textW)/2 - bounds.Min.X.Ceil()
	y := b.Min.Y + (b.Dy()-textH)/2 - bounds.Min.Y.Ceil()

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(label)
}

// nextPowerOfTwo returns the next power of two >= n.
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	if n&(n-1) == 0 {
		return n // already a power of two
	}
	power := 1
	for power < n {
		power <<= 1
	}
	return power
}

// prevPowerOfTwo returns the previous power of two <= n.
func prevPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	power := 1
	for power*2 <= n {
		power <<= 1
	}
	return power
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func randByte(rng *rand.Rand) uint8 {
	//nolint:gosec // Intn(256) is always within uint8.
	return uint8(rng.Intn(256))
}
