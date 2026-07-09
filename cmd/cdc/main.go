// Command cdc (Change Directory Cache) bookmarks directories and jumps into them.
//
// Install:
//
//	go install github.com/danemorgridge/change-directory-cache/cmd/cdc@latest
//
// Then `cdc save <name>`, `cdc o <query>`, `cdc list`, `cdc rm`. `cdc o` opens a
// new shell in the matched directory (type exit to return) so no shell hook or
// PATH setup is required.
package main

import (
	"fmt"
	"os"

	"github.com/danemorgridge/change-directory-cache/internal/command"
)

func usage() {
	fmt.Fprint(os.Stderr, `cdc — change directory cache

Usage:
  cdc open|o <query>   open a shell in a cached directory matching <query>
  cdc save|s [name]    save the current directory (name defaults to the folder name)
  cdc rm               remove the current directory from the cache
  cdc list|ls [-p]     list cached directories, sorted by name (or by path with -p)
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	args := os.Args[2:]
	var err error

	switch os.Args[1] {
	case "open", "o":
		err = command.Open(args)
	case "save", "s":
		err = command.Save(args)
	case "rm":
		err = command.Rm(args)
	case "list", "ls":
		err = command.List(args)
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "cdc: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "cdc: "+err.Error())
		os.Exit(1)
	}
}
