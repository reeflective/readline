# Temporary Notes for readline (~/code/github.com/reeflective/readline)

This file is for your private notes, prompts, and scratchpad content.
It is not directly sent to Gemini, but you can easily copy/paste from here.

The behavior in the shell is the following:

I enter a first line: `testing \`
I press enter, and the shell correctly goes to a new line.
I then type again: `testing \`
and the prompt goes back to its initial position on the first line (which means it goes one up one line too much.
Starting from there, everytime I enter a character (thus causing the shell to redisplay the line) the cursor goes up 2 lines, while it should not.

Based on this, I want you to tell me what do you suspect in the code is not working correctly.

Investigate the codebase and identify the issue.

Consider this snippet from the codebase.

	helpersMoved := e.displayHelpers()
	if helpersMoved {
		e.cursorHintToLineStart()
		e.lineStartToCursorPos()
	} else {
		e.lineEndToCursorPos()
	}


I'm pretty sure the problem is in this snippet.
I suspect that cursorHintToLineStart() or lineStartToCursorPos() are miscalculating something when
the line is a multiline string.
