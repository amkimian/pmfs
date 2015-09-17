package main

import "github.com/amkimian/pmfs/fs"
import "github.com/amkimian/pmfs/memory"

type ShellExecutor struct {
	rfs fs.RootFileSystem
	Cwd string
}

func (se *ShellExecutor) Init() {
	var mh memory.MemoryFileSystem
	se.rfs.Init(&mh, "")
	se.rfs.Format(100, 100)
	se.Cwd = "/"
}

func (se *ShellExecutor) ExecuteLine(line string) []string {
	// The line will be something like
	// cd /
	// The first token will be a command, the rest of the tokens will depend
	// on the command. Each command will either alter the cwd, or execute something
	// against the filesystem, ultimately returning something that we then
	// convert to a series of strings that we can display
	return make([]string, 0)
}
