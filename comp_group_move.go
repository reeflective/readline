package readline

// aMoveTabGridHighlight - Moves the highlighting for currently selected completion item (grid display)
func (g *CompletionGroup) aMoveTabGridHighlight(rl *Instance, x, y int) (done bool) {

	g.tcPosX += x
	g.tcPosY += y

	if g.tcPosX < 1 {
		g.tcPosX = g.tcMaxX
		g.tcPosY--
	}

	if g.tcPosX > g.tcMaxX {
		g.tcPosX = 1
		g.tcPosY++
	}

	if g.tcPosY < 1 {
		g.tcPosY = rl.tcUsedY
	}

	if g.tcPosY > rl.tcUsedY {
		g.tcPosY = 1
		return true
	}

	if (g.tcMaxX*(g.tcPosY-1))+g.tcPosX > len(g.Suggestions) {
		if x < 0 {
			g.tcPosX = len(g.Suggestions) - (g.tcMaxX * (g.tcPosY - 1))
			return true
		}

		if x > 0 {
			g.tcPosX = 1
			g.tcPosY = 1
			return true
		}

		if y < 0 {
			g.tcPosY--
			return true
		}

		if y > 0 {
			g.tcPosY = 1
			return true
		}

		return true
	}

	return false
}

// aMoveTabMapHighlight - Moves the highlighting for currently selected completion item (map/list display)
func (g *CompletionGroup) aMoveTabMapHighlight(x, y int) (done bool) {

	g.tcPosY += x
	g.tcPosY += y

	if g.tcPosY < 1 {
		g.tcPosY = 1 // We had suppressed it for some time, don't know why.
		g.tcOffset--
	}

	if g.tcPosY > g.tcMaxY {
		g.tcPosY--
		g.tcOffset++
	}

	if g.tcOffset+g.tcPosY < 1 && len(g.Suggestions) > 0 {
		g.tcPosY = g.tcMaxY
		g.tcOffset = len(g.Suggestions) - g.tcMaxY
	}

	if g.tcOffset < 0 {
		g.tcOffset = 0
	}

	if g.tcOffset+g.tcPosY > len(g.Suggestions) {
		g.tcPosY = 1
		g.tcOffset = 0
		return true
	}
	return false
}
