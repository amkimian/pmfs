// In memory file system
//
// Example:
//  var f fs.RootFileSystem
//  var mh memory.MemoryFileSystem
//
// 	f.Init(&mh, "")
//	f.Format(100, 100)
//
//  f.WriteFile("/fred/alan", []byte("Hello world"))
package memory

import (
	"fmt"

	"github.com/amkimian/pmfs/fs"
)

// The MemoryFileSystem simply contains a map of blocks and an ever increasing next node id
type MemoryFileSystem struct {
	Blocks          map[fs.BlockNode][]byte
	UnusedNodeStart int
}

// Initialize the file system (does nothing for the memory filesystem)
func (mfs *MemoryFileSystem) Init(configuration string) {
}

// Format the file system - clearing the in memory map of blocks and reseting the node id
func (mfs *MemoryFileSystem) Format(blockCount int, blockSize int) {
	// Initialize a new memory file system
	//fmt.Println("Format memory filesystem")
	mfs.Blocks = make(map[fs.BlockNode][]byte)
	mfs.UnusedNodeStart = 20000
}

// Simply returns the next free node id
func (mfs *MemoryFileSystem) GetFreeBlockNode(NodeType fs.BlockNodeType) fs.BlockNode {
	var node fs.BlockNode
	node.Type = NodeType
	node.Id = mfs.UnusedNodeStart
	mfs.UnusedNodeStart = mfs.UnusedNodeStart + 1
	return node
}

func (mfs *MemoryFileSystem) GetFreeDataBlockNode(parent fs.BlockNode, key string) fs.BlockNode {
	var node fs.BlockNode
	node.Type = fs.DATA
	node.RelativeTo = parent.Id
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

func (mfs *MemoryFileSystem) DumpInfo() {
	fmt.Printf("Memory file system: next block id %v\n", mfs.UnusedNodeStart)
	fmt.Printf("Total blocks %v\n", len(mfs.Blocks))
}
