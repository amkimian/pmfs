package shell

import (
	"fmt"
	"strings"

	"github.com/amkimian/pmfs/util"
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
	executor.Cwd = util.ResolvePath(executor.Cwd, parameters[0])
	ret := make([]string, 1)
	ret[0] = "changed directory"
	return ret
}

func executeAddFile(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	executor.Rfs.WriteFile(filePath, []byte(remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Created file %s", parameters[0])
	return ret
}

func executeAppend(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	executor.Rfs.AppendFile(filePath, []byte(remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Appended to file %s", parameters[0])
	return ret
}

func executeAppendLine(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	executor.Rfs.AppendFile(filePath, []byte("\n"+remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Appended with cr to file %s", parameters[0])
	return ret
}

func executeLS(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	ret, _ := executor.Rfs.ListDirectory(filePath)
	return ret
}

func executeCat(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	arr, _ := executor.Rfs.ReadFile(filePath)
	// Need to convert it into a string, then split on \n
	return strings.Split(string(arr), "\n")
}

func executeStat(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	fileNode, _ := executor.Rfs.StatFile(filePath)
	fullString := fmt.Sprintf("Size : %d\nAccessed : %v\nCreated  : %v\nModified : %v\n", fileNode.Stats.Size, fileNode.Stats.Accessed, fileNode.Stats.Created, fileNode.Stats.Modified)
	return strings.Split(fullString, "\n")
}
