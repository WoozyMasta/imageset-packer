package packer

import "sort"

// sortImagesForPacking sorts the images for packing.
func sortImagesForPacking(images []ImageInfo, cfg Config) {
	sort.Slice(images, func(i, j int) bool {
		wi := images[i].Width + 2*cfg.Gap
		hi := images[i].Height + 2*cfg.Gap
		wj := images[j].Width + 2*cfg.Gap
		hj := images[j].Height + 2*cfg.Gap

		mi := wi
		if hi > mi {
			mi = hi
		}
		mj := wj
		if hj > mj {
			mj = hj
		}
		if mi != mj {
			return mi > mj
		}

		ai := wi * hi
		aj := wj * hj
		if ai != aj {
			return ai > aj
		}

		if hi != hj {
			return hi > hj
		}

		return wi > wj
	})
}
