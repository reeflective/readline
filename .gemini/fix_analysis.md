# Fix Analysis: Multiline Display & Cursor Positioning

## The Issue
The cursor jumps to the top of the input area and fails to return to the bottom during multiline editing, specifically when `multiline-column` options are disabled.

## Analysis of `MultilineColumnPrint`
The function `MultilineColumnPrint` behaves differently based on configuration:
1.  **`multiline-column-numbered` (ON):**
    *   Iterates through all lines.
    *   Prints a string containing `\n` for each line.
    *   **Effect:** The cursor physically moves down `N` lines (where `N` is the number of logical lines).
2.  **`multiline-column` / Default (OFF):**
    *   The loop body or case is not entered (or returns empty string).
    *   **Effect:** The function prints nothing. The cursor does **not** move.

## The Logic Flaw in `displayMultilinePrompts` (or `Refresh`)
The calling code typically looks like this:
```go
if e.line.Lines() > 1 {
    term.MoveCursorUp(e.lineRows)         // Move to Top
    e.prompt.MultilineColumnPrint()       // Print Columns (Expectation: Moves Down)
    // Missing: Explicit return to bottom if Print() didn't move us.
}
```

*   **Scenario A (Numbered ON):** `MoveCursorUp` moves up. `Print` moves down (mostly). The cursor ends up near the bottom. The error is small (difference between logical newlines and wrapped rows).
*   **Scenario B (All OFF):** `MoveCursorUp` moves up. `Print` does nothing. The cursor stays at the top. **This is the bug.**

## Proposed Fix
The fix requires two adjustments:
1.  **Conditional Execution:** Only perform the "Move Up -> Print" sequence if a column mode is actually enabled. If disabled, there is no need to move up just to print nothing.
2.  **Cursor Correction:** When enabled, ensure the cursor returns to the correct bottom position by accounting for line wrapping.

**Algorithm:**
1.  Check if `multiline-column`, `numbered`, or `custom` is enabled.
2.  **If Enabled:**
    *   Move Cursor Up `e.lineRows`.
    *   Call `MultilineColumnPrint`.
    *   Move Cursor Down `e.lineRows - e.line.Lines()`. (Correction for wrapped lines vs logical newlines).
3.  **If Disabled:**
    *   Skip the block. The cursor remains at the bottom (where `displayLine` left it).

## Secondary Prompts (`└ `)
The current logic only prints the Secondary Prompt (`PS2`) for the *current* (last) line. Previous lines display the column indicator (`|` or number). If columns are disabled, previous lines display nothing. This appears to be intended behavior, or at least separate from the cursor jump bug.

```