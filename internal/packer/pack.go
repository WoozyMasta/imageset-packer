package packer

import (
	"fmt"
	"image"
	"image/draw"
)

// Pack packs the input images into a single atlas using MaxRects.
func Pack(images []ImageInfo, cfg Config) (*Result, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	if len(images) == 0 {
		size := cfg.MinSize
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		return &Result{Width: size, Height: size, Image: img}, nil
	}

	imgs := make([]ImageInfo, len(images))
	copy(imgs, images)
	sortImagesForPacking(imgs, cfg)

	w, h := findOptimalSize(imgs, cfg)
	if w > cfg.MaxSize || h > cfg.MaxSize {
		return nil, fmt.Errorf("required texture size (%dx%d) exceeds MaxSize=%d", w, h, cfg.MaxSize)
	}

	bin := newMaxRects(w, h, cfg.AllowRotate)
	atlas := image.NewRGBA(image.Rect(0, 0, w, h))

	placements := make([]Placement, 0, len(imgs))

	rule := cfg.Rule
	if rule < BestShortSideFit || rule > ContactPoint {
		rule = BestShortSideFit
	}

	for _, im := range imgs {
		pw := im.Width + 2*cfg.Gap
		ph := im.Height + 2*cfg.Gap

		rect, ok := bin.Insert(pw, ph, rule)
		if !ok {
			return nil, fmt.Errorf("failed to place %q into %dx%d", im.Name, w, h)
		}

		x := rect.X + cfg.Gap
		y := rect.Y + cfg.Gap

		placements = append(placements, Placement{
			Name:    im.Name,
			X:       x,
			Y:       y,
			Width:   im.Width,
			Height:  im.Height,
			Rotated: rect.Rotated,
		})

		// Rotation support: if rotated, draw into a temp buffer rotated.
		if rect.Rotated {
			rot := rotate90RGBA(im.Image)
			draw.Draw(atlas, image.Rect(x, y, x+im.Height, y+im.Width), rot, image.Point{}, draw.Src)
		} else {
			draw.Draw(atlas, image.Rect(x, y, x+im.Width, y+im.Height), im.Image, image.Point{}, draw.Src)
		}
	}

	return &Result{
		Width:      w,
		Height:     h,
		Placements: placements,
		Image:      atlas,
	}, nil
}

// validateConfig validates the configuration.
func validateConfig(cfg Config) error {
	if cfg.MinSize <= 0 || cfg.MaxSize <= 0 || cfg.MinSize > cfg.MaxSize {
		return fmt.Errorf("invalid config: MinSize=%d MaxSize=%d", cfg.MinSize, cfg.MaxSize)
	}

	if cfg.Gap < 0 {
		return fmt.Errorf("invalid config: Gap=%d", cfg.Gap)
	}

	return nil
}
