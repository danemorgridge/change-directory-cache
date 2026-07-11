package command

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
)

// useTempHome points cache.Load/Save at a throwaway home directory so tests
// never touch the real ~/.directory-cache.json.
func useTempHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	// os.UserHomeDir consults HOME on unix and USERPROFILE on Windows.
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}
}

func writeImportFile(t *testing.T, entries []cache.Entry) string {
	t.Helper()
	data, err := (&cache.Cache{Entries: entries}).Marshal()
	if err != nil {
		t.Fatalf("marshaling import file: %v", err)
	}
	fp := filepath.Join(t.TempDir(), "import.json")
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		t.Fatalf("writing import file: %v", err)
	}
	return fp
}

func TestImportSkipsMissingAndDuplicates(t *testing.T) {
	useTempHome(t)

	// Two real directories to import, plus one bogus path.
	realA := t.TempDir()
	realB := t.TempDir()
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	// Pre-seed the cache with realA so the import sees it as a duplicate.
	c, err := cache.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	pathA, _ := cache.NormalizePath(realA)
	c.Add("existing-a", pathA)
	if err := c.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	fp := writeImportFile(t, []cache.Entry{
		{Name: "dup-a", Path: realA},       // duplicate of existing-a -> ignored
		{Name: "new-b", Path: realB},       // added
		{Name: "gone", Path: missing},      // missing dir -> skipped
		{Name: "new-b-again", Path: realB}, // duplicate within the file -> ignored
	})

	if err := Import([]string{fp}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	got, err := cache.Load()
	if err != nil {
		t.Fatalf("reloading cache: %v", err)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("expected 2 entries after import, got %d: %v", len(got.Entries), got.Entries)
	}
	// existing-a stays; new-b was added exactly once.
	if e, _ := got.FindByPath(pathA); e == nil || e.Name != "existing-a" {
		t.Fatalf("existing entry was altered: %v", e)
	}
	pathB, _ := cache.NormalizePath(realB)
	if e, _ := got.FindByPath(pathB); e == nil || e.Name != "new-b" {
		t.Fatalf("expected new-b to be imported, got %v", e)
	}
	// The missing directory must not be added.
	if e, _ := got.FindByPath(missing); e != nil {
		t.Fatalf("missing directory should not be imported: %v", e)
	}
}

func TestImportRequiresFileArg(t *testing.T) {
	useTempHome(t)
	if err := Import(nil); err == nil {
		t.Fatal("expected an error when no file is given")
	}
}

func TestImportRejectsInvalidJSON(t *testing.T) {
	useTempHome(t)
	fp := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(fp, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Import([]string{fp}); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestExportToFileRoundTripsThroughImport(t *testing.T) {
	useTempHome(t)

	real := t.TempDir()
	c, err := cache.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p, _ := cache.NormalizePath(real)
	c.Add("proj", p)
	if err := c.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	out := filepath.Join(t.TempDir(), "export.json")
	if err := Export([]string{out}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	entries, err := cache.ParseEntries(mustRead(t, out))
	if err != nil {
		t.Fatalf("ParseEntries: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "proj" {
		t.Fatalf("unexpected exported entries: %v", entries)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return data
}
