package packer

import (
	"image"
	"image/color"
	"testing"
)

func TestPackValidateConfig(t *testing.T) {
	t.Parallel()

	img := solid(8, 8)
	images := []ImageInfo{{Name: "a", Width: 8, Height: 8, Image: img}}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "ok",
			cfg: Config{
				MinSize: 16, MaxSize: 64, Gap: 1,
				Rule: BestShortSideFit,
			},
		},
		{
			name: "invalid-min-max",
			cfg: Config{
				MinSize: 64, MaxSize: 16, Gap: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid-gap",
			cfg: Config{
				MinSize: 16, MaxSize: 64, Gap: -1,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Pack(images, tc.cfg)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPackEmptyImagesReturnsMinSizeAtlas(t *testing.T) {
	t.Parallel()

	cfg := Config{MinSize: 32, MaxSize: 64, Gap: 0}
	res, err := Pack(nil, cfg)
	if err != nil {
		t.Fatalf("Pack(nil): %v", err)
	}
	if res.Width != 32 || res.Height != 32 {
		t.Fatalf("empty pack size = %dx%d, want 32x32", res.Width, res.Height)
	}
}

func TestPackPlacementsWithinBoundsNoOverlap(t *testing.T) {
	t.Parallel()

	images := []ImageInfo{
		{Name: "a", Width: 10, Height: 12, Image: solid(10, 12)},
		{Name: "b", Width: 8, Height: 8, Image: solid(8, 8)},
		{Name: "c", Width: 5, Height: 14, Image: solid(5, 14)},
	}
	cfg := Config{
		MinSize: 32, MaxSize: 64, Gap: 1,
		Rule: BottomLeft,
	}

	res, err := Pack(images, cfg)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if len(res.Placements) != len(images) {
		t.Fatalf("placements=%d, want %d", len(res.Placements), len(images))
	}

	for i := range res.Placements {
		p := res.Placements[i]
		if p.X < 0 || p.Y < 0 || p.X+p.Width > res.Width || p.Y+p.Height > res.Height {
			t.Fatalf("placement %q out of bounds: %+v atlas=%dx%d", p.Name, p, res.Width, res.Height)
		}
	}

	for i := 0; i < len(res.Placements); i++ {
		for j := i + 1; j < len(res.Placements); j++ {
			a := res.Placements[i]
			b := res.Placements[j]
			if overlaps(a.X, a.Y, a.Width, a.Height, b.X, b.Y, b.Width, b.Height) {
				t.Fatalf("placements overlap: %q and %q", a.Name, b.Name)
			}
		}
	}
}

func TestPackExceedsMaxSize(t *testing.T) {
	t.Parallel()

	images := []ImageInfo{
		{Name: "huge", Width: 512, Height: 512, Image: solid(512, 512)},
	}
	cfg := Config{
		MinSize: 64,
		MaxSize: 256,
		Gap:     0,
		Rule:    BestShortSideFit,
	}

	_, err := Pack(images, cfg)
	if err == nil {
		t.Fatal("expected error when required atlas exceeds MaxSize")
	}
}

func solid(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	return img
}

func overlaps(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
