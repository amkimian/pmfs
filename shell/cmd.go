package shell

import (
	"fmt"

	"github.com/amkimian/pmfs/fs"
)
import "github.com/amkimian/pmfs/memory"

type ParserExecFunction func(parameters []string, remainingCommand string, executor *ShellExecutor) []string

type ParserCommand struct {
	numberOfKnownParameters int
	runFn                   ParserExecFunction
}

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

func (cp *CommandParser) parse(full string, executor *ShellExecutor) []string {
	cp.fullCommand = full
	cp.commandToken, cp.remainingCommand = grabToken(full)
	parserCommand, ok := parserCommands[cp.commandToken]
	if ok {
		cp.parameters, cp.remainingCommand = grabTokens(cp.remainingCommand, parserCommand.numberOfKnownParameters)
		return parserCommand.runFn(cp.parameters, cp.remainingCommand, executor)
	} else {
		fmt.Printf("Command token is '%s'\n", cp.commandToken)
	}
	ret := make([]string, 1)
	ret[0] = "Don't know what to do"
	return ret
}

func grabTokens(line string, tokenCount int) ([]string, string) {
	//fmt.Println("Grabbing tokens")
	ret := make([]string, tokenCount)
	for i := 0; i < tokenCount && len(line) > 0; i++ {
		//fmt.Printf("Grabbing next token from '%s'\n", line)
		ret[i], line = grabToken(line)
	}
	return ret, line
}

// delimiter is space
func grabToken(line string) (string, string) {
	tokenChars := make([]byte, 0)
	var endPoint int
	var found = false
	var i = 0
	for i, found = 0, false; !found && (i < len(line)); i++ {
		if line[i] == ' ' {
			found = true
			endPoint = i
		} else {
			tokenChars = append(tokenChars, line[i])
		}
	}
	if !found {
		return string(tokenChars), ""
	} else {
		return string(tokenChars), line[endPoint+1:]
	}
}

func (se *ShellExecutor) Init() {
	var mh memory.MemoryFileSystem
	se.rfs.Init(&mh, "")
	se.rfs.Format(100, 100)
	se.Cwd = "/alan"
}

func (se *ShellExecutor) ExecuteLine(line string) []string {
	// The line will be something like
	// cd /
	// The first token will be a command, the rest of the tokens will depend
	// on the command. Each command will either alter the cwd, or execute something
	// against the filesystem, ultimately returning something that we then
	// convert to a series of strings that we can display
	cp := CommandParser{}
	ret := cp.parse(line, se)
	return ret
}
