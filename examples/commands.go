package main

import "github.com/jessevdk/go-flags"

// This file declares a go-flags parser and a few commands.

var commandParser = flags.NewNamedParser("example", flags.Default)
