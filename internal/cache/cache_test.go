package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sample() *Cache {
	return &Cache{Entries: []Entry{
		{Name: "work", Path: "/home/dane/work"},
		{Name: "Docs", Path: "/home/dane/Documents"},
		{Name: "src", Path: "/home/dane/source"},
	}}
}

func TestFindByPath(t *testing.T) {
	c := sample()

	// Exact match.
	e, i := c.FindByPath("/home/dane/work")
	if e == nil || i != 0 || e.Name != "work" {
		t.Fatalf("exact match failed: entry=%v idx=%d", e, i)
	}

	// Case-insensitive match.
	e, i = c.FindByPath("/HOME/DANE/DOCUMENTS")
	if e == nil || i != 1 || e.Name != "Docs" {
		t.Fatalf("case-insensitive match failed: entry=%v idx=%d", e, i)
	}

	// No match.
	if e, i := c.FindByPath("/nope"); e != nil || i != -1 {
		t.Fatalf("expected no match, got entry=%v idx=%d", e, i)
	}
}

func TestAddAndRemoveAt(t *testing.T) {
	c := &Cache{}
	c.Add("a", "/a")
	c.Add("b", "/b")
	c.Add("c", "/c")
	if len(c.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(c.Entries))
	}

	c.RemoveAt(1) // remove "b"
	if len(c.Entries) != 2 {
		t.Fatalf("expected 2 entries after remove, got %d", len(c.Entries))
	}
	if c.Entries[0].Name != "a" || c.Entries[1].Name != "c" {
		t.Fatalf("unexpected entries after remove: %v", c.Entries)
	}
}

func TestMatch(t *testing.T) {
	c := sample()

	// Matches on name, case-insensitively.
	if got := c.Match("WORK"); len(got) != 1 || got[0].Name != "work" {
		t.Fatalf("name match failed: %v", got)
	}

	// Matches on path substring.
	if got := c.Match("dane"); len(got) != 3 {
		t.Fatalf("path substring match failed: expected 3, got %d", len(got))
	}

	// No match returns empty.
	if got := c.Match("zzz"); len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}
}

func TestSortByName(t *testing.T) {
	c := sample()
	c.SortByName()
	want := []string{"Docs", "src", "work"} // case-insensitive ordering
	for i, name := range want {
		if c.Entries[i].Name != name {
			t.Fatalf("SortByName[%d] = %q, want %q (order: %v)", i, c.Entries[i].Name, name, names(c))
		}
	}
}

func TestSortByPath(t *testing.T) {
	c := sample()
	c.SortByPath()
	want := []string{"Documents", "source", "work"}
	for i, base := range want {
		if filepath.Base(c.Entries[i].Path) != base {
			t.Fatalf("SortByPath[%d] base = %q, want %q", i, filepath.Base(c.Entries[i].Path), base)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	got, err := NormalizePath(".")
	if err != nil {
		t.Fatalf("NormalizePath error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("expected absolute path, got %q", got)
	}
	cwd, _ := os.Getwd()
	if got != filepath.Clean(cwd) {
		t.Fatalf("NormalizePath(\".\") = %q, want %q", got, filepath.Clean(cwd))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	fp := filepath.Join(t.TempDir(), "cache.json")

	c := &Cache{path: fp}
	c.Add("work", "/home/dane/work")
	c.Add("docs", "/home/dane/docs")
	if err := c.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Load reads from FilePath(), so read the file back directly to verify Save.
	loaded := &Cache{path: fp}
	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if err := json.Unmarshal(data, loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(loaded.Entries) != 2 || loaded.Entries[0].Name != "work" || loaded.Entries[1].Path != "/home/dane/docs" {
		t.Fatalf("round trip mismatch: %v", loaded.Entries)
	}
}

func TestSaveNilEntriesWritesEmptyArray(t *testing.T) {
	fp := filepath.Join(t.TempDir(), "cache.json")
	c := &Cache{path: fp} // Entries is nil
	if err := c.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if c.Entries == nil {
		t.Fatal("Save should initialize Entries to a non-nil slice")
	}
	if !strings.Contains(string(data), `"entries": []`) {
		t.Fatalf("expected empty entries array in output, got:\n%s", data)
	}
}

func TestParseEntriesToleratesBOM(t *testing.T) {
	body := []byte(`{"entries":[{"name":"a","path":"/a"}]}`)
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, body...)

	entries, err := ParseEntries(withBOM)
	if err != nil {
		t.Fatalf("ParseEntries with BOM: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "a" {
		t.Fatalf("unexpected entries: %v", entries)
	}

	// Empty input (and a lone BOM) yields no entries, no error.
	if entries, err := ParseEntries(nil); err != nil || entries != nil {
		t.Fatalf("empty input: entries=%v err=%v", entries, err)
	}
	if entries, err := ParseEntries([]byte{0xEF, 0xBB, 0xBF}); err != nil || entries != nil {
		t.Fatalf("lone BOM: entries=%v err=%v", entries, err)
	}
}

func names(c *Cache) []string {
	out := make([]string, len(c.Entries))
	for i, e := range c.Entries {
		out[i] = e.Name
	}
	return out
}
