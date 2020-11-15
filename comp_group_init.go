package readline

// Because the group might have different display types, we have to init and setup for the one desired
func (g *CompletionGroup) init(rl *Instance) {

	// Details common to all displays
	rl.modeTabCompletion = true
	g.checkCycle(rl) // Based on the number of groups given to the shell, allows cycling or not
	g.checkMaxLength(rl)

	// Details specific to tab display modes
	switch g.DisplayType {

	case TabDisplayGrid:
		g.initGrid(rl)
	case TabDisplayMap:
		g.initMap(rl)
	case TabDisplayList:
		g.initMap(rl)
	}
}

// initGrid - Grid display details
func (g *CompletionGroup) initGrid(rl *Instance) {

	// Max number of suggestions per line, for this group
	tcMaxLength := 1
	for i := range g.Suggestions {
		if len(g.Suggestions[i]) > tcMaxLength {
			tcMaxLength = len([]rune(g.Suggestions[i]))
		}
	}

	g.tcPosX = 1
	g.tcPosY = 1
	g.tcOffset = 0

	g.tcMaxX = GetTermWidth() / (tcMaxLength + 2)
	if g.tcMaxX < 1 {
		g.tcMaxX = 1 // avoid a divide by zero error
	}
	if g.MaxLength == 0 {
		g.MaxLength = 10 // Handle default value if not set
	}
	g.tcMaxY = g.MaxLength

}

// initMap - Map & List display details
func (g *CompletionGroup) initMap(rl *Instance) {

	// We make the map anyway, especially if we need to use it later
	if g.Descriptions == nil {
		g.Descriptions = make(map[string]string)
	}

	// Max number of suggestions per line, for this group
	// Here, we have decided that tcMaxLength is managed by group, and not rl
	// Therefore we might have made a mistake. Keep that in mind
	g.tcMaxLength = 1
	for i := range g.Suggestions {
		if g.DisplayType == TabDisplayList {
			if len(g.Suggestions[i]) > g.tcMaxLength {
				g.tcMaxLength = len([]rune(g.Suggestions[i]))
			}
		} else {
			if len(g.Descriptions[g.Suggestions[i]]) > g.tcMaxLength {
				g.tcMaxLength = len(g.Descriptions[g.Suggestions[i]])
			}
		}
	}

	g.tcPosX = 1
	g.tcPosY = 1
	g.tcOffset = 0

	g.tcMaxX = 1
	if len(g.Suggestions) > g.MaxLength {
		g.tcMaxY = g.MaxLength
	} else {
		g.tcMaxY = len(g.Suggestions)
	}
}
