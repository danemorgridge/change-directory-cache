package command

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
)

// List prints a table of cached directories. By default it sorts by name; the
// -p (or --path) flag sorts by full path instead.
func List(args []string) error {
	byPath := false
	for _, a := range args {
		if a == "-p" || a == "--path" {
			byPath = true
		}
	}

	c, err := cache.Load()
	if err != nil {
		return err
	}
	if len(c.Entries) == 0 {
		fmt.Fprintln(os.Stderr, "Cache is empty.")
		return nil
	}

	if byPath {
		c.SortByPath()
	} else {
		c.SortByName()
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tFOLDER")
	for _, e := range c.Entries {
		fmt.Fprintf(w, "%s\t%s\n", e.Name, e.Path)
	}
	return w.Flush()
}
