package readline

const (
	vimInsertStr      = "[I]"
	vimReplaceOnceStr = "[V]"
	vimReplaceManyStr = "[R]"
	vimDeleteStr      = "[D]"
	vimKeysStr        = "[N]"
)

func (rl *Instance) refreshVimStatus() {
	rl.Prompt.compute(rl)
	rl.updateHelpers()
}

// viHintMessage - lmorg's way of showing Vim status is to overwrite the hint.
// Currently not used, as there is a possibility to show the current Vim mode in the prompt.
func (rl *Instance) viHintMessage() {
	// switch rl.modeViMode {
	// case vimKeys:
	// 	rl.hintText = []rune("-- VIM KEYS -- (press `i` to return to normal editing mode)")
	// case vimInsert:
	// 	rl.hintText = []rune("-- INSERT --")
	// case vimReplaceOnce:
	// 	rl.hintText = []rune("-- REPLACE CHARACTER --")
	// case vimReplaceMany:
	// 	rl.hintText = []rune("-- REPLACE --")
	// case vimDelete:
	// 	rl.hintText = []rune("-- DELETE --")
	// default:
	// 	rl.getHintText()
	// }

	rl.clearHelpers()
	rl.renderHelpers()
}
