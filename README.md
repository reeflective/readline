
<div align="center">
  <a href="https://github.com/reeflective/readline">
    <img alt="" src="" width="600">
  </a>

  <br> <h1>  Readline </h1>

</div>


<!-- Badges -->
<p align="center">
  <a href="https://github.com/reeflective/readline/actions/workflows/go.yml/badge.svg?branch=master">
    <img src=""
      alt="Github Actions (workflows)" />
  </a>

  <a href="https://github.com/reeflective/readline">
    <img src="https://img.shields.io/github/go-mod/go-version/reeflective/readline.svg"
      alt="Go module version" />
  </a>

  <a href="https://godoc.org/reeflective/go/readline">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg"
      alt="GoDoc reference" />
  </a>

  <a href="https://goreportcard.com/report/github.com/reeflective/readline">
    <img src="https://goreportcard.com/badge/github.com/reeflective/readline"
      alt="Go Report Card" />
  </a>

  <a href="https://codecov.io/gh/reeflective/readline">
    <img src="https://codecov.io/gh/reeflective/readline/branch/master/graph/badge.svg"
      alt="codecov" />
  </a>

  <a href="https://opensource.org/licenses/BSD-3-Clause">
    <img src="https://img.shields.io/badge/License-BSD_3--Clause-blue.svg"
      alt="License: BSD-3" />
  </a>
</p>

This library is a modern, pure Go readline implementation, enhanced with editing and user 
interface features commonly found in modern shells, all in little more than 10K lines of code.
Its kemap-based model and completion engine is heavily inspired from the Z-Shell architecture.
It is used, between others, to power the [console](https://github.com/reeflective/console) library.

## Features

### Editing
- Near-native Emacs and Vim modes.
- Configurable bind keymaps, with live reload and sane defaults.
- Extended list of line edition/movement/control widgets (emacs and Vim).
- Vim Insert and Replace (once/many).
- Many Vim text objects.
- Support for Vim Visual/Operator pending mode & cursor styles indications.
- All Vim registers, with completion support.
- Extended surround select/change/add fonctionality, with highlighting.
- Keywords switching (operators, booleans, hex/binary/digit) with iterations.
- Undo/redo history.
- Command-line edition in $EDITOR.
- Support for an arbitrary number of history sources.

### Interface
- Support for most of `oh-my-posh` prompts (PS1/PS2/RPROMPT/transient/tooltip).
- Extended completion system, keymap-based and configurable, easy to populate & use.
- Multiple completion display styles, with color support.
- Completion & History incremental search system & highlighting (fuzzy-search).
- Automatic & context-aware suffix removal for efficient flags/path/list completion.
- Optional asynchronous autocomplete.
- Usage/hint message display.
- Support for syntax highlighting

## Showcases

An example application using readline can be found [here]().

- Emacs edition
- Vim edition
- Vim selection & movements
- Vim surround
- Keyword swithing
- Vim registers & completion
- Undo/redo line history
- History movements & completion
- Completion classic
- Completion isearch
- Suffix autoremoval
- Prompts
- Logging

## Credits

- While most of the code has been rewritten from scratch, the original library used is [lmorg/readline](https://github.com/lmorg/readline).
  I would have never ventured myself doing this if he had not ventured writing a Vim mode core in the first place. 
- Some of the Vim code is inspired or translated from [zsh-vi-mode](https://github.com/jeffreytse/zsh-vi-mode).
