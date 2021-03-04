package readline

import "regexp"

func (rl *Instance) getHintText() {
<<<<<<< HEAD

	if !rl.modeAutoFind && !rl.modeTabFind {
		// Return if no hints provided by the user/engine
		if rl.HintText == nil {
			rl.resetHintText()
			return
		}
		// The hint text also works with the virtual completion line system.
		// This way, the hint is also refreshed depending on what we are pointing
		// at with our cursor.
		rl.hintText = rl.HintText(rl.getCompletionLine())
	}
=======
	if rl.HintText == nil {
		rl.resetHintText()
		return
	}

	// The hint text also works with the virtual completion line system.
	// This way, the hint is also refreshed depending on what we are pointing
	// at with our cursor.
	rl.hintText = rl.HintText(rl.getCompletionLine())
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
}

func (rl *Instance) writeHintText() {
	if len(rl.hintText) == 0 {
		rl.hintY = 0
		return
	}

	width := GetTermWidth()

	re := regexp.MustCompile(`\r?\n`)
	newlines := re.Split(string(rl.hintText), -1)
<<<<<<< HEAD
	offset := len(newlines)
=======
	offset := len(newlines) - 1
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae

	wrapped, hintLen := WrapText(string(rl.hintText), width)
	offset += hintLen
	rl.hintY = offset

	hintText := string(wrapped)

<<<<<<< HEAD
	if len(hintText) > 0 {
		print("\r" + rl.HintFormatting + string(hintText) + seqReset)
	}
=======
	// I HAVE PUT THIS OUT, AS I'M NOT SURE WE REALLY NEED IT
	// if rl.modeTabCompletion && !rl.modeTabFind {
	//         cell := (rl.tcMaxX * (rl.tcPosY - 1)) + rl.tcOffset + rl.tcPosX - 1
	//         description := rl.tcDescriptions[rl.tcSuggestions[cell]]
	//         if description != "" {
	//                 hintText = []rune(description)
	//         }
	// }

	print("\r\n" + rl.HintFormatting + string(hintText) + seqReset)
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
}

func (rl *Instance) resetHintText() {
	rl.hintY = 0
	rl.hintText = []rune{}
}
<<<<<<< HEAD
=======

// func (rl *Instance) writeHintText() {
//         if len(rl.hintText) == 0 {
//                 rl.hintY = 0
//                 return
//         }
//
//         width := GetTermWidth()
//
//        // Determine how many lines hintText spans over
//         // (Currently there is no support for carridge returns / new lines)
//         hintLength := strLen(string(rl.hintText))
//         n := float64(hintLength) / float64(width)
//         if float64(int(n)) != n {
//                 n++
//         }
//         rl.hintY = int(n)
//
//         if rl.hintY > 3 {
//                 rl.hintY = 3
//                 rl.hintText = rl.hintText[:(width*3)-4]
//                 rl.hintText = append(rl.hintText, '.', '.', '.')
//         }
//         hintText := rl.hintText
//
//         // I HAVE PUT THIS OUT, AS I'M NOT SURE WE REALLY NEED IT
//         // if rl.modeTabCompletion && !rl.modeTabFind {
//         //         cell := (rl.tcMaxX * (rl.tcPosY - 1)) + rl.tcOffset + rl.tcPosX - 1
//         //         description := rl.tcDescriptions[rl.tcSuggestions[cell]]
//         //         if description != "" {
//         //                 hintText = []rune(description)
//         //         }
//         // }
//
//         print("\r\n" + rl.HintFormatting + string(hintText) + seqReset)
// }
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
