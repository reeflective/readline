package readline

import (
	"regexp"
	"sort"
)

// region is a part of the input line for which to apply some highlighting.
type region struct {
	regionType string
	bpos       int
	epos       int
	fg         string
	bg         string
}

func (rl *Instance) addRegion(rtype string, bpos, epos int, fg, bg string) {
	hl := region{
		regionType: rtype,
		bpos:       bpos,
		epos:       epos,
		fg:         fg,
		bg:         bg,
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

	// Find any highlighting already applied on the line,
	// and keep the indexes so that we can skip those.
	var colors [][]int

	colorMatch, err := regexp.Compile(`\x1b\[[0-9;]+m`)
	if err == nil {
		colors = colorMatch.FindAllStringIndex(string(line), -1)
	}

	// regions that started highlighting, but not done yet.
	pending := make([]*region, 0)
	lineIndex := -1
	skip := 0

	// Build the string.
	for rawIndex := range line {
		var posHl []rune
		var newHl *region

		// While in a color escape, keep reading runes.
		if skip > 0 {
			skip--
			continue
		}

		// If starting a color escape code, add offset and read.
		if len(colors) > 0 && colors[0][0] == rawIndex {
			skip += colors[0][1] - colors[0][0] - 1
			colors = colors[1:]
			continue
		}

		// Or we are reading a printed rune.
		lineIndex += 1

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
					posHl = append(posHl, []rune(seqReset)...)
				}
			}
		}

		// If we have a new higlighting, apply it.
		if newHl != nil {
			posHl = append(posHl, []rune(newHl.bg)...)
			posHl = append(posHl, []rune(newHl.fg)...)
		} else if len(pending) > 0 && doneReset {
			backHl := pending[len(pending)-1]
			posHl = append(posHl, []rune(backHl.bg)...)
			posHl = append(posHl, []rune(backHl.fg)...)
		}

		// Add to the line, with the raw index since
		// we must take into account embedded colors.
		if len(posHl) > 0 {
			hl[rawIndex] = posHl
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

	highlighted += seqReset

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

	// Either Vim visual, or emacs active region.
	if rl.main == emacs && !rl.activeRegion ||
		(rl.main == vicmd || rl.main == viins) && rl.local != visual {
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
