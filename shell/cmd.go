package shell

import (
	"fmt"

	"github.com/amkimian/pmfs/fs"
)
import "github.com/amkimian/pmfs/memory"

type ShellExecutor struct {
	rfs fs.RootFileSystem
	Cwd string
}

// A CommandParser takes a line and parses it into the name of the command (e.g. cd)
// and a list of known parameters (depends on command), and then the rest of the command remains
// e.g. cd always takes one parameter, so parameters would (should) be a 1 length array and remainingCommand blank
type CommandParser struct {
	fullCommand      string
	remainingCommand string
	commandToken     string
	parameters       []string
}

func (cp *CommandParser) parse(full string) {
	cp.fullCommand = full
	cp.commandToken, cp.remainingCommand = grabToken(full)
	// Now we switch on commandToken
	if cp.commandToken == "cd" {
		cp.parameters, cp.remainingCommand = grabTokens(cp.remainingCommand, 1)
	} else {
		fmt.Printf("Command token is '%s'\n", cp.commandToken)
	}
}

func grabTokens(line string, tokenCount int) ([]string, string) {
	fmt.Println("Grabbing tokens")
	ret := make([]string, tokenCount)
	for i := 0; i < tokenCount && len(line) > 0; i++ {
		fmt.Printf("Grabbing next token from '%s'\n", line)
		ret[i], line = grabToken(line)
	}
	return ret, line
}

// delimiter is space
func grabToken(line string) (string, string) {
	tokenChars := make([]byte, 0)
	var endPoint int
	for i, found := 0, false; !found && (i < len(line)); i++ {
		if line[i] == ' ' {
			found = true
			endPoint = i
		} else {
			fmt.Printf("Adding %c\n", line[i])
			tokenChars = append(tokenChars, line[i])
		}
	}
	return string(tokenChars), line[endPoint+1:]
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
