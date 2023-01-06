package readline

import (
	_ "embed" // We embed the default configuration file
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

const (
	configFileName = "readline.yml"
)

//go:embed readline.yml
var defaultConfig string

// ErrNoSystemConfig indicates that no readline configuration file
// could be found in  any of the default user system paths:
//
// $XDG_CONFIG_HOME/reeflective/.reeflective.yml
// $HOME/.reeflective.yml.
//
var ErrNoSystemConfig = errors.New("no user configuration found in user directories")

// config stores all configurable elements for the shell, including the
// complete list of keymaps. The configuration is always written/exported
// as a YAML file, and any file to be imported as a configuration is also
// unmarshaled as YAML.
type config struct {
	rl   *Instance
	node *yaml.Node // Stores the configuration file bytes including comments.

	// InputMode - The shell can be used in Vim editing mode, or Emacs (classic).
	InputMode InputMode `yaml:"inputMode"`
	Vim       struct {
		// Cursors
		InsertCursor          CursorStyle `yaml:"insertCursor"`
		NormalCursor          CursorStyle `yaml:"normalCursor"`
		OperatorPendingCursor CursorStyle `yaml:"operatorPendingCursor"`
		VisualCursor          CursorStyle `yaml:"visualCursor"`
		ReplaceCursor         CursorStyle `yaml:"replaceCursor"`
	}
	Emacs struct {
		Cursor CursorStyle
	}
	Keymaps map[keymapMode]keymap

	// PromptTransient enables the use of transient prompt.
	PromptTransient bool `yaml:"promptTransient"`
	// The shell displays fish-like autosuggestions (the first
	// matching history line is displayed dimmed).
	HistoryAutosuggest bool `yaml:"historyAutosuggest"`
	// HistoryAutoWrite defines whether items automatically get written to history.
	// Enabled by default. Set to false to disable.
	HistoryAutoWrite bool `yaml:"historyAutoWrite"`
	// The maximum number of completion rows to print at any time.
	MaxTabCompleterRows int `yaml:"maxTabCompleterRows"`
	// Autocomplete asynchrously generates completions as text is typed in the line.
	AutoComplete bool `yaml:"autoComplete"`
}

// Config returns the current configuration of the readline instance.
func (rl *Instance) Config() *config {
	return rl.config
}

// Load looks for a configuration at the specified path (including file name).
// It returns an error either if the file is not found, if the shell fails to read it,
// or if its fails to unmarshal/load it.
func (c *config) Load(path string) (err error) {
	if err = c.load(path); err != nil {
		return err
	}

	// Successfully read the file, so watch for changes.
	go c.rl.watchConfig(path)

	return nil
}

func (c *config) load(path string) (err error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// First load the yaml as node with comments
	if err = yaml.Unmarshal(bytes, c.node); err != nil {
		return
	}

	// And then unmarshal the node onto the struct.
	if err = c.node.Decode(c); err != nil {
		return
	}

	// Rebind all widgets
	c.rl.bindWidgets()
	c.rl.loadInterruptHandlers()
	c.rl.updateKeymaps()
	c.rl.updateCursor()

	return nil
}

// LoadFromBytes loads a configuration from a bytes array.
// It returns an error either if the shell fails to read it,
// or if its fails to unmarshal/load it.
func (c *config) LoadFromBytes(config []byte) (err error) {
	// First load the yaml as node with comments
	if err = yaml.Unmarshal(config, c.node); err != nil {
		return
	}

	// And then unmarshal the node onto the struct.
	return c.node.Decode(c)
}

// LoadSystem looks for some default places in the user filesystem for
// a configuration file, and if one is found, stops looking for further paths and loads it.
// It returns an error either if the file is not found in any of those paths, if the shell
// fails to read it, or if its fails to unmarshal/load it.
//
// The shell looks for the following paths, in this order:
// $XDG_CONFIG_HOME/reeflective/readline.yml
// $HOME/.readline.yml.
//
func (c *config) LoadSystem() (err error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		xdgConfigHomePath := filepath.Join(xdgConfigHome, "reeflective", configFileName)
		if _, err := os.Stat(xdgConfigHomePath); err == nil {
			return c.Load(xdgConfigHomePath)
		}
	}

	userPath := os.Getenv("HOME")
	configHomePath := filepath.Join(userPath, "."+configFileName)
	if _, err := os.Stat(configHomePath); err == nil {
		return c.Load(configHomePath)
	}

	return ErrNoSystemConfig
}

// Save saves the current shell configuration to the specified path.
// If the path is a directory, the configuration file is saved as '$path/.readline.yml'.
// If any directory in the path does not exist, they are created (equivalent to mkdir -p).
// It returns an error if the config cannot be written, or if the shell fails to marshal it.
func (c *config) Save(path string) (err error) {
	// Create the complete path (including file) if it does not exist
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return
		}
	}

	// If the path is a directory, append the default filename to it, and create it.
	if file, ferr := os.Stat(path); ferr == nil && file.IsDir() {
		path = filepath.Join(path, "."+configFileName)
	}

	// We have a complete path, marshal and write the configuration.
	config, err := c.Export()
	if err != nil {
		return err
	}

	return os.WriteFile(path, config, os.FileMode(0o644))
}

// SaveSystem saves the current configuration to one of the default
// locations (identical to to those looked up by LoadConfigUser()).
//
// It checks paths in this order:
// - If $XDG_CONFIG_HOME is defined, saves to $XDG_CONFIG_HOME/reeflective/readline.yml
// - Else, saves to $HOME/.readline.yml.
//
func (c *config) SaveSystem() (err error) {
	var path string

	// In home by default.
	userPath := os.Getenv("HOME")
	path = filepath.Join(userPath, "."+configFileName)

	// Or in configuration directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		path = filepath.Join(xdgConfigHome, "reeflective/", configFileName)
	}

	return c.Save(path)
}

// SaveDefault writes the default configuration file to the specified path.
// If the path is empty, the file will save it to:
//
// - If $XDG_CONFIG_HOME is defined, saves to $XDG_CONFIG_HOME/reeflective/readline.yml
// - Else, saves to $HOME/.readline.yml
//
// This function is useful if you tried to load the system configuration, but
// that you could not find any. This thus gives the user a new default config.
func (c *config) SaveDefault(path string) (err error) {
	// Use the default directory if the path is empty
	if path == "" {
		// In home by default.
		userPath := os.Getenv("HOME")
		path = filepath.Join(userPath, "."+configFileName)

		// Or in configuration directory
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			path = filepath.Join(xdgConfigHome, "reeflective/", configFileName)
		}
	}

	// Create the complete path (including file) if it does not exist
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return
		}
	}

	// If the path is a directory, append the default filename to it, and create it.
	if file, ferr := os.Stat(path); ferr == nil && file.IsDir() {
		path = filepath.Join(path, configFileName)
	}

	return os.WriteFile(path, []byte(defaultConfig), os.FileMode(0o644))
}

// Export exports the current shell configuration as a byte array.
// Note that currently comments are not preserved in the output, so
// it is advised to first save a copy with SaveDefault() if you want
// to have indications on settings, and then load => merge/overwrite it.
func (c *config) Export() (config []byte, err error) {
	// The node marshals the configuration first, without comments.
	config, err = yaml.Marshal(c)
	if err != nil {
		return
	}

	// Now, unmarshal those values onto the node, which contains the comments.
	if err = yaml.Unmarshal(config, c.node); err != nil {
		return
	}

	return yaml.Marshal(c.node)
}

// loadDefaultConfig loads all default keymaps for input/shell control.
// This function is always executed, even before loading any custom config.
func (rl *Instance) loadDefaultConfig() {
	config := &config{
		rl:   rl,
		node: &yaml.Node{},

		// Input settings
		InputMode: Emacs,
		Vim: struct {
			// Cursors
			InsertCursor          CursorStyle `yaml:"insertCursor"`
			NormalCursor          CursorStyle `yaml:"normalCursor"`
			OperatorPendingCursor CursorStyle `yaml:"operatorPendingCursor"`
			VisualCursor          CursorStyle `yaml:"visualCursor"`
			ReplaceCursor         CursorStyle `yaml:"replaceCursor"`
		}{
			InsertCursor:          CursorBlinkingBeam,
			NormalCursor:          CursorBlinkingBlock,
			OperatorPendingCursor: CursorBlinkingUnderline,
			VisualCursor:          CursorBlock,
			ReplaceCursor:         CursorBlinkingUnderline,
		},
		Emacs: struct {
			Cursor CursorStyle
		}{
			Cursor: CursorBlinkingBlock,
		},
		Keymaps: make(map[keymapMode]keymap),

		// Prompt
		PromptTransient: false,

		// History
		HistoryAutoWrite:   true,
		HistoryAutosuggest: false,

		// Completions
		MaxTabCompleterRows: 50,
		AutoComplete:        true,
	}

	// Load keymaps, which are defaults themselves:
	// This overwrites the default bindkeys in the reeflectiverc file,
	// but since they should be the same, we just overwrite with this.
	config.Keymaps[emacs] = emacsKeys
	config.Keymaps[viins] = viinsKeys
	config.Keymaps[vicmd] = vicmdKeys
	config.Keymaps[viopp] = vioppKeys
	config.Keymaps[visual] = visualKeys

	config.Keymaps[menuselect] = menuselectKeys
	config.Keymaps[isearch] = menuselectKeys

	// First load the default configuration to preserve
	// the comment, and then apply our default values onto it.
	yaml.Unmarshal([]byte(defaultConfig), config.node)
	config.node.Decode(config)

	// Override the main input mode based on $EDITOR
	if hasVi, _ := regexp.MatchString("vi", os.Getenv("EDITOR")); hasVi {
		config.InputMode = Vim
	}

	rl.config = config
}

// watchConfig watches for changes in the config file,
// and reloads it each time a change is notified.
func (rl *Instance) watchConfig(path string) {
	watcher, werr := fsnotify.NewWatcher()
	if werr != nil {
		rl.hint = []rune(fmt.Sprintf("Failed to create file watcher: %s", werr.Error()))
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op.Has(fsnotify.Write) {
					rl.reloadConfig(event)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				changeHint := fmt.Sprintf("Error reloading config: %s", err.Error())
				rl.hint = []rune(changeHint)
				rl.redisplay()
			}
		}
	}()

	err := watcher.Add(path)
	if err != nil {
		rl.hint = []rune(fmt.Sprintf("Failed to monitor config: %s", err.Error()))
	}
	<-done
}

func (rl *Instance) reloadConfig(event fsnotify.Event) {
	loadErr := rl.config.load(event.Name)

	if loadErr != nil {
		errStr := strings.ReplaceAll(loadErr.Error(), "\n", "")
		changeHint := fmt.Sprintf(seqFgRed+"Config reload error: %s", errStr)
		rl.hint = append([]rune{}, []rune(changeHint)...)
	} else {
		changeHint := fmt.Sprintf(seqFgGreen+"Config reloaded: %s", event.Name)
		rl.hint = append([]rune{}, []rune(changeHint)...)
	}

	// Since the shell is currently waiting for input, it's going to hijack
	// any cursor position query result in completions, disable it temporarily.
	enabled := rl.EnableGetCursorPos
	rl.EnableGetCursorPos = false
	rl.renderHelpers()
	rl.EnableGetCursorPos = enabled
}
