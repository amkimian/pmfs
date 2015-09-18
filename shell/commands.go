package shell

import (
	"fmt"
	"strings"
)

var parserCommands = map[string]ParserCommand{
	"cd":         ParserCommand{1, executeCD},
	"add":        ParserCommand{1, executeAddFile},
	"ls":         ParserCommand{1, executeLS},
	"cat":        ParserCommand{1, executeCat},
	"append":     ParserCommand{1, executeAppend},
	"appendLine": ParserCommand{1, executeAppendLine},
	"stat":       ParserCommand{1, executeStat},
}

func executeCD(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	//fmt.Printf("Running CD with parameters %v, remainingCommand %s", parameters, remainingCommand)
	executor.Cwd = parameters[0]
	ret := make([]string, 1)
	ret[0] = "changed directory"
	return ret
}

func executeAddFile(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	// TODO we need a standard way of resolving a file/path taking into account the current working directory

	executor.rfs.WriteFile(parameters[0], []byte(remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Created file %s", parameters[0])
	return ret
}

func executeAppend(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	executor.rfs.AppendFile(parameters[0], []byte(remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Appended to file %s", parameters[0])
	return ret
}

func executeAppendLine(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	executor.rfs.AppendFile(parameters[0], []byte("\n"+remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Appended with cr to file %s", parameters[0])
	return ret
}

func executeLS(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	// Again, morph parameters[0] to a full path

	ret, _ := executor.rfs.ListDirectory(parameters[0])
	return ret
}

func executeCat(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	arr, _ := executor.rfs.ReadFile(parameters[0])
	// Need to convert it into a string, then split on \n
	return strings.Split(string(arr), "\n")
}

func executeStat(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	stat, _ := executor.rfs.StatFile(parameters[0])
	fullString := fmt.Sprintf("Size : %d\nAccessed : %v\nCreated  : %v\nModified : %v\n", stat.Size, stat.Accessed, stat.Created, stat.Modified)
	return strings.Split(fullString, "\n")
}
