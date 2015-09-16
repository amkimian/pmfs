package fs

import (
	"bytes"
	"strings"
)

// A general filesystem

func (rfs *RootFileSystem) Init(handler BlockHandler, configuration string) {
	rfs.BlockHandler = handler
	rfs.Configuration = configuration
	rfs.BlockHandler.Init(configuration)
}

func (rfs *RootFileSystem) Format(bc int, bs int) {
	rfs.BlockHandler.Format(bc, bs)
	// Write Raw Directory node
	rdn := DirectoryNode{Folders: make(map[string]BlockNode), Files: make(map[string]BlockNode), Continuation: NilBlock}
	blockNode := rfs.BlockHandler.GetFreeBlockNode(DIRECTORY)
	rdn.Node = blockNode

	Root := rfs.BlockHandler.SaveRawBlock(blockNode, rawBlock(rdn))
	sb := SuperBlockNode{SuperBlock, bc, bs, Root}

	rfs.BlockHandler.SaveRawBlock(SuperBlock, rawBlock(sb))
	rfs.SuperBlock = sb
}

func (rfs *RootFileSystem) ListDirectory(path string) ([]string, error) {
	// As a test, simply return the names of things at this path. Later on we'll return a structure that defines names, types and stats

	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	var dnReal *DirectoryNode
	var err error

	if path == "/" {
		dnReal = dn
	} else {
		parts := strings.Split(path, "/")
		dnReal, err = dn.findDirectoryNode(parts[1:], rfs.BlockHandler)
		// And now get the names of things and add them to "entries"
		// for now, don't do the continuation
		if err != nil {
			return nil, err
		}
	}
	size := len(dnReal.Folders) + len(dnReal.Files)
	entries := make([]string, size)
	point := 0
	for name, _ := range dnReal.Folders {
		entries[point] = name
		point++
	}
	for name, _ := range dnReal.Files {
		entries[point] = name
		point++
	}
	return entries, nil
}

func (rfs *RootFileSystem) WriteFile(fileName string, contents []byte) error {
	// Find record for this fileName from RootFileSystem
	// After splitting on /
	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn, err := dn.findNode(parts[1:], rfs.BlockHandler, true)
	if err == nil {
		// Found the node, write data to node, overwriting any existing data (the node will be a filenode)
		rfs.BlockHandler.FreeBlocks(fn.Blocks)
		contin := fn.Continuation
		for contin != NilBlock {
			fileBytes := rfs.BlockHandler.GetRawBlock(contin)
			fileNode := getFileNode(fileBytes)
			rfs.BlockHandler.FreeBlocks(fileNode.Blocks)
			contin = fileNode.Continuation
		}

		// Loop through contents, writing out bytes in blocks of size rfs.SuperBlock.BlockSize
		// For each DataNode, add to the fn.Blocks
		// If fn.Blocks gets too large, create a new FileNode, put that as the Continuation, write this fileNode out
		// Move to that node

		for i := 0; i < len(contents); i = i + rfs.SuperBlock.BlockSize {
			var toWrite []byte
			if i+rfs.SuperBlock.BlockSize > len(contents) {
				toWrite = contents[i:]
			} else {
				toWrite = contents[i : i+rfs.SuperBlock.BlockSize]
			}
			if len(fn.Blocks) >= 20 { // Arbitary
				newContNode := rfs.BlockHandler.GetFreeBlockNode(FILE)
				newContFileNode := FileNode{Node: newContNode, Blocks: make([]BlockNode, 20), Continuation: NilBlock}
				fn.Continuation = newContNode
				rfs.BlockHandler.SaveRawBlock(fn.Node, rawBlock(fn))
				fn = &newContFileNode
			}
			newDataNode := rfs.BlockHandler.GetFreeBlockNode(DATA)
			fn.Blocks = append(fn.Blocks, newDataNode)
			rfs.BlockHandler.SaveRawBlock(newDataNode, toWrite)
		}

		rfs.BlockHandler.SaveRawBlock(fn.Node, rawBlock(fn))

		return nil
	} else {
		// Something went wrong, what to do? (probably propogate the error)
		return err
	}

}

func (rfs *RootFileSystem) ReadFile(fileName string) ([]byte, error) {
	// Traverse the directory node system to find the BlockNode for the FileNode
	// Load that up, and read from the Blocks, appending to a single bytebuffer and then return that
	// If the ContinuationNode is set, load that one and carry on there

	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn, err := dn.findNode(parts[1:], rfs.BlockHandler, false)

	if err == nil {
		buffer := new(bytes.Buffer)
		done := false
		for !done {
			for i := range fn.Blocks {
				data := rfs.BlockHandler.GetRawBlock(fn.Blocks[i])
				buffer.Write(data)
			}
			if fn.Continuation == NilBlock {
				done = true
			} else {
				fileBytes := rfs.BlockHandler.GetRawBlock(fn.Continuation)
				fn = getFileNode(fileBytes)
			}
		}
		return buffer.Bytes(), nil
	} else {
		return nil, err
	}
}
