package fs

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strings"
)

// A general filesystem

type BlockNodeType int

const (
	SUPERBLOCK BlockNodeType = iota
	DIRECTORY
	FILE
	DATA
	NIL
)

type RootFileSystem struct {
	BlockHandler  BlockHandler
	Configuration string
	SuperBlock    SuperBlockNode
}

type BlockNode struct {
	Type BlockNodeType
	Id   int
}

var NilBlock BlockNode = BlockNode{NIL, -1}
var SuperBlock BlockNode = BlockNode{SUPERBLOCK, 0}

type DirectoryNode struct {
	Node         BlockNode
	Folders      map[string]BlockNode
	Files        map[string]BlockNode
	Continuation BlockNode
}

type SuperBlockNode struct {
	Node          BlockNode
	BlockCount    int
	BlockSize     int
	RootDirectory BlockNode
}

type FileNode struct {
	Node         BlockNode
	Blocks       []BlockNode
	Continuation BlockNode
}

type BlockHandler interface {
	Init(configuration string)
	Format(blockCount int, blockSize int)
	GetFreeBlockNode(NodeType BlockNodeType) BlockNode
	GetRawBlock(node BlockNode) []byte
	SaveRawBlock(node BlockNode, data []byte) BlockNode
	FreeBlocks(blocks []BlockNode)
}

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

// Returns the BlockNode and whether it is a directory or not
func (dn *DirectoryNode) findNode(paths []string, handler BlockHandler, createFileNode bool) (*FileNode, error) {
	if len(paths) == 1 {
		// This should be looking in the Files section and create if not exist (depending on createFileNode)
		nodeId, ok := dn.Files[paths[0]]
		var fileNode *FileNode
		if !ok {
			if createFileNode {
				nodeId = handler.GetFreeBlockNode(FILE)
				fileNode = &FileNode{Node: nodeId, Blocks: make([]BlockNode, 20), Continuation: NilBlock}
				handler.SaveRawBlock(nodeId, rawBlock(fileNode))
				dn.Files[paths[0]] = nodeId
				handler.SaveRawBlock(dn.Node, rawBlock(dn))

				return fileNode, nil
			} else {
				return nil, errors.New("File not found")
			}
		} else {
			fileNode = getFileNode(handler.GetRawBlock(nodeId))
		}
		return fileNode, nil
	} else {
		// This should look in the directories section and create a new directory node if that does not exist (depending on createFileNode)
		// Then recurse with a subset of the paths
		newDnId, ok := dn.Folders[paths[0]]
		var newDn *DirectoryNode
		if !ok {
			if createFileNode {
				// Create new DirectoryNode, write that, update this DirectoryNode, write that
				newDnId = handler.GetFreeBlockNode(DIRECTORY)
				newDn = &DirectoryNode{Node: newDnId, Folders: make(map[string]BlockNode), Files: make(map[string]BlockNode), Continuation: NilBlock}
				handler.SaveRawBlock(newDnId, rawBlock(newDn))
				dn.Folders[paths[0]] = newDnId
				handler.SaveRawBlock(dn.Node, rawBlock(dn))

			} else {
				return nil, errors.New("Directory not found")
			}
		} else {
			newDn = getDirectoryNode(handler.GetRawBlock(newDnId))
		}

		return newDn.findNode(paths[1:], handler, createFileNode)
	}
}

func (rfs *RootFileSystem) WriteFile(fileName string, contents []byte) error {
	// Find record for this fileName from RootFileSystem
	// After splitting on /
	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn, err := dn.findNode(parts, rfs.BlockHandler, true)
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
	fn, err := dn.findNode(parts, rfs.BlockHandler, false)

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

func rawBlock(sb interface{}) []byte {
	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.
	err := enc.Encode(sb)
	if err == nil {
		return buffer.Bytes()
	}
	return nil
}

func getDirectoryNode(contents []byte) *DirectoryNode {
	buffer := bytes.NewBuffer(contents)

	dec := gob.NewDecoder(buffer)
	var ret DirectoryNode
	dec.Decode(&ret)
	return &ret
}

func getFileNode(contents []byte) *FileNode {
	buffer := bytes.NewBuffer(contents)

	dec := gob.NewDecoder(buffer)
	var ret FileNode
	dec.Decode(&ret)
	return &ret
}
