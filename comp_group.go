package readline

// CompletionGroup - A group of items offered to completion, by category.
// The output, if there are multiple groups available for a given completion input,
// will look like ZSH's completion system.
// The type is exported, because it will be easier to populate groups from the outside,
// then gather them and pass them as parameters to the TabCompleter function.
type CompletionGroup struct {
	Name        string
	Description string

	// Same as readline old system
	Suggestions  []string
	Descriptions map[string]string // Items descriptions
	DisplayType  TabDisplayType    // Map, list or normal
	MaxLength    int               // Each group can be limited in the number of comps offered

	// Values used by the shell
	tcPosX      int
	tcPosY      int
	tcMaxX      int
	tcMaxY      int
	tcOffset    int
	tcMaxLength int // Used when display is map/list, for determining message width

	// allowCycle - is true if we want to cycle through suggestions because they overflow MaxLength
	// This is set by the shell when it has detected this group is alone in the suggestions.
	// Might be the case of things like remote processes .
	allowCycle bool

	current bool // This is to say we are currently cycling through this group, for highlighting choice
}

// Because the group might have different display types, we have to init and setup for the one desired
func (g *CompletionGroup) init(rl *Instance) {

	// Details common to all displays
	g.checkCycle(rl) // Based on the number of groups given to the shell, allows cycling or not

	// Details specific to tab display modes
	switch g.DisplayType {

	case TabDisplayGrid:
		g.initGrid(rl)

	case TabDisplayMap:
		g.initMap(rl)

	case TabDisplayList:
		g.initMap(rl)
	}

	// Here, handle all things for completion search functions
}

// initGrid - Grid display details
func (g *CompletionGroup) initGrid(rl *Instance) {

	// Max number of suggestions per line, for this group
	tcMaxLength := 1
	for i := range g.Suggestions {
		if len(rl.tcPrefix+g.Suggestions[i]) > tcMaxLength {
			tcMaxLength = len([]rune(rl.tcPrefix + g.Suggestions[i]))
		}
	}

	g.tcPosX = 1
	g.tcPosY = 1
	g.tcOffset = 0

	g.tcMaxX = GetTermWidth() / (tcMaxLength + 2)
	if rl.tcMaxX < 1 {
		rl.tcMaxX = 1 // avoid a divide by zero error
	}
	if g.MaxLength == 0 {
		g.MaxLength = 10 // Handle default value if not set
	}
	g.tcMaxY = g.MaxLength
}

// initMap - Map display details
func (g *CompletionGroup) initMap(rl *Instance) {

	// Max number of suggestions per line, for this group
	// Here, we have decided that tcMaxLength is managed by group, and not rl
	// Therefore we might have made a mistake. Keep that in mind
	g.tcMaxLength = 1
	for i := range g.Suggestions {
		if g.DisplayType == TabDisplayList {
			if len(rl.tcPrefix+g.Suggestions[i]) > g.tcMaxLength {
				g.tcMaxLength = len([]rune(rl.tcPrefix + g.Suggestions[i]))
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
		// if len(suggestions) > rl.MaxTabCompleterRows {
		g.tcMaxY = g.MaxLength
		// rl.tcMaxY = rl.MaxTabCompleterRows
	} else {
		g.tcMaxY = len(g.Suggestions)
		// rl.tcMaxY = len(suggestions)
	}
}

// checkCycle - Based on the number of groups given to the shell, allows cycling or not
func (g *CompletionGroup) checkCycle(rl *Instance) {

	if len(rl.atcGroups) == 1 {
		g.allowCycle = true
	}

	// 5 different groups might be a good but conservative beginning.
	if len(rl.atcGroups) >= 5 {
		g.allowCycle = false
	}

}
