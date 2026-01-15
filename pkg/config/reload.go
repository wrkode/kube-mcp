package config

import (
	"os"
	"os/signal"
	"syscall"
)

// ReloadCallback is a function that gets called when configuration is reloaded.
// It receives the new configuration and should apply runtime-reloadable settings.
type ReloadCallback func(*Config) error

// SetupReload sets up SIGHUP signal handling for hot reload.
// On Windows, this is a no-op.
// The callback function is called with the new configuration when reload occurs.
func SetupReload(loader *Loader, callback ReloadCallback) error {
	// Check if we're on Windows
	if isWindows() {
		return nil
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for range sigChan {
			if err := loader.Reload(); err != nil {
				// Log error (we'll need a logger here, but for now just continue)
				_ = err
				continue
			}

			// Call callback to apply reloadable settings
			if callback != nil {
				cfg := loader.Get()
				if err := callback(cfg); err != nil {
					// Log error
					_ = err
				}
			}
		}
	}()

	return nil
}

// isWindows checks if the current OS is Windows.
func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
