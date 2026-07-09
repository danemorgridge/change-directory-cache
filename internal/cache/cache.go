// Package cache handles persistence of the directory bookmarks as JSON in the
// user's home directory (~/.directory-cache.json).
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is a single saved directory.
type Entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Cache is the on-disk collection of entries.
type Cache struct {
	Entries []Entry `json:"entries"`

	path string // location of the backing file; not serialized
}

// FilePath returns the location of the cache file.
func FilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".directory-cache.json"), nil
}

// Load reads the cache from disk. A missing or empty file yields an empty cache.
func Load() (*Cache, error) {
	fp, err := FilePath()
	if err != nil {
		return nil, err
	}
	c := &Cache{path: fp}
	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return c, nil
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save writes the cache back to disk.
func (c *Cache) Save() error {
	if c.Entries == nil {
		c.Entries = []Entry{}
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(c.path, data, 0o644)
}

// NormalizePath resolves p to a cleaned absolute path.
func NormalizePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

// FindByPath returns the entry (and its index) whose path matches, comparing
// case-insensitively. Returns nil, -1 when not found.
func (c *Cache) FindByPath(path string) (*Entry, int) {
	for i := range c.Entries {
		if strings.EqualFold(c.Entries[i].Path, path) {
			return &c.Entries[i], i
		}
	}
	return nil, -1
}

// Add appends a new entry.
func (c *Cache) Add(name, path string) {
	c.Entries = append(c.Entries, Entry{Name: name, Path: path})
}

// RemoveAt deletes the entry at index i.
func (c *Cache) RemoveAt(i int) {
	c.Entries = append(c.Entries[:i], c.Entries[i+1:]...)
}

// Match returns all entries whose name or path contains query, case-insensitively.
func (c *Cache) Match(query string) []Entry {
	q := strings.ToLower(query)
	var out []Entry
	for _, e := range c.Entries {
		if strings.Contains(strings.ToLower(e.Name), q) ||
			strings.Contains(strings.ToLower(e.Path), q) {
			out = append(out, e)
		}
	}
	return out
}

// SortByName sorts entries by name, case-insensitively.
func (c *Cache) SortByName() {
	sort.SliceStable(c.Entries, func(i, j int) bool {
		return strings.ToLower(c.Entries[i].Name) < strings.ToLower(c.Entries[j].Name)
	})
}

// SortByPath sorts entries by full path, case-insensitively.
func (c *Cache) SortByPath() {
	sort.SliceStable(c.Entries, func(i, j int) bool {
		return strings.ToLower(c.Entries[i].Path) < strings.ToLower(c.Entries[j].Path)
	})
}
