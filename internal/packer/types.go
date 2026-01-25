package packer

import "image"

// Config controls atlas packing behavior.
type Config struct {
	MinSize       int     // minimum texture size (power of 2)
	MaxSize       int     // maximum texture size (power of 2)
	Gap           int     // gap between images
	AspectPenalty float64 // 0..1, 0 means no penalty, 1 means maximum penalty
	Rule          Rule    // packing rule: BestShortSideFit, BestLongSideFit, BestAreaFit, BottomLeft, ContactPoint
	PreferHeight  bool    // prefer height over width for aspect ratio
	ForceSquare   bool    // force square texture
	AllowRotate   bool    // optional, improves packing a lot for tall/wide sprites
}

// Rule is the packing heuristic used to place rectangles.
type Rule int

const (
	BestShortSideFit Rule = iota // BestShortSideFit is the best short side fit.
	BestLongSideFit              // BestLongSideFit is the best long side fit.
	BestAreaFit                  // BestAreaFit is the best area fit.
	BottomLeft                   // BottomLeft is the bottom left fit.
	ContactPoint                 // ContactPoint is the contact point fit.
)

// ImageInfo describes a source image to pack.
type ImageInfo struct {
	Image  image.Image // Image to pack.
	Name   string      // Name of the image.
	Width  int         // Width of the image.
	Height int         // Height of the image.
}

// Placement describes where an image ended up in the atlas.
type Placement struct {
	Name    string // Name of the image.
	X       int    // X position of the image.
	Y       int    // Y position of the image.
	Width   int    // Width of the image.
	Height  int    // Height of the image.
	Rotated bool   // Whether the image was rotated.
}

// Result holds the packed atlas and placements.
type Result struct {
	Image      image.Image // Image of the packed atlas.
	Placements []Placement // Placements of the images in the atlas.
	Width      int         // Width of the packed atlas.
	Height     int         // Height of the packed atlas.
}
