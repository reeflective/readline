
<div align="center">
  <br> <h1> Readline </h1>
  <p>  Shell library with modern and simple UI features </p>
</div>


<!-- Badges -->
<p align="center">
  <a href="https://github.com/reeflective/readline/actions/workflows/go.yml">
    <img src="https://github.com/reeflective/readline/actions/workflows/go.yml/badge.svg?branch=master"
      alt="Github Actions (workflows)" />
  </a>

  <a href="https://github.com/reeflective/readline">
    <img src="https://img.shields.io/github/go-mod/go-version/reeflective/readline.svg"
      alt="Go module version" />
  </a>

  <a href="https://pkg.go.dev/github.com/reeflective/readline">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg"
      alt="GoDoc reference" />
  </a>

  <a href="https://goreportcard.com/report/github.com/reeflective/readline">
    <img src="https://goreportcard.com/badge/github.com/reeflective/readline"
      alt="Go Report Card" />
  </a>

  <a href="https://codecov.io/gh/reeflective/readline">
    <img src="https://codecov.io/gh/reeflective/readline/branch/main/graph/badge.svg"
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

### Core 
- Pure Go, almost-only standard library
- Full .inputrc support (all commands/options)
- Extensive test suite and full coverage of core code
- [Extended list](https://github.com/reeflective/readline/wiki/Keymaps-&-Widgets) of additional commands/options (edition/completion/history)
- Complete multiline edition/movement support
- Command-line edition in `$EDITOR`/`$VISUAL` support
- Programmable API, with failure-safe access to core components
- Support for an arbitrary number of history sources

### Emacs / Standard
- Native Emacs commands
- Emacs-style macro engine (not working accross multiple calls)
- Keywords switching (operators, booleans, hex/binary/digit) with iterations
- Command/mode cursor status indicator
- Complete undo/redo history
- Command status/arg/iterations hint display

### Vim
- Near-native Vim mode
- Vim text objects (code blocks/words/blank/shellwords)
- Extended surround select/change/add fonctionality, with highlighting
- Vim Visual/Operator pending mode & cursor styles indications
- Vim Insert and Replace (once/many)
- All Vim registers, with completion support
- Vim-style macro recording (`q<a>`) and invocation (`@<a>`)

### Interface
- Support for PS1/PS2/RPROMPT/transient/tooltip prompts (compatible with [oh-my-posh](https://github.com/JanDeDobbeleer/oh-my-posh))
- Extended completion system, keymap-based and configurable, easy to populate & use
- Multiple completion display styles, with color support.
- Completion & History incremental search system & highlighting (fuzzy-search).
- Automatic & context-aware suffix removal for efficient flags/path/list completion.
- Optional asynchronous autocomplete
- Builtin & programmable syntax highlighting


## Documentation

Readline is used by the [console library](https://github.com/reeflective/console) and its [example binary](https://github.com/reeflective/console/tree/main/example). To get a grasp of the 
functionality provided by readline and its default configuration, install and start the binary.

* [Introduction](https://github.com/reeflective/readline/wiki/Introduction-&-Features)
* [Configuration file](https://github.com/reeflective/readline/wiki/Configuration-File)
* [Keymaps & Widgets](https://github.com/reeflective/readline/wiki/Keymaps-&-Widgets)
* [Prompts](https://github.com/reeflective/readline/wiki/Prompts)
* [History Sources](https://github.com/reeflective/readline/wiki/History-Sources)
* [Vim mode](https://github.com/reeflective/readline/wiki/Vim-Mode)
* [Custom callbacks](https://github.com/reeflective/readline/wiki/Custom-Callbacks)
* [Completions & Hints](https://github.com/reeflective/readline/wiki/Completions-&-Hints)
* [Other features](https://github.com/reeflective/readline/wiki/Other-Features)


## Showcases

<details>
  <summary>- Emacs edition</summary>
 <dd><em>(This extract is quite a pity, because its author is not using Emacs and does not know many of its shortcuts)</em></dd>
<img src="https://github.com/reeflective/readline/blob/assets/emacs.gif"/>
</details>
<details>
  <summary>- Vim edition</summary>
<img src="https://github.com/reeflective/readline/blob/assets/vim.gif"/>
</details>
<details>
  <summary>- Undo/redo line history </summary>
<img src="https://github.com/reeflective/readline/blob/assets/undo.gif"/>
</details>
<details>
  <summary>- Keyword switching </summary>
<img src="https://github.com/reeflective/readline/blob/assets/switch-keywords.gif"/>
</details>
<details>
  <summary>- Vim selection & movements (basic) </summary>
<img src="https://github.com/reeflective/readline/blob/assets/vim-selection.gif"/>
</details>
<details>
  <summary>- Vim surround (selection and change) </summary>
 <dd><em>Basic surround selection changes/adds</em></dd>
<img src="https://github.com/reeflective/readline/blob/assets/vim-surround.gif"/>
 <dd><em>Surround and change in shellwords, matching brackets, etc.</em></dd>
<img src="https://github.com/reeflective/readline/blob/assets/vim-surround-2.gif"/>
</details>
<details>
  <summary>- Vim registers (with completion) </summary>
<img src="https://github.com/reeflective/readline/blob/assets/vim-registers.gif"/>
</details>
<details>
  <summary>- History movements/completion/use/search </summary>
 <dd><em></em></dd>
History movement, completion and some other other widgets
<img src="https://github.com/reeflective/readline/blob/assets/history.gif"/>
 <dd><em>History cycling and search</em></dd>
<img src="https://github.com/reeflective/readline/blob/assets/history-search.gif"/>
</details>
<details>
  <summary>- Completion </summary>
 <dd><em>Classic mode & incremental search mode</em></dd>
<img src="https://github.com/reeflective/readline/blob/assets/completion.gif"/>
 <dd><em>Smart terminal estate management</em></dd>
<img src="https://github.com/reeflective/readline/blob/assets/completion-size.gif"/>
</details>
<details>
  <summary>- Suffix autoremoval </summary>
<img src="https://github.com/reeflective/readline/blob/assets/suffix-autoremoval.gif"/>
</details>
<details>
  <summary>- Prompts </summary>
<img src="https://github.com/reeflective/readline/blob/assets/prompts.gif"/>
</details>
<details>
  <summary>- Logging </summary>
<img src="https://github.com/reeflective/readline/blob/assets/logging.gif"/>
</details>
<details>
  <summary>- Configuration hot-reload </summary>
<img src="https://github.com/reeflective/readline/blob/assets/configuration-reload.gif"/>
</details>


## Status

This library is in a pre-release status, although pretending to be quite bug-free as compared to its feature set.
Please open a PR or an issue if you wish to bring enhancements to it. 
Other contributions, as well as bug fixes and reviews are also welcome.

## Credits

- While most of the code has been rewritten from scratch, the original library used 
  is [lmorg/readline](https://github.com/lmorg/readline). I would have never ventured myself doing this if he had not 
  ventured writing a Vim mode core in the first place. 
- Some of the Vim code is inspired or translated from [zsh-vi-mode](https://github.com/jeffreytse/zsh-vi-mode).
