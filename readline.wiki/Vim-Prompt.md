
## Prompt system 

### Single-line prompts

The single-line prompt is the default one. The following combination, with Vim mode and status enabled,
```go
c.shell.Multiline = false
c.shell.SetPrompt("readline")
```
will give you the following prompt (note that the Vim status indicator is always appended at the beginning)
```
[I] readline > 
```
Even in single-mode, you specify the prompt sign (the arrow above is a default):
```go
c.shell.MultilinePrompt = " $"
```
```
[I] readline $
```

Emacs will produce the same result, without the Vim status indicator.


### 2-line prompts

Setting up a 2-line prompt has only one difference:
```go
c.shell.Multiline = true
```

Combined with the settings above, will give this:
```
readline
[I] $
```

Therefore `SetPrompt()` will only control the first line, and `MultilinePrompt` and `ShowVimMode` will control 
behavior for the second line. You can obviously set them at any moment: the next readline execution loop will recompute.
