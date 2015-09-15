package memory

import (
	"fmt"

	"github.com/amkimian/pmfs/fs"
)

type MemoryFileSystem struct {
	Blocks          map[fs.BlockNode][]byte
	UnusedNodeStart fs.BlockNode
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

func (mfs *MemoryFileSystem) GetRawBlock(node fs.BlockNode) []byte {
	return mfs.Blocks[node]
}

func (mfs *MemoryFileSystem) SaveRawBlock(node fs.BlockNode, data []byte) fs.BlockNode {
	if node == -1 {
		node = mfs.UnusedNodeStart
		mfs.UnusedNodeStart = mfs.UnusedNodeStart + 1
	}
	mfs.Blocks[node] = data
	return node
}
