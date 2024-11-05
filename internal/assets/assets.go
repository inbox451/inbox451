package assets

import (
	"fmt"

	"github.com/knadh/stuffbin"
)

// FS is the global stuffbin filesystem
var FS stuffbin.FileSystem

// InitAssets initializes the stuffbin filesystem
func InitAssets(execPath string) error {
	// Try loading embedded assets first
	fs, err := stuffbin.UnStuff(execPath)
	if err != nil {
		// Development mode: fall back to local filesystem
		fs, err = stuffbin.NewLocalFS("/")
		if err != nil {
			return fmt.Errorf("failed to initialize local file system: %w", err)
		}
	}

	FS = fs
	return nil
}
