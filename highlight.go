package readline

import (
	"regexp"
	"sort"
)

// highlightLine adds highlighting of the region if we are in a visual mode.
func (rl *Instance) highlightLine(line []rune) string {
	// Add an highlight region if we have visual or active region.
	vhl := rl.addVisualHighlight()

	// Sort the highlighted sections
	sorted := rl.sortHighlights(vhl)

	// Get the highlighting.
	colors := rl.getHighlights(line, sorted)

	var highlighted string

	// And apply highlighting before each rune.
	for i, r := range line {
		if highlight, found := colors[i]; found {
			highlighted += string(highlight)
		}

		highlighted += string(r)
	}

	highlighted += seqReset

	return highlighted
}

func (rl *Instance) addVisualHighlight() *selection {
	bpos, epos, _ := rl.calcSelection()

	visual := rl.visualSelection()
	if visual == nil {
		return nil
	}

	// Make a copy so that we don't overwrite any nil ending position.
	vhl := &selection{
		bpos:       bpos,
		epos:       epos,
		active:     visual.active,
		regionType: visual.regionType,
		fg:         visual.fg,
		bg:         visual.bg,
	}

	return vhl
}

func (rl *Instance) sortHighlights(vhl *selection) []*selection {
	// first sort out the regions by bpos
	sorted := make([]*selection, 0)
	bpos := make([]int, 0)

	for _, reg := range rl.marks {
		bpos = append(bpos, reg.bpos)
	}
	sort.Ints(bpos)

	for _, pos := range bpos {
		for _, reg := range rl.marks {
			if reg.bpos == pos {
				if reg.regionType == "visual" {
					if vhl != nil && rl.local == visual {
						sorted = append(sorted, vhl)
					}
					break
				}
				sorted = append(sorted, reg)
				break
			}
		}
	}

	return sorted
}

func (rl *Instance) getHighlights(line []rune, sorted []*selection) map[int][]rune {
	hl := make(map[int][]rune)

	// Find any highlighting already applied on the line,
	// and keep the indexes so that we can skip those.
	var colors [][]int

	colorMatch := regexp.MustCompile(`\x1b\[[0-9;]+m`)
	colors = colorMatch.FindAllStringIndex(string(line), -1)

	// marks that started highlighting, but not done yet.
	regions := make([]*selection, 0)
	pos := -1
	skip := 0

	// Build the string.
	for rawIndex := range line {
		var posHl []rune
		var newHl *selection

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
		pos++

		// First check if we have a new highlighter to apply
		for _, hl := range sorted {
			if hl.bpos == pos {
				newHl = hl
				regions = append(regions, hl)
			}
		}

		// Add new colors if any, and reset if some are done.
		regions, posHl = rl.hlAdd(regions, posHl, pos)
		posHl = rl.hlReset(regions, newHl, posHl)

		// Add to the line, with the raw index since
		// we must take into account embedded colors.
		if len(posHl) > 0 {
			hl[rawIndex] = posHl
		}
	}

	return hl
}

func (rl *Instance) hlAdd(regions []*selection, line []rune, pos int) ([]*selection, []rune) {
	for i, reg := range regions {
		if reg.epos == pos {
			regions = append(regions[:i], regions[i+1:]...)
			if reg.fg != "" {
				line = append(line, []rune(seqFgDefault)...)
			}
			if reg.bg != "" {
				line = append(line, []rune(seqBgDefault)...)
			}
		}
	}

	return regions, line
}

func (rl *Instance) hlReset(regions []*selection, newHl *selection, line []rune) []rune {
	if newHl != nil {
		line = append(line, []rune(newHl.bg)...)
		line = append(line, []rune(newHl.fg)...)
	} else if len(regions) > 0 {
		backHl := regions[len(regions)-1]
		line = append(line, []rune(backHl.bg)...)
		line = append(line, []rune(backHl.fg)...)
	}

	return line
}

func (rl *Instance) resetRegions() {
	rl.marks = make([]*selection, 0)
}
