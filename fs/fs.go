package fs

import (
	"bytes"
	"strings"
)

// Initialize a filesystem - storing the handler and then calling the Init method of the
// file system handler
func (rfs *RootFileSystem) Init(handler BlockHandler, configuration string) {
	rfs.BlockHandler = handler
	rfs.Configuration = configuration
	rfs.BlockHandler.Init(configuration)
}

// Dump to std out information about this filesystem (system specific)
func (rfs *RootFileSystem) Dump() {
	rfs.BlockHandler.DumpInfo()
}

// Format (initialize the contents) of this filesystem
func (rfs *RootFileSystem) Format(bc int, bs int) {
	rfs.BlockHandler.Format(bc, bs)
	// Write Raw Directory node
	rdn := DirectoryNode{Folders: make(map[string]BlockNode), Files: make(map[string]BlockNode), Continuation: NilBlock}
	rdn.Stats.setNow()
	blockNode := rfs.BlockHandler.GetFreeBlockNode(DIRECTORY)
	rdn.Node = blockNode

	Root := rfs.BlockHandler.SaveRawBlock(blockNode, rawBlock(rdn))
	sb := SuperBlockNode{SuperBlock, bc, bs, Root}

	rfs.BlockHandler.SaveRawBlock(SuperBlock, rawBlock(sb))
	rfs.SuperBlock = sb
}

// Return (currently) the list of entries at this point in the filesystem hiearchy
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

func addBlock(blocks []BlockNode, block BlockNode) []BlockNode {
	n := len(blocks)
	if n == cap(blocks) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]BlockNode, len(blocks), 2*len(blocks)+1)
		copy(newSlice, blocks)
		blocks = newSlice
	}
	blocks = blocks[0 : n+1]
	blocks[n] = block
	return blocks
}

func addBlocks(blocks []BlockNode, blockArr []BlockNode) []BlockNode {
	for i := range blockArr {
		blocks = addBlock(blocks, blockArr[i])
	}
	return blocks
}

func getKeys(maps map[string]BlockNode) []string {
	keys := make([]string, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

// Delete the contents (that the fileName points to)
func (rfs *RootFileSystem) DeleteFile(fileName string) error {
	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn, err := dn.findNode(parts[1:], rfs.BlockHandler, false)
	dnReal, err2 := dn.findDirectoryNode(parts[1:len(parts)-1], rfs.BlockHandler)
	if err == nil && err2 == nil {
		blocks := make([]BlockNode, 100)
		blocks = addBlock(blocks, fn.Node)
		blocks = addBlocks(blocks, fn.Blocks)
		contin := fn.Continuation
		for contin != NilBlock {
			fileBytes := rfs.BlockHandler.GetRawBlock(contin)
			fileNode := getFileNode(fileBytes)
			blocks = addBlocks(blocks, fileNode.Blocks)
			contin = fileNode.Continuation
		}
		rfs.BlockHandler.FreeBlocks(blocks)
		// Now update the directory node
		delete(dnReal.Files, parts[len(parts)-1])
		rfs.BlockHandler.SaveRawBlock(dnReal.Node, rawBlock(dnReal))

		return nil
	}
	return err
}

// Writes a file, creating if it doesn't exist, overwriting if it does
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

		fn.Stats.Size = len(contents)

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
				newContFileNode.Stats.setNow()
				fn.Continuation = newContNode
				rfs.BlockHandler.SaveRawBlock(fn.Node, rawBlock(fn))
				fn = &newContFileNode
			}
			newDataNode := rfs.BlockHandler.GetFreeBlockNode(DATA)
			fn.Blocks = append(fn.Blocks, newDataNode)
			fn.Stats.modified()
			rfs.BlockHandler.SaveRawBlock(newDataNode, toWrite)
		}

		fn.Stats.modified()
		rfs.BlockHandler.SaveRawBlock(fn.Node, rawBlock(fn))

		return nil
	} else {
		// Something went wrong, what to do? (probably propogate the error)
		return err
	}

}

// Return the stats of the passed file
func (rfs *RootFileSystem) StatFile(fileName string) (*FileStats, error) {
	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn, err := dn.findNode(parts[1:], rfs.BlockHandler, false)
	if err == nil {
		return &fn.Stats, nil
	} else {
		return nil, err
	}
}

// Read all of the contents of the given file
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
