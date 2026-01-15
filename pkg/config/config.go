// Package config provides configuration loading and management for kube-mcp.
//
// The configuration system supports:
//   - TOML format configuration files
//   - Base configuration file + drop-in directory (conf.d/*.toml)
//   - Deep merging of configurations
//   - Hot reload on SIGHUP (not available on Windows)
//   - Default value application
package config
