// Package shell launches an interactive shell in a chosen directory. This is how
// `cdc open` "moves" you into a bookmark without needing any shell hook: a child
// process can't change its parent's directory, so instead it starts a fresh
// shell already sitting in the target directory. Type `exit` to return.
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Spawn starts an interactive shell with its working directory set to dir and
// waits for it to exit. The shell's own exit status is ignored (e.g. a failed
// last command shouldn't look like a cdc failure).
func Spawn(dir string) error {
	name, args := shellCommand()
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	c.Env = append(os.Environ(), "CDC_ACTIVE=1")

	fmt.Fprintf(os.Stderr, "cdc → %s  (type 'exit' to return)\n", dir)
	if err := c.Start(); err != nil {
		return fmt.Errorf("could not start shell %q: %w", name, err)
	}
	_ = c.Wait()
	return nil
}

// shellCommand picks which shell to launch. CDC_SHELL overrides everything; on
// Windows we prefer %COMSPEC% (cmd.exe) so there's no PowerShell dependency,
// falling back to a Unix-style $SHELL if one is set (e.g. Git Bash).
func shellCommand() (string, []string) {
	if custom := os.Getenv("CDC_SHELL"); custom != "" {
		return custom, nil
	}

	if runtime.GOOS == "windows" {
		if comspec := os.Getenv("COMSPEC"); comspec != "" {
			return comspec, nil
		}
		if sh := os.Getenv("SHELL"); sh != "" {
			return sh, nil
		}
		return "cmd.exe", nil
	}

	if sh := os.Getenv("SHELL"); sh != "" {
		return sh, nil
	}
	return "/bin/sh", nil
}
