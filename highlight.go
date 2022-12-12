package readline

import (
	"sort"
)

// region is a part of the input line for which to apply some highlighting.
type region struct {
	regionType  string
	bpos        int
	epos        int
	highlightFg string
	highlightBg string
}

func (rl *Instance) addRegion(rtype string, bpos, epos int, fg, bg string) {
	hl := region{
		regionType:  rtype,
		bpos:        bpos,
		epos:        epos,
		highlightFg: fg,
		highlightBg: bg,
	}

	rl.regions = append(rl.regions, hl)
}

func (rl *Instance) getHighlights(line []rune) map[int][]rune {
	hl := make(map[int][]rune)

	// first sort out the regions by bpos
	sorted := make([]*region, 0)
	bpos := make([]int, 0)

	for _, reg := range rl.regions {
		bpos = append(bpos, reg.bpos)
	}
	sort.Ints(bpos)

	for _, pos := range bpos {
		for _, reg := range rl.regions {
			if reg.bpos == pos {
				sorted = append(sorted, &reg)
				break
			}
		}
	}

	// regions that started highlighting, but not done yet.
	pending := make([]*region, 0)

	// Build the string.
	for lineIndex := range line {
		var posHl []rune
		var newHl *region

		// First check if we have a new highlighter to apply
		for _, reg := range sorted {
			if reg.bpos == lineIndex {
				newHl = reg
				pending = append(pending, reg)
			}
		}

		// If we have a region that is done highlighting, reset
		doneReset := false
		for i, reg := range pending {
			if reg.epos == lineIndex {
				pending = append(pending[:i], pending[i+1:]...)
				if !doneReset {
					posHl = append(posHl, []rune(RESET)...)
				}
			}
		}

		// If we have a new higlighting, apply it.
		if newHl != nil {
			posHl = append(posHl, []rune(newHl.highlightBg)...)
			posHl = append(posHl, []rune(newHl.highlightFg)...)
		} else if len(pending) > 0 && doneReset {
			backHl := pending[len(pending)-1]
			posHl = append(posHl, []rune(backHl.highlightBg)...)
			posHl = append(posHl, []rune(backHl.highlightFg)...)
		}

		// Add to the line.
		if len(posHl) > 0 {
			hl[lineIndex] = posHl
		}
	}

	return hl
}

func (rl *Instance) resetRegions() {
	rl.regions = make([]region, 0)
}

// highlightLine adds highlighting of the region if we are in a visual mode.
func (rl *Instance) highlightLine(line []rune) string {
	// Add an highlight region if we have visual or active region.
	rl.addVisualHighlight(line)

	var highlighted string

	// Get the highlighting.
	hl := rl.getHighlights(line)

	// Try to apply highlighting before each rune.
	for i, r := range line {
		if highlight, found := hl[i]; found {
			highlighted += string(highlight)
		}

		highlighted += string(r)
	}

	highlighted += RESET

	return highlighted
}

func (rl *Instance) addVisualHighlight(line []rune) {
	// Remove old visual region
	for i, reg := range rl.regions {
		if reg.regionType == "visual" {
			if len(rl.regions) > i {
				rl.regions = append(rl.regions[:i], rl.regions[i+1:]...)
			}
		}
	}

	if rl.local != visual {
		return
	}

	// And create the new one.
	var start, end int
	if rl.mark < rl.pos {
		start = rl.mark
		end = rl.pos
	} else {
		start = rl.pos
		end = rl.mark
	}
	end++

	// Adjust if we are in visual line mode
	if rl.local == visual && rl.visualLine {
		end = len(line) - 1
	}

	rl.addRegion("visual", start, end, "", seqBgBlueDark)
}
