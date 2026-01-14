# Display Engine Refactoring Plan

## Goals
- Fix cursor positioning issues in multiline editing.
- Ensure robust rendering of prompts, input, and helpers.
- Standardize the display sequence to avoid "lost cursor" states.

## Components
1.  **Prompts:** Primary (PS1), Secondary (PS2/`└ `), Multiline Columns (`│ `), Right/Tooltip.
2.  **Input:** Buffer text, Syntax Highlighting, Visual Selection, Auto-suggestions.
3.  **Helpers:** Hints, Completions.

## Proposed Rendering Sequence (Refresh Cycle)

1.  **Preparation & Coordinates**
    *   Hide Cursor.
    *   Reset Cursor to start of input area (after Primary Prompt).
    *   Compute Coordinates: `StartPos`, `LineHeight`, `CursorPos` (row/col).
    *   Check for buffer height changes to clear potential artifacts below.

2.  **Primary Prompt**
    *   Reprint only if invalidated (e.g., transient prompt, clear screen).
    *   Otherwise, assume cursor starts at `StartPos`.

3.  **Input Area Rendering**
    *   **Input Line:** Print the full input buffer (highlighted + auto-suggestion). Cursor ends at the end of the input text.
    *   **Right Prompt:**
        *   Calculate position relative to the end of the input.
        *   Move cursor, print prompt, restore cursor to end of input.
    *   **Multiline Indicators (Columns/Secondary):**
        *   Iterate through input lines.
        *   Move to start of each line.
        *   Print column indicators (`│ `, numbers) and secondary prompts (`└ `).
        *   **Crucial:** Return cursor to the **end of the input text** after this pass.

4.  **Helpers Rendering**
    *   Move cursor below the last line of input.
    *   **Hints:** Print hints.
    *   **Completions:** Print completion menu/grid.
    *   **Cleanup:** Clear any remaining lines below if the new render is shorter.

5.  **Final Cursor Positioning**
    *   Move cursor from bottom of helpers -> End of Input.
    *   Move cursor from End of Input -> Actual Cursor Position (using `CursorPos`).
    *   Show Cursor.
