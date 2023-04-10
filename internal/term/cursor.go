package term

func MoveCursorUp(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dA", i)
}

func MoveCursorDown(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dB", i)
}

func MoveCursorForwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dC", i)
}

func MoveCursorBackwards(i int) {
	if i < 1 {
		return
	}

	printf("\x1b[%dD", i)
}
