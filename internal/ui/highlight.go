package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/reeflective/readline/inputrc"
	"github.com/reeflective/readline/internal/color"
	"github.com/reeflective/readline/internal/core"
)

// Highlight applies visual/selection highlighting to a line.
// The provided line might already have been highlighted by a user-provided
// highlighter: this function accounts for any embedded color sequences.
func Highlight(line []rune, selection core.Selection, config *inputrc.Config) string {
	// Sort regions and extract colors/positions.
	sorted := sortHighlights(selection)
	colors := getHighlights(line, sorted, config)

	var highlighted string

	// And apply highlighting before each rune.
	for i, r := range line {
		if highlight, found := colors[i]; found {
			highlighted += string(highlight)
		}

		highlighted += string(r)
	}

	// Finally, highlight comments using a regex.
	comment := strings.Trim(config.GetString("comment-begin"), "\"")
	commentPattern := fmt.Sprintf(`(^|\s)%s.*`, comment)

	if commentsMatch, err := regexp.Compile(commentPattern); err == nil {
		highlighted = commentsMatch.ReplaceAllString(highlighted, fmt.Sprintf("%s${0}%s", color.FgBlackBright, color.Reset))
	}

	highlighted += color.Reset

	return highlighted
}

func sortHighlights(vhl core.Selection) []core.Selection {
	all := make([]core.Selection, 0)
	sorted := make([]core.Selection, 0)
	bpos := make([]int, 0)

	for _, reg := range vhl.Surrounds() {
		all = append(all, reg)
		rbpos, _ := reg.Pos()
		bpos = append(bpos, rbpos)
	}

	all = append(all, vhl)

	if vhl.Active() && vhl.IsVisual() {
		vbpos, _ := vhl.Pos()
		bpos = append(bpos, vbpos)
	}

	sort.Ints(bpos)

	for _, pos := range bpos {
		for _, reg := range all {
			bpos, _ := reg.Pos()

			if bpos == pos && reg.Active() && reg.IsVisual() {
				sorted = append(sorted, reg)
				break
			}
		}
	}

	return sorted
}

func getHighlights(line []rune, sorted []core.Selection, config *inputrc.Config) map[int][]rune {
	highlights := make(map[int][]rune)

	// Find any highlighting already applied on the line,
	// and keep the indexes so that we can skip those.
	var colors [][]int

	colorMatch := regexp.MustCompile(`\x1b\[[0-9;]+m`)
	colors = colorMatch.FindAllStringIndex(string(line), -1)

	// marks that started highlighting, but not done yet.
	regions := make([]core.Selection, 0)
	pos := -1
	skip := 0

	// Build the string.
	for rawIndex := range line {
		var posHl []rune
		var newHl core.Selection

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
			bpos, _ := hl.Pos()

			if bpos == pos {
				newHl = hl
				regions = append(regions, hl)
			}
		}

		// Add new colors if any, and reset if some are done.
		regions, posHl = hlReset(regions, posHl, pos, config)
		posHl = hlAdd(regions, newHl, posHl, config)

		// Add to the line, with the raw index since
		// we must take into account embedded colors.
		if len(posHl) > 0 {
			highlights[rawIndex] = posHl
		}
	}

	return highlights
}

func hlReset(regions []core.Selection, line []rune, pos int, config *inputrc.Config) ([]core.Selection, []rune) {
	for i, reg := range regions {
		_, epos := reg.Pos()
		foreground, background := reg.Highlights()
		matcher := reg.Type == "matcher"

		if epos == pos {
			regions = append(regions[:i], regions[i+1:]...)

			if foreground != "" {
				line = append(line, []rune(color.FgDefault)...)
			}

			if background != "" {
				if background, _ = strconv.Unquote(config.GetString("active-region-end-color")); background == "" && !matcher {
					line = append(line, []rune(color.ReverseReset)...)
				} else {
					line = append(line, []rune(color.BgDefault)...)
				}
			}
		}
	}

	return regions, line
}

func hlAdd(regions []core.Selection, newHl core.Selection, line []rune, config *inputrc.Config) []rune {
	var (
		fg, bg  string
		matcher bool
		hl      core.Selection
	)

	if newHl.Active() {
		hl = newHl
	} else if len(regions) > 0 {
		hl = regions[len(regions)-1]
	}

	fg, bg = hl.Highlights()
	matcher = hl.Type == "matcher"

	// Update the highlighting with inputrc settings if any.
	if bg != "" && !matcher {
		if bg, _ = strconv.Unquote(config.GetString("active-region-start-color")); bg == "" {
			bg = color.Reverse
		}
	}

	// Add highlightings
	line = append(line, []rune(bg)...)
	line = append(line, []rune(fg)...)

	return line
}
