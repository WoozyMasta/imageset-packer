package packer

// mrRect represents a rectangle.
type mrRect struct {
	X, Y, W, H int
	Rotated    bool
}

// maxRects represents a max rectangle packer.
type maxRects struct {
	used        []mrRect
	free        []mrRect
	w, h        int
	allowRotate bool
}

// newMaxRects creates a new max rectangle packer.
func newMaxRects(w, h int, allowRotate bool) *maxRects {
	m := &maxRects{
		w:           w,
		h:           h,
		allowRotate: allowRotate,
		used:        make([]mrRect, 0, 128),
		free:        make([]mrRect, 0, 128),
	}
	m.free = append(m.free, mrRect{X: 0, Y: 0, W: w, H: h})
	return m
}

// Insert inserts a rectangle into the max rectangle packer.
func (m *maxRects) Insert(w, h int, rule Rule) (mrRect, bool) {
	best := mrRect{}
	pri, sec := 1<<30, 1<<30
	found := false

	for i := 0; i < len(m.free); i++ {
		fr := m.free[i]

		// non-rotated
		if fr.W >= w && fr.H >= h {
			p1, s1 := m.score(rule, fr, w, h)
			if p1 < pri || (p1 == pri && s1 < sec) {
				pri, sec = p1, s1
				best = mrRect{X: fr.X, Y: fr.Y, W: w, H: h, Rotated: false}
				found = true
			}
		}

		// rotated
		if m.allowRotate && fr.W >= h && fr.H >= w {
			p2, s2 := m.score(rule, fr, h, w)
			if p2 < pri || (p2 == pri && s2 < sec) {
				pri, sec = p2, s2
				best = mrRect{X: fr.X, Y: fr.Y, W: h, H: w, Rotated: true}
				found = true
			}
		}
	}

	if !found {
		return mrRect{}, false
	}

	m.place(best)
	return best, true
}

// place places a rectangle into the max rectangle packer.
func (m *maxRects) place(used mrRect) {
	for i := 0; i < len(m.free); {
		if m.splitFree(i, used) {
			m.free = removeAt(m.free, i)
			continue
		}
		i++
	}

	m.pruneFree()
	m.used = append(m.used, used)
}

// score scores a rectangle into the max rectangle packer.
func (m *maxRects) score(rule Rule, fr mrRect, rw, rh int) (pri, sec int) {
	switch rule {
	case BestShortSideFit: // BestShortSideFit is the best short side fit.
		leftoverH := fr.W - rw
		if leftoverH < 0 {
			leftoverH = -leftoverH
		}
		leftoverV := fr.H - rh
		if leftoverV < 0 {
			leftoverV = -leftoverV
		}

		shortSide := leftoverH
		longSide := leftoverV
		if leftoverV < shortSide {
			shortSide = leftoverV
		}
		if leftoverH > longSide {
			longSide = leftoverH
		}

		return shortSide, longSide

	case BestLongSideFit: // BestLongSideFit is the best long side fit.
		leftoverH := fr.W - rw
		if leftoverH < 0 {
			leftoverH = -leftoverH
		}
		leftoverV := fr.H - rh
		if leftoverV < 0 {
			leftoverV = -leftoverV
		}

		shortSide := leftoverH
		longSide := leftoverV
		if leftoverV < shortSide {
			shortSide = leftoverV
		}
		if leftoverH > longSide {
			longSide = leftoverH
		}

		return longSide, shortSide

	case BestAreaFit: // BestAreaFit is the best area fit.
		areaFit := fr.W*fr.H - rw*rh
		leftoverH := fr.W - rw
		if leftoverH < 0 {
			leftoverH = -leftoverH
		}

		leftoverV := fr.H - rh
		if leftoverV < 0 {
			leftoverV = -leftoverV
		}

		shortSide := leftoverH
		if leftoverV < shortSide {
			shortSide = leftoverV
		}

		return areaFit, shortSide

	case BottomLeft: // BottomLeft is the bottom left fit.
		// primary: top side Y, secondary: X
		return fr.Y + rh, fr.X

	case ContactPoint: // ContactPoint is the contact point fit.
		// maximize contact => minimize negative
		return -m.contactScore(fr.X, fr.Y, rw, rh), 0

	default:
		return 1 << 30, 1 << 30
	}
}

// contactScore calculates the contact score.
func (m *maxRects) contactScore(x, y, w, h int) int {
	score := 0
	if x == 0 || x+w == m.w {
		score += h
	}
	if y == 0 || y+h == m.h {
		score += w
	}

	for i := 0; i < len(m.used); i++ {
		u := m.used[i]
		if u.X == x+w || u.X+u.W == x {
			score += commonInterval(u.Y, u.Y+u.H, y, y+h)
		}
		if u.Y == y+h || u.Y+u.H == y {
			score += commonInterval(u.X, u.X+u.W, x, x+w)
		}
	}

	return score
}

// commonInterval calculates the common interval.
func commonInterval(a0, a1, b0, b1 int) int {
	if a1 <= b0 || b1 <= a0 {
		return 0
	}

	end := a1
	if b1 < end {
		end = b1
	}
	start := a0
	if b0 > start {
		start = b0
	}

	return end - start
}

// splitFree splits the free rectangle into two.
func (m *maxRects) splitFree(freeIdx int, used mrRect) bool {
	fr := m.free[freeIdx]

	// SAT
	if used.X >= fr.X+fr.W || used.X+used.W <= fr.X || used.Y >= fr.Y+fr.H || used.Y+used.H <= fr.Y {
		return false
	}

	// horizontal splits
	if used.X < fr.X+fr.W && used.X+used.W > fr.X {
		// top
		if used.Y > fr.Y && used.Y < fr.Y+fr.H {
			m.free = append(m.free, mrRect{X: fr.X, Y: fr.Y, W: fr.W, H: used.Y - fr.Y})
		}
		// bottom
		if used.Y+used.H < fr.Y+fr.H {
			m.free = append(m.free, mrRect{X: fr.X, Y: used.Y + used.H, W: fr.W, H: fr.Y + fr.H - (used.Y + used.H)})
		}
	}

	// vertical splits
	if used.Y < fr.Y+fr.H && used.Y+used.H > fr.Y {
		// left
		if used.X > fr.X && used.X < fr.X+fr.W {
			m.free = append(m.free, mrRect{X: fr.X, Y: fr.Y, W: used.X - fr.X, H: fr.H})
		}
		// right
		if used.X+used.W < fr.X+fr.W {
			m.free = append(m.free, mrRect{X: used.X + used.W, Y: fr.Y, W: fr.X + fr.W - (used.X + used.W), H: fr.H})
		}
	}

	return true
}

// pruneFree prunes the free rectangles.
func (m *maxRects) pruneFree() {
	for i := 0; i < len(m.free); i++ {
		a := m.free[i]
		for j := i + 1; j < len(m.free); j++ {
			b := m.free[j]
			if containedIn(a, b) {
				m.free = removeAt(m.free, i)
				i--
				break
			}

			if containedIn(b, a) {
				m.free = removeAt(m.free, j)
				j--
			}
		}
	}
}

// containedIn checks if a rectangle is contained in another rectangle.
func containedIn(a, b mrRect) bool {
	return a.X >= b.X && a.Y >= b.Y && a.X+a.W <= b.X+b.W && a.Y+a.H <= b.Y+b.H
}

// removeAt removes an item at a given index.
func removeAt[T any](s []T, i int) []T {
	if i < 0 || i >= len(s) {
		return s
	}

	copy(s[i:], s[i+1:])
	return s[:len(s)-1]
}
