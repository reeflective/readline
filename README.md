# Readline: A fork from [lmorg/readline](https://github.com/lmorg/readline) console.

The original project by lmorg is absolutely great. The console already has superb capabilities,
and implemented in such a way that for everything that follows, I only needed 2 days.

This fork is version of the console that enhances various things:
- **Better completion system** *à-la-ZSH* offering multiple categories of completions, each with their
   own display types, concurrently.
- **Better Vim mode and shell refresh**. This means that for instance, the shell gives a clear and real-time 
   indication of the current Vim mode. Overall, enhanced support for live refresh of the whole prompt line.

## TODO 

- A simple function for refreshing the input line *ONLY FOR THOSE KNOWING WHAT THEY DO*
    instance.RefreshPrompt([]rune{})

