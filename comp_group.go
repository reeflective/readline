package readline

// CompletionGroup - A group/category of items offered to completion, with its own
// name, descriptions and completion display format/type.
// The output, if there are multiple groups available for a given completion input,
// will look like ZSH's completion system.
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
	isCurrent  bool // This is to say we are currently cycling through this group, for highlighting choice
}

// updateTabFind - When searching through all completion groups (whether it be command history or not),
// we ask each of them to filter its own items and return the results to the shell for aggregating them.
// The rx parameter is passed, as the shell already checked that the search pattern is valid.
func (g *CompletionGroup) updateTabFind(rl *Instance) {

	suggs := make([]string, 0)

	for i := range g.Suggestions {
		if rl.regexSearch.MatchString(g.Suggestions[i]) {
			suggs = append(suggs, g.Suggestions[i])
		} else if g.DisplayType == TabDisplayList && rl.regexSearch.MatchString(g.Descriptions[g.Suggestions[i]]) {
			// this is a list so lets also check the descriptions
			suggs = append(suggs, g.Suggestions[i])
		}
	}

	// We overwrite the group's items, (will be refreshed as soon as something is typed in the search)
	g.Suggestions = suggs

	// Finally, the group computes its new printing settings
	g.init(rl)
}

// checkCycle - Based on the number of groups given to the shell, allows cycling or not
func (g *CompletionGroup) checkCycle(rl *Instance) {

	if len(rl.tcGroups) == 1 {
		g.allowCycle = true
	}

	// 5 different groups might be a good but conservative beginning.
	if len(rl.tcGroups) >= 5 {
		g.allowCycle = false
	}

}

// checkMaxLength - Based on the number of groups given to the shell, check/set MaxLength defaults
func (g *CompletionGroup) checkMaxLength(rl *Instance) {

	// This means the user forgot to set it
	if g.MaxLength == 0 {
		if len(rl.tcGroups) < 5 {
			g.MaxLength = 10
		}

		// 5 different groups might be a good but conservative beginning.
		if len(rl.tcGroups) >= 5 {
			g.MaxLength = 7
		}
	}
}
