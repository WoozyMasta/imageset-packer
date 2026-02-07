package packer

import "testing"

func TestSortImagesForPacking(t *testing.T) {
	t.Parallel()

	images := []ImageInfo{
		{Name: "b", Width: 8, Height: 8},  // max=8
		{Name: "c", Width: 10, Height: 1}, // max=10 area=10
		{Name: "e", Width: 9, Height: 2},  // max=9 area=18
		{Name: "a", Width: 10, Height: 2}, // max=10 area=20
	}

	sortImagesForPacking(images, Config{Gap: 0})

	got := []string{images[0].Name, images[1].Name, images[2].Name, images[3].Name}
	want := []string{"a", "c", "e", "b"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted order = %v, want %v", got, want)
		}
	}
}

func TestFindOptimalSizeForceSquare(t *testing.T) {
	t.Parallel()

	images := []ImageInfo{
		{Name: "one", Width: 40, Height: 20},
	}
	cfg := Config{
		MinSize: 16, MaxSize: 128, Gap: 0,
		ForceSquare: true,
		Rule:        BestShortSideFit,
	}

	w, h := findOptimalSize(images, cfg)
	if w != h {
		t.Fatalf("force square size = %dx%d, want square", w, h)
	}
	if w < 40 {
		t.Fatalf("force square width = %d, want >= 40", w)
	}
}

func TestFindOptimalSizePreferHeightTieBreak(t *testing.T) {
	t.Parallel()

	images := []ImageInfo{
		{Name: "a", Width: 30, Height: 20},
		{Name: "b", Width: 30, Height: 20},
	}
	base := Config{
		MinSize: 16, MaxSize: 64, Gap: 0,
		AspectPenalty: 0.25,
		Rule:          BestShortSideFit,
	}

	w1, h1 := findOptimalSize(images, base)
	if !(w1 == 64 && h1 == 32) {
		t.Fatalf("prefer width default size = %dx%d, want 64x32", w1, h1)
	}

	base.PreferHeight = true
	w2, h2 := findOptimalSize(images, base)
	if !(w2 == 32 && h2 == 64) {
		t.Fatalf("prefer height size = %dx%d, want 32x64", w2, h2)
	}
}
