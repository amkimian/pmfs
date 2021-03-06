// The fs package represents the abstract file system
package fs

import "time"

// BlockNodeType differentiates between the different types of Node in a filesystem
type BlockNodeType int

const (
	SUPERBLOCK BlockNodeType = iota
	DIRECTORY
	FILE
	ROUTE
	DATA
	SEARCHINDEX
	SEARCHTREE
	NIL
)

// A File in the file system can be either a normal file (containing data) or
// a mounted filesystem. With a custom (non NORMAL) file type the file system
// understands the format of the contained data
type FileType int

const (
	NORMAL FileType = iota
	MOUNT
)

type BlockStructure struct {
	Blocks []Block
}

type Block struct {
	Key   string
	Value interface{}
}

// The RootFileSystem contains the information about this filesystem and is the main
// entry point for working with a file system.
type RootFileSystem struct {
	BlockHandler  BlockHandler
	Configuration string
	SuperBlock    SuperBlockNode
	Notification  chan string
	ChangeCache   Cache
}

// A BlockNode has a type and a unique id in the filesystem
type BlockNode struct {
	RelativeTo int
	Type       BlockNodeType
	Id         int
}

// The NilBlock is used to represent "no node"
var NilBlock BlockNode = BlockNode{-1, NIL, -1}

// The SuperBlock block node is always at position 0 in the filesystem
var SuperBlock BlockNode = BlockNode{-1, SUPERBLOCK, 0}

// This contains permissions and stats about a directory or file
type FileStats struct {
	// When this file was created
	Created time.Time
	// When this file was modified
	Modified time.Time
	// When this file was last accessed (currently not written to except at the start)
	Accessed time.Time
	// Who is the owner of the file
	Owner int
	// What group owner for the file
	Group int
	// What permission mask for this file
	Permissions int
	// What is the size of the data associated with this file
	Size int
}

// A DirectoryNode contains information about a directory in a file system
type DirectoryNode struct {
	Node         BlockNode
	Stats        FileStats
	Folders      map[string]BlockNode
	Files        map[string]BlockNode
	Continuation BlockNode
	Attributes   map[string]interface{}
}

// This is the topmost node in a filesystem, always stored at node 0
type SuperBlockNode struct {
	Node            BlockNode
	BlockCount      int
	BlockSize       int
	RootDirectory   BlockNode
	SearchIndexNode BlockNode
}

type DataRoute struct {
	Node           BlockNode // for the default route this will be the null node
	RouteName      string
	DataBlockNames []string
}

// A FileNode contains information about a file in a file system (which may contain "special" data depending on its type)
type FileNode struct {
	Node            BlockNode
	Stats           FileStats
	Type            FileType
	DataBlocks      map[string]BlockNode // The key is the block id within this file. For some file types we manage this, for others the user does
	DefaultRoute    DataRoute
	AlternateRoutes map[string]BlockNode // The block node points to a data structure containing a DataRoute
	Version         int
	Attributes      map[string]interface{}
	LatestTag       string
}

// The storage for a file system must implement this
type BlockHandler interface {
	Init(configuration string)
	Format(blockCount int, blockSize int)
	GetFreeBlockNode(NodeType BlockNodeType) BlockNode
	GetFreeDataBlockNode(parent BlockNode, id string) BlockNode
	GetRawBlock(node BlockNode) []byte
	SaveRawBlock(node BlockNode, data []byte) BlockNode
	FreeBlocks(blocks []BlockNode)
	DumpInfo()
}
