// Package command implements the cdc subcommands.
package command

import (
	"fmt"
	"strings"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
	"github.com/danemorgridge/change-directory-cache/internal/picker"
	"github.com/danemorgridge/change-directory-cache/internal/shell"
)

// Open matches the query against cached names and paths and opens a new shell in
// the matching directory. A single match opens directly; several open the
// interactive picker first.
func Open(args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("open: a search query is required")
	}

	c, err := cache.Load()
	if err != nil {
		return err
	}

	matches := c.Match(query)
	switch len(matches) {
	case 0:
		return fmt.Errorf("open: no cached directory matches %q", query)
	case 1:
		return shell.Spawn(matches[0].Path)
	default:
		chosen, ok, err := picker.Pick(matches)
		if err != nil {
			return err
		}
		if !ok {
			return nil // user canceled
		}
		return shell.Spawn(chosen.Path)
	}
}
