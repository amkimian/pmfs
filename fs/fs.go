package fs

import (
	"bytes"
	"encoding/gob"
	"strings"
)

// A general filesystem

type RootFileSystem struct {
	BlockHandler  BlockHandler
	Configuration string
	SuperBlock    SuperBlockNode
}

type BlockNode int

type DirectoryNode struct {
	Folders      map[string]BlockNode
	Files        map[string]BlockNode
	Continuation BlockNode
}

type SuperBlockNode struct {
	BlockCount    int
	BlockSize     int
	RootDirectory BlockNode
}

type FileNode struct {
	Blocks       []BlockNode
	Continuation BlockNode
}

type BlockHandler interface {
	Init(configuration string)
	Format(blockCount int, blockSize int)
	GetRawBlock(node BlockNode) []byte
	SaveRawBlock(node BlockNode, data []byte) BlockNode
}

func (rfs *RootFileSystem) Init(handler BlockHandler, configuration string) {
	rfs.BlockHandler = handler
	rfs.Configuration = configuration
	rfs.BlockHandler.Init(configuration)
}

func (rfs *RootFileSystem) Format(bc int, bs int) {
	rfs.BlockHandler.Format(bc, bs)
	// Write Raw Directory node
	rdn := DirectoryNode{make(map[string]BlockNode), make(map[string]BlockNode), 0}
	Root := rfs.BlockHandler.SaveRawBlock(-1, rawBlock(rdn))
	sb := SuperBlockNode{bc, bs, Root}
	rfs.BlockHandler.SaveRawBlock(0, rawBlock(sb))
	rfs.SuperBlock = sb
}

// Returns the BlockNode and whether it is a directory or not
func (dn *DirectoryNode) findNode(paths []string, handler BlockHandler, createFileNode bool) FileNode {
	if len(paths) == 1 {
		// This should be looking in the Files section and create if not exist (depending on createFileNode)
		nodeId, ok := dn.Files[paths[0]]
		var fileNode FileNode
		if !ok {
			// Create FileNode, update the Files section, write this DirectoryNode back
			// Write FileNode back
			// Return that fileNode (blank really at the moment)
		} else {
			fileNode = getFileNode(handler.GetRawBlock(nodeId))
		}
		return fileNode
	} else {
		// This should look in the directories section and create a new directory node if that does not exist (depending on createFileNode)
		// Then recurse with a subset of the paths
		newDnId, ok := dn.Folders[paths[0]]
		var newDn DirectoryNode
		if !ok {
			// Create new DirectoryNode, write that, update this DirectoryNode, write that
		} else {
			newDn = getDirectoryNode(handler.GetRawBlock(newDnId))
		}

		return newDn.findNode(paths[1:], handler, createFileNode)
	}
}

func (rfs *RootFileSystem) WriteFile(fileName string, contents []byte) {
	// Find record for this fileName from RootFileSystem
	// After splitting on /
	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn := dn.findNode(parts, rfs.BlockHandler, true)
	// Create block nodes if non-existent
	// Eventually get to the last one, and write the bytes to a new (or existing one, overwriting)
	// there. (suitably slicing based on BlockSize)

}

func (rfs *RootFileSystem) ReadFile(fileName string) []byte {
	// Traverse the directory node system to find the BlockNode for the FileNode
	// Load that up, and read from the Blocks, appending to a single bytebuffer and then return that
	// If the ContinuationNode is set, load that one and carry on there
	return nil
}

func rawBlock(sb interface{}) []byte {
	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.
	err := enc.Encode(sb)
	if err == nil {
		return buffer.Bytes()
	}
	return nil
}

func getDirectoryNode(contents []byte) DirectoryNode {
	buffer := bytes.NewBuffer(contents)

	dec := gob.NewDecoder(buffer)
	var ret DirectoryNode
	dec.Decode(&ret)
	return ret
}

func getFileNode(contents []byte) FileNode {
	buffer := bytes.NewBuffer(contents)

	dec := gob.NewDecoder(buffer)
	var ret FileNode
	dec.Decode(&ret)
	return ret
}
