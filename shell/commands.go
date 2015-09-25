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
	"rm":         ParserCommand{1, executeRm},
	"mv":         ParserCommand{2, executeMv},
	"tags":       ParserCommand{1, executeTags},
}

func executeTags(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	verList, err := executor.Rfs.GetTags(filePath)
	if err != nil {
		return makeError(err)
	}
	return verList
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
	ret[0] = fmt.Sprintf("Created file %s", filePath)
	return ret
}

func executeAppend(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	executor.Rfs.AppendFile(filePath, []byte(remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Appended to file %s", filePath)
	return ret
}

func executeAppendLine(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	executor.Rfs.AppendFile(filePath, []byte("\n"+remainingCommand))
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Appended with cr to file %s", filePath)
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
	fileNode, err := executor.Rfs.StatFile(filePath)
	if err == nil {
		fullString := fmt.Sprintf("Size : %d\nAccessed : %v\nCreated  : %v\nModified : %v\nBlocks: %v\nDefault Route: %v\n", fileNode.Stats.Size, fileNode.Stats.Accessed, fileNode.Stats.Created, fileNode.Stats.Modified, fileNode.DataBlocks, fileNode.DefaultRoute)
		return strings.Split(fullString, "\n")
	} else {
		return makeError(err)
	}
}

func makeError(err error) []string {
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("%v", err)
	return ret
}
func executeRm(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	filePath := util.ResolvePath(executor.Cwd, parameters[0])
	executor.Rfs.DeleteFile(filePath)
	ret := make([]string, 1)
	ret[0] = fmt.Sprintf("Removed %s", filePath)
	return ret
}

func executeMv(parameters []string, remainingCommand string, executor *ShellExecutor) []string {
	sourceFilePath := util.ResolvePath(executor.Cwd, parameters[0])
	targetFilePath := util.ResolvePath(executor.Cwd, parameters[1])
	ret := make([]string, 1)
	err := executor.Rfs.MoveFileOrFolder(sourceFilePath, targetFilePath)
	if err == nil {
		ret[0] = fmt.Sprintf("Moved %s to %s", sourceFilePath, targetFilePath)
	} else {
		ret[0] = fmt.Sprintf("Source path %s, Error %v", sourceFilePath, err)
	}
	return ret
}
