package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var params []string
	var i int
	for i = range args {
		arg := args[i]
		if len(arg) == 0 || arg[0] != '-' {
			break
		}
		if arg == "--" {
			i++
			break
		}

		params = append(params, arg)
	}

	var line []string
	for _, arg := range args[i:] {
		line = append(line, fmt.Sprintf("'%s'", strings.NewReplacer("'", "\\'", "\\", "\\\\").Replace(arg)))
	}

	newargs := []string{"env", "xargs"}
	newargs = append(newargs, params...)
	newargs = append(newargs, "sh", "-c", fmt.Sprintf("</dev/tty %s $@", strings.Join(line, " ")), "sh")

	return unix.Exec("/usr/bin/env", newargs, os.Environ())
}
