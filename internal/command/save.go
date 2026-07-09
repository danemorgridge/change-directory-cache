package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/danemorgridge/change-directory-cache/internal/cache"
)

// Save stores the current directory under a name (all text after the verb,
// spaces allowed). With no name it defaults to the folder's basename. If the
// folder is already cached it reports the existing name and asks whether to
// update it.
func Save(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path, err := cache.NormalizePath(cwd)
	if err != nil {
		return err
	}

	rawName := strings.TrimSpace(strings.Join(args, " "))
	name := rawName
	if name == "" {
		name = filepath.Base(path)
	}

	c, err := cache.Load()
	if err != nil {
		return err
	}

	if existing, idx := c.FindByPath(path); existing != nil {
		fmt.Fprintf(os.Stderr, "This folder is already saved as %q.\n", existing.Name)
		update, err := promptYesNo("Update the name?")
		if err != nil {
			return err
		}
		if !update {
			fmt.Fprintln(os.Stderr, "Ignored; cache unchanged.")
			return nil
		}
		newName := rawName
		if newName == "" {
			newName, err = promptName(existing.Name)
			if err != nil {
				return err
			}
			if newName == "" {
				newName = existing.Name
			}
		}
		c.Entries[idx].Name = newName
		if err := c.Save(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Updated name to %q.\n", newName)
		return nil
	}

	c.Add(name, path)
	if err := c.Save(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Saved %q -> %s\n", name, path)
	return nil
}

func promptYesNo(question string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", question)
	line, err := readLine()
	if err != nil {
		return false, err
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes", nil
}

func promptName(current string) (string, error) {
	fmt.Fprintf(os.Stderr, "New name [%s]: ", current)
	line, err := readLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// stdin is a single shared reader; a fresh bufio.Reader per call would buffer
// and then discard input meant for the next prompt.
var stdin = bufio.NewReader(os.Stdin)

func readLine() (string, error) {
	line, err := stdin.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return line, nil
}
