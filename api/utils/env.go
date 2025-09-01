package utils

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var (
	loadOnce   sync.Once
	loadErr    error
	loadedPath string
)

// LoadRootDotEnv searches upward from the current working directory to find
// the repository root .env file and loads it into the process environment.
// It is safe to call multiple times; the file will be loaded only once.
func LoadRootDotEnv() (string, error) {
	loadOnce.Do(func() {
		cwd, err := os.Getwd()
		if err != nil {
			loadErr = err
			return
		}

		// Walk up directories looking for a .env file
		dir := cwd
		for {
			candidate := filepath.Join(dir, ".env")
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				if err := godotenv.Load(candidate); err != nil {
					loadErr = err
					return
				}
				loadedPath = candidate
				return
			}

			parent := filepath.Dir(dir)
			if parent == dir { // reached filesystem root
				break
			}
			dir = parent
		}
		loadErr = errors.New(".env file not found when searching upward from current directory")
	})

	return loadedPath, loadErr
}
