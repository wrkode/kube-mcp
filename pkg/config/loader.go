package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Loader handles loading and reloading configuration.
type Loader struct {
	basePath    string
	confDPath   string
	config      *Config
	reloadFuncs []func(*Config) error
}

// NewLoader creates a new configuration loader.
func NewLoader(basePath, confDPath string) *Loader {
	return &Loader{
		basePath:    basePath,
		confDPath:   confDPath,
		reloadFuncs: make([]func(*Config) error, 0),
	}
}

// Load loads the configuration from the base file and drop-in directory.
func (l *Loader) Load() (*Config, error) {
	config := &Config{}

	// Load base configuration file if it exists
	if l.basePath != "" {
		if err := l.loadFile(l.basePath, config); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
	}

	// Load drop-in configuration files
	if l.confDPath != "" {
		entries, err := os.ReadDir(l.confDPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read conf.d directory: %w", err)
		}

		// Sort entries to ensure consistent loading order
		dropInFiles := make([]string, 0)
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toml") {
				dropInFiles = append(dropInFiles, filepath.Join(l.confDPath, entry.Name()))
			}
		}

		// Load each drop-in file and merge
		for _, filePath := range dropInFiles {
			if err := l.loadFile(filePath, config); err != nil {
				return nil, fmt.Errorf("failed to load drop-in config %s: %w", filePath, err)
			}
		}
	}

	// Apply defaults
	l.applyDefaults(config)

	l.config = config
	return config, nil
}

// loadFile loads a TOML file and merges it into the config.
func (l *Loader) loadFile(path string, config *Config) error {
	// Expand user home directory
	expandedPath := expandPath(path)

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return err
	}

	var fileConfig Config
	if err := toml.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Deep merge into existing config
	mergeConfig(config, &fileConfig)

	return nil
}

// Reload reloads the configuration and calls all registered reload functions.
func (l *Loader) Reload() error {
	newConfig, err := l.Load()
	if err != nil {
		return err
	}

	// Call all registered reload functions
	for _, fn := range l.reloadFuncs {
		if err := fn(newConfig); err != nil {
			return fmt.Errorf("reload function failed: %w", err)
		}
	}

	l.config = newConfig
	return nil
}

// OnReload registers a function to be called when configuration is reloaded.
func (l *Loader) OnReload(fn func(*Config) error) {
	l.reloadFuncs = append(l.reloadFuncs, fn)
}

// Get returns the current configuration.
func (l *Loader) Get() *Config {
	return l.config
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
