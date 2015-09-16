package fs

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
