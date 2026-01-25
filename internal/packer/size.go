package packer

import "math"

// findOptimalSize finds the optimal size for the atlas.
func findOptimalSize(images []ImageInfo, cfg Config) (width, height int) {
	minW, minH := 0, 0
	for _, im := range images {
		w := im.Width + 2*cfg.Gap
		h := im.Height + 2*cfg.Gap
		if w > minW {
			minW = w
		}
		if h > minH {
			minH = h
		}
	}

	size := cfg.MinSize
	if minW > size {
		size = nextPowerOfTwo(minW)
	}
	if minH > size {
		size = nextPowerOfTwo(minH)
	}

	bestW, bestH := size, size
	bestScore := math.MaxFloat64

	rule := cfg.Rule
	if rule < BestShortSideFit || rule > ContactPoint {
		rule = BestShortSideFit
	}

	try := func(w, h int) {
		if w > cfg.MaxSize || h > cfg.MaxSize {
			return
		}

		if cfg.ForceSquare {
			if w != h {
				return
			}
		} else {
			if absPowerDiff(w, h) > 1 {
				return
			}
		}

		if !canFitMaxRects(images, w, h, cfg, rule) {
			return
		}

		aspect := float64(max(w, h)) / float64(min(w, h))
		score := float64(w*h) * (1.0 + cfg.AspectPenalty*(aspect-1.0))

		// Stable tie-break: default horizontal, PreferHeight => vertical
		if score < bestScore || (score == bestScore && ((!cfg.PreferHeight && w >= h) || (cfg.PreferHeight && h >= w))) {
			bestScore = score
			bestW, bestH = w, h
		}
	}

	if cfg.ForceSquare {
		for s := size; s <= cfg.MaxSize; s *= 2 {
			try(s, s)
		}
		return bestW, bestH
	}

	if cfg.PreferHeight {
		for w := size; w <= cfg.MaxSize; w *= 2 {
			for h := size; h <= cfg.MaxSize; h *= 2 {
				try(w, h)
			}
		}
	} else {
		for h := size; h <= cfg.MaxSize; h *= 2 {
			for w := size; w <= cfg.MaxSize; w *= 2 {
				try(w, h)
			}
		}
	}

	return bestW, bestH
}

// canFitMaxRects checks if the images can fit into the atlas.
func canFitMaxRects(images []ImageInfo, w, h int, cfg Config, rule Rule) bool {
	bin := newMaxRects(w, h, cfg.AllowRotate)
	for _, im := range images {
		pw := im.Width + 2*cfg.Gap
		ph := im.Height + 2*cfg.Gap

		_, ok := bin.Insert(pw, ph, rule)
		if !ok {
			return false
		}
	}

	return true
}
