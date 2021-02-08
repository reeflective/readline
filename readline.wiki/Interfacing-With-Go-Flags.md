
Coming back to the little function we declared in the [introduction](https://github.com/maxlandon/readline/wiki/Embedding-Readline-In-A-Project):

## Working Setup

Wit the following setup we have a 2-line prompt with Vim input and Vim status refresh:
```go
func (c *console) setup() (err error) {

	// Input mode & defails
	c.shell.InputMode = readline.Vim // Could be readline.Emacs for emacs input mode.
	c.shell.ShowVimMode = true
	c.shell.VimModeColorize = true

	// Prompt: we want a two-line prompt, with a custom indicator after the Vim status
	c.shell.SetPrompt("readline ")
	c.shell.Multiline = true
	c.shell.MultilinePrompt = " > "

	// History: by default the history is in-memory, use it with Ctrl-R
        c.shell.AltHistory = &AnotherHistorySource{}

	return
}

```

## Declaring a root Command Parser

Now we want to use this library with commands, obviously, and we directly want to 
use the default tab/hint and syntax completers given in the `completers` package.

We use, to this end, the [go-flags](https://github.com/jessevdk/go-flags) library for declaring:

- A `Parser` type, which will handle dispatching everything to commands and subcommands, flags, etc.
- One or more `Command`, which are nothing more than structs parsed by go-flags.

In fact, everyt `Command` embeds a `Parser` type, so that it can have any number of subcommands.
Declaring the root parser, with several options enabled:
```go

func BindParser() (parser  *flags.Parser){

        // A new root parser
        parser = flags.NewNamedParser("application", flags.Default)

        // Setup a few options
        // IgnoreUnkown - All unknown arguments are passed to the parser/command []args
        // HelpFlag     - Adds a -h/--help flag to the parser and each of its commands
        parser.Options = flags.IgnoreUnkown | flags.HelpFlag

        // NOTE: see the function print help below, used when parsing any error from the parser,
        // after executing the commmand: (help flag is an error for go-flags).

        // We actually bind the underlying commands to the parser, see later...
        err := bindCommands(parser)
        if err != nil {
                // An error when registering your structs, see the go-flags library
        }
}
```

This comand, as you will see below, will be called on each loop execution: 
therefore, you can add new commands to a parser, change it all together, etc.


## Using the default Completion Engine 

The readline library has a default completion engine that powers tab completions, hints and syntax highlighting:
It accepts a `*Parser` type, which is used for everything: you can bind everyting like this:

```go
// setup - The console sets up various elements such as the completion system, hints,
// syntax highlighting, prompt system, commands binding, and client environment loading.
func (c *console) setup() (err error) {

	// Input mode & Prompt defails...

	// Instantiate a default completer associated with the parser
	// declared in commands.go, and embedded into the console struct.
	// The error is muted, because we don't pass an nil parser, therefore no problems.
	defaultCompleter, _ := completers.NewCommandCompleter(c.parser)

	// Register the completer for command/option completions, hints and syntax highlighting.
	c.shell.TabCompleter = defaultCompleter.TabCompleter
	c.shell.HintText = defaultCompleter.HintCompleter
	c.shell.SyntaxHighlighter = defaultCompleter.SyntaxHighlighter

	return
}
```

As well, if you happen to swap the parser itself (not the commands in it, to be clear) from the readline shell
at the next loop, you can produce a new completer with function above again.

Then, your are ready to bind the parser, and start.


## Binding the Parser to readline, and Starting

Because readline runs in a loop, there is no defined execution lifetime for our command structs, 
and we need to reset all of their arguments/option values. In addition, we might want to swap the parser
after some command executions, or register new commands to it. 

Therefore, the parser is called before each readline execution loop, in the following function.
Includes a bit of parsing and expansion logic, but you can get rid of it.

```go

// Start the readline shell loop
func (c *console) Start() (err error) {

	// Setup console elements (prompt, completers and so on)
	err = c.setup()
	if err != nil {
		return fmt.Errorf("Console setup failed: %s", err)
	}

	// Start input loop
	for {
                // Bind the command parser to the console, before starting to read user input
                c.parser = BindParser()

                // If you have changing parsers you would call setup again, here
                c.setup() // Refreshes completers, etc

                // Start reading the user input, yield completions, etc.
		line, _ := c.Readline()

		// Split and sanitize input
		sanitized, empty := sanitizeInput(line)
		if empty {
			continue
		}

		// Process various tokens on input (environment variables, paths, etc.)
		// These tokens will be expaneded by completers anyway, so this is not absolutely required.
		envParsed, _ := completers.ParseEnvironmentVariables(sanitized)

		// Other types of tokens, needed by commands who expect a certain type
		// of arguments, such as paths with spaces.
		tokenParsed := c.parseTokens(envParsed)

		// Let the parser execute the command 
		if _, parserErr := c.parser.ParseArgs(tokenParsed); parserErr != nil {

                        // If the parser encountered an error, or has receive the error help flag
                        // we handle this in the custom function you'll find below
                        c.HandleParserErrors(c.parser, parserErr, args)
		}
	}
}

```


## Practical functions (handle parser behavior)

The following functions show how to handle default errors raised by a go-flags Parser.

```go
// HandleParserErrors - The parsers may return various types of Errors, this function handles them.
func (c *console) HandleParserErrors(parser *flags.Parser, in error, args []string) (err error) {

	// If there is an error, cast it to a parser error, else return
	var parserErr *flags.Error
	if in == nil {
		return
	}
	parserErr, ok := in.(*flags.Error)
	if !ok {
		return
	}
	if parserErr == nil {
		return
	}

	// If command is not found, handle special (either through OS shell, or exits, etc.)
	if parserErr.Type == flags.ErrUnknownCommand {
                return err // Or do custom logic at this point
	}

	// If the error type is a detected -h, --help flag, print custom help.
	if parserErr.Type == flags.ErrHelp {
		cmd := c.findHelpCommand(args, parser)

		// If command is nil, it means the help was requested as
		// the menu help: print all commands for the context.
		if cmd == nil {
			PrintMenuHelp(parser) // See examples/help.go file for the function
			return
		}

		// Else print the help for a specific command
		PrintCommandHelp(cmd) // See examples/help.go file for the function
		return
	}

	// Else, we print the raw parser error
	fmt.Println(ParserError + parserErr.Error())

	return
}

// findHelpCommand - A -h, --help flag was invoked in the output.
// Find the root or any subcommand.
func (c *console) findHelpCommand(args []string, parser *flags.Parser) *flags.Command {

	var root *flags.Command
	for _, cmd := range parser.Commands() {
		if cmd.Name == args[0] {
			root = cmd
		}
	}
	if root == nil {
		return nil
	}
	if len(args) == 1 || len(root.Commands()) == 0 {
		return root
	}

	var sub *flags.Command
	if len(args) > 1 {
		for _, s := range root.Commands() {
			if s.Name == args[1] {
				sub = s
			}
		}
	}
	if sub == nil {
		return root
	}
	if len(args) == 2 || len(sub.Commands()) == 0 {
		return sub
	}

	return nil
}


```
