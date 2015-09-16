package memory

import (
	"fmt"

	"github.com/amkimian/pmfs/fs"
)

type MemoryFileSystem struct {
	Blocks          map[fs.BlockNode][]byte
	UnusedNodeStart int
}

func (mfs *MemoryFileSystem) Init(configuration string) {
	fmt.Println("Starting memory file system")
	mfs.UnusedNodeStart = 20000
}

func (mfs *MemoryFileSystem) Format(blockCount int, blockSize int) {
	// Initialize a new memory file system
	fmt.Println("Format memory filesystem")
	mfs.Blocks = make(map[fs.BlockNode][]byte)
}

func (mfs *MemoryFileSystem) GetFreeBlockNode(NodeType fs.BlockNodeType) fs.BlockNode {
	var node fs.BlockNode
	node.Type = NodeType
	node.Id = mfs.UnusedNodeStart
	mfs.UnusedNodeStart = mfs.UnusedNodeStart + 1
	return node
}

func (mfs *MemoryFileSystem) GetRawBlock(node fs.BlockNode) []byte {
	return mfs.Blocks[node]
}

func (mfs *MemoryFileSystem) SaveRawBlock(node fs.BlockNode, data []byte) fs.BlockNode {
	mfs.Blocks[node] = data
	return node
}

func (mfs *MemoryFileSystem) FreeBlocks(blocks []fs.BlockNode) {
	for i := range blocks {
		delete(mfs.Blocks, blocks[i])
	}
}
