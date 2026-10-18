package config

import (
	"fmt"
	"runtime"
)

// Build information. Populated at build-time.
var (
	// Version is the current version of the application
	Version = "dev"
	// CommitHash is the git commit hash
	CommitHash = "unknown"
	// BuildTime is the build timestamp
	BuildTime = "unknown"
	// GoVersion is the Go version used to build the application
	GoVersion = runtime.Version()
)

// GetVersionInfo returns a map with all version information
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":     Version,
		"commit_hash": CommitHash,
		"build_time":  BuildTime,
		"go_version":  GoVersion,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return fmt.Sprintf("Docffy v%s (%s)", Version, CommitHash)
}

// GetFullVersionInfo returns detailed version information
func GetFullVersionInfo() string {
	return fmt.Sprintf(`Docffy
Version: %s
Commit: %s
Build Time: %s
Go Version: %s
`, Version, CommitHash, BuildTime, GoVersion)
}
