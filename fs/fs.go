package fs

import (
	"bytes"
	"encoding/gob"
)

// A general filesystem

type RootFileSystem struct {
	BlockHandler  BlockHandler
	Configuration string
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
}

func (rfs *RootFileSystem) WriteFile(fileName string, contents []byte) {
	// Find record for this fileName from RootFileSystem
	// After splitting on /
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
