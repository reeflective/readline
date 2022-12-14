package readline

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = ".reeflective.yml"
)

//go:embed reeflective.yml
var defaultConfig string

// ErrNoSystemConfig indicates that no readline configuration file
// could be found in  any of the default user system paths:
//
// $XDG_CONFIG_HOME/reeflective/.reeflective.yml
// $HOME/.reeflective.yml
//
var ErrNoSystemConfig = errors.New("no user configuration found in user directories")

// config stores all configurable elements for the shell, including the
// complete list of keymaps. The configuration is always written/exported
// as a YAML file, and any file to be imported as a configuration is also
// unmarshaled as YAML.
type config struct {
	node *yaml.Node // Stores the configuration file bytes including comments.

	//
	// Input modes and keymaps
	//

	// InputMode - The shell can be used in Vim editing mode, or Emacs (classic).
	InputMode InputMode `yaml:"inputMode"`

	Vim struct {
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

	//
	// Line settings
	//

	// The shell displays fish-like autosuggestions (the first
	// matching history line is displayed dimmed).
	HistoryAutosuggest bool `yaml:"historyAutosuggest"`
	// HistoryAutoWrite defines whether items automatically get written to history.
	// Enabled by default. Set to false to disable.
	HistoryAutoWrite bool `yaml:"historyAutoWrite"`

	//
	// Completion settings
	//

	MaxTabCompleterRows int `yaml:"maxTabCompleterRows"`

	//
	// Other helpers
	//

	// HintColor any ANSI escape codes you wish to use for hint formatting.
	HintFormatting string `yaml:"hintFormatting"`
}

// Config returns the current configuration of the readline instance.
func (rl *Instance) Config() *config {
	return rl.config
}

// Load looks for a configuration at the specified path (including file name).
// It returns an error either if the file is not found, if the shell fails to read it,
// or if its fails to unmarshal/load it.
func (c *config) Load(path string) (err error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// First load the yaml as node with comments
	if err = yaml.Unmarshal(bytes, c.node); err != nil {
		return
	}

	// And then unmarshal the node onto the struct.
	return c.node.Decode(c)
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
// $XDG_CONFIG_HOME/reeflective/.reeflective.yml
// $HOME/.reeflective.yml
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
	configHomePath := filepath.Join(userPath, configFileName)
	if _, err := os.Stat(configHomePath); err == nil {
		return c.Load(configHomePath)
	}

	return ErrNoSystemConfig
}

// Save saves the current shell configuration to the specified path.
// If the path is a directory, the configuration file is saved as '$path/.reeflective.yml'.
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
		path = filepath.Join(path, configFileName)
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
// - If $XDG_CONFIG_HOME is defined, saves to $XDG_CONFIG_HOME/reeflective/.reeflective.yml
// - Else, saves to $HOME/.reeflective.yml
//
func (c *config) SaveSystem() (err error) {
	var path string

	// In home by default.
	userPath := os.Getenv("HOME")
	path = filepath.Join(userPath, configFileName)

	// Or in configuration directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		path = filepath.Join(xdgConfigHome, "reeflective/", configFileName)
	}

	return c.Save(path)
}

// SaveDefault writes the default configuration file to the specified path.
// This function is mostly useful if you want to manually edit the configuration
// and that you need comments to explain the various fields: the other Save() methods
// will not preserve comments when writing the configuration.
func (c *config) SaveDefault(path string) (err error) {
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

		// History
		HistoryAutoWrite: true,

		// Completions
		MaxTabCompleterRows: 50,

		// Other helpers
		HintFormatting: DIM,
	}

	// Load keymaps, which are defaults themselves:
	// This overwrites the default bindkeys in the reeflectiverc file,
	// but since they should be the same, we just overwrite with this.
	config.Keymaps[emacs] = emacsKeys
	config.Keymaps[viins] = viinsKeys
	config.Keymaps[vicmd] = vicmdKeys
	config.Keymaps[viopp] = vioppKeys
	config.Keymaps[visual] = visualKeys

	// First load the default configuration to preserve
	// the comment, and then apply our default values onto it.
	// yaml.Unmarshal([]byte(defaultConfig), config.node)
	// config.node.Decode(config)

	rl.config = config
}
