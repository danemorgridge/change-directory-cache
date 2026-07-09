package command

import (
	"fmt"
	"os"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
)

// Rm removes the current directory from the cache, erroring if it isn't cached.
func Rm(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path, err := cache.NormalizePath(cwd)
	if err != nil {
		return err
	}

	c, err := cache.Load()
	if err != nil {
		return err
	}

	existing, idx := c.FindByPath(path)
	if existing == nil {
		return fmt.Errorf("rm: current folder is not in the cache: %s", path)
	}
	name := existing.Name
	c.RemoveAt(idx)
	if err := c.Save(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Removed %q -> %s\n", name, path)
	return nil
}
