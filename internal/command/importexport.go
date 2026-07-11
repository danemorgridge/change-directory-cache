package command

import (
	"fmt"
	"os"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
)

// Export writes the current cache as JSON. With a file argument it writes to
// that file; with no argument (or "-") it writes to stdout so it can be piped.
func Export(args []string) error {
	c, err := cache.Load()
	if err != nil {
		return err
	}
	data, err := c.Marshal()
	if err != nil {
		return err
	}

	dest := ""
	if len(args) > 0 {
		dest = args[0]
	}
	if dest == "" || dest == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}

	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Exported %d %s to %s\n", len(c.Entries), plural(len(c.Entries), "entry", "entries"), dest)
	return nil
}

// Import merges the entries from a JSON file into the cache. Entries whose
// directory no longer exists are reported and skipped; entries that duplicate a
// path already in the cache (or already imported) are reported and ignored.
func Import(args []string) error {
	if len(args) == 0 || args[0] == "" {
		return fmt.Errorf("import: a file to import is required")
	}
	src := args[0]

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}
	entries, err := cache.ParseEntries(data)
	if err != nil {
		return fmt.Errorf("import: %s does not contain valid cache JSON: %w", src, err)
	}

	c, err := cache.Load()
	if err != nil {
		return err
	}

	var added, missing, duplicate int
	for _, e := range entries {
		path, err := cache.NormalizePath(e.Path)
		if err != nil {
			path = e.Path // fall back to the raw value so the report is still useful
		}

		if !dirExists(path) {
			fmt.Fprintf(os.Stderr, "Skipped %q: directory does not exist: %s\n", e.Name, path)
			missing++
			continue
		}
		if existing, _ := c.FindByPath(path); existing != nil {
			fmt.Fprintf(os.Stderr, "Skipped duplicate %q -> %s (already saved as %q)\n", e.Name, path, existing.Name)
			duplicate++
			continue
		}

		c.Add(e.Name, path)
		added++
	}

	if added > 0 {
		if err := c.Save(); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "Imported %d, skipped %d missing, ignored %d duplicate.\n", added, missing, duplicate)
	return nil
}

// dirExists reports whether path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
