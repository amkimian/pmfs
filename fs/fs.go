package fs

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Initialize a filesystem - storing the handler and then calling the Init method of the
// file system handler
func (rfs *RootFileSystem) Init(handler BlockHandler, configuration string) {
	rfs.BlockHandler = handler
	rfs.Configuration = configuration
	rfs.BlockHandler.Init(configuration)
	rfs.Notification = make(chan string)
	rfs.ChangeCache.Init(rfs)
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

	dn, _ := rfs.ChangeCache.GetDirectoryNode(rfs.SuperBlock.RootDirectory)

	var dnReal *DirectoryNode
	var err error

	if path == "/" {
		dnReal = dn
	} else {
		parts := strings.Split(path, "/")
		dnReal, err = dn.findDirectoryNode(parts[1:], rfs)
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

func getKeys(maps map[string]BlockNode) []string {
	keys := make([]string, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

func (rfs *RootFileSystem) deliverMessage(msg string) {
	rfs.Notification <- msg
}

// Delete the contents (that the fileName points to)
func (rfs *RootFileSystem) DeleteFile(fileName string) error {
	parts := strings.Split(fileName, "/")
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	fn, err := dn.findNode(parts[1:], rfs, false)
	dnReal, err2 := dn.findDirectoryNode(parts[1:len(parts)-1], rfs)
	if err == nil && err2 == nil {
		rfs.deliverMessage("Removing blocks")
		blocks := make([]BlockNode, 0)
		blocks = append(blocks, fn.Node)
		for _, v := range fn.DataBlocks {
			blocks = append(blocks, v)
		}
		for _, v := range fn.AlternateRoutes {
			blocks = append(blocks, v)
		}

		rfs.BlockHandler.FreeBlocks(blocks)
		// Now update the directory node
		delete(dnReal.Files, parts[len(parts)-1])
		rfs.BlockHandler.SaveRawBlock(dnReal.Node, rawBlock(dnReal))

		return nil
	}
	return err
}

// Moving something means taking the definition (FileNode or DirectoryNode blockId) in one
// DirectoryNode entry and moving it to another DirectoryNode entry, creating that DirectoryNode
// in the target if it doesn't exist. Note that if we move filesystems we will have to actually
// copy the file/folder (<-- eek) and then delete from source
func (rfs *RootFileSystem) MoveFileOrFolder(source string, target string) error {
	// Get source DirectoryNode for this entity
	parts := strings.Split(source, "/")
	lastName := parts[len(parts)-1]
	rawRoot := rfs.BlockHandler.GetRawBlock(rfs.SuperBlock.RootDirectory)
	dn := getDirectoryNode(rawRoot)
	rfs.deliverMessage(fmt.Sprintf("Searching for source, parts is %v", parts))
	rfs.deliverMessage(fmt.Sprintf("Folders are %v", dn.Folders))
	sourceNode, err := dn.findParentDirectoryNode(parts[1:], rfs, false)
	if err != nil {
		return err
	} else {
		isFolderMove := false
		rfs.deliverMessage("Looking to find file or folder")
		blockId, ok := sourceNode.Files[lastName]
		if !ok {
			blockId, ok = sourceNode.Folders[lastName]
			if !ok {
				return errors.New("Source not found")
			}
			isFolderMove = true
		}
		targPaths := strings.Split(target, "/")
		lastTargName := targPaths[len(targPaths)-1]
		targetNode, err2 := dn.findParentDirectoryNode(targPaths[1:], rfs, true)
		if err2 != nil {
			return errors.New("Could not create or find target")
		}
		withinNode := false
		if isFolderMove {
			_, alreadyExists := targetNode.Folders[lastTargName]
			if alreadyExists {
				return errors.New("Target folder already exists")
			} else {
				// Move folder
				targetNode.Folders[lastTargName] = blockId
				delete(sourceNode.Folders, lastName)
			}
		} else {
			_, alreadyExists := targetNode.Files[lastTargName]
			if alreadyExists {
				return errors.New("Target file already exists")
			} else {
				// Move file
				if sourceNode.Node.Id == targetNode.Node.Id {
					withinNode = true
				}
				targetNode.Files[lastTargName] = blockId
				if withinNode {
					delete(targetNode.Files, lastName)
				} else {
					delete(sourceNode.Files, lastName)
				}
			}
		}
		// Now save them
		if !withinNode {
			rfs.BlockHandler.SaveRawBlock(sourceNode.Node, rawBlock(sourceNode))
		}
		rfs.BlockHandler.SaveRawBlock(targetNode.Node, rawBlock(targetNode))
		return nil
	}
}

// Appends the content to the given file, creating the file if it doesn't exist
func (rfs *RootFileSystem) AppendFile(fileName string, contents []byte) error {
	fn, err := rfs.retrieveFn(fileName, true)

	var currentData []byte
	if err == nil {
		// We need to find the last data block, and append to the data of that block so that it is filled up,
		// then add any remaining data to a new data block
		rfs.deliverMessage("Finding latest block")
		var currentBlockId BlockNode
		if len(fn.DefaultRoute.DataBlockNames) == 0 {
			rfs.deliverMessage("There is no data, creating...")
			var keyName = "00000"
			currentBlockId = rfs.BlockHandler.GetFreeDataBlockNode(fn.Node, keyName)
			currentData = make([]byte, 0, rfs.SuperBlock.BlockSize)
			fn.DataBlocks[keyName] = currentBlockId
			fn.DefaultRoute.DataBlockNames = append(fn.DefaultRoute.DataBlockNames, keyName)
			rfs.BlockHandler.SaveRawBlock(fn.Node, rawBlock(fn))
		} else {
			rfs.deliverMessage("Found last block")
			currentBlockId = fn.DataBlocks[fn.DefaultRoute.DataBlockNames[len(fn.DefaultRoute.DataBlockNames)-1]]
			currentData = rfs.BlockHandler.GetRawBlock(currentBlockId)
		}
		// Now fn will contain the appropriate filenode, and currentBlockId will be the currentBlockId to append to
		currentData, contents = safeAppend(currentData, contents, rfs.SuperBlock.BlockSize)
		rfs.deliverMessage("Saving data")
		rfs.BlockHandler.SaveRawBlock(currentBlockId, currentData)
		if len(contents) != 0 {
			rfs.saveNewData(fn, contents)
		}
	} else {
		return err
	}
	return nil
}

func safeAppend(target []byte, source []byte, maxSize int) ([]byte, []byte) {
	lt := len(target)
	toCopy := cap(target) - lt
	if toCopy > len(source) {
		toCopy = len(source)
	}
	target = append(target, source[0:toCopy]...)
	return target, source[toCopy:]
}

func (rfs *RootFileSystem) saveNewData(fn *FileNode, contents []byte) {

	fn.Stats.Size = len(contents)
	for i := 0; i < len(contents); i = i + rfs.SuperBlock.BlockSize {
		var toWrite []byte
		if i+rfs.SuperBlock.BlockSize > len(contents) {
			toWrite = contents[i:]
		} else {
			toWrite = contents[i : i+rfs.SuperBlock.BlockSize]
		}

		keyName := fmt.Sprintf("%05d", i)
		newDataNode := rfs.BlockHandler.GetFreeDataBlockNode(fn.Node, keyName)
		fn.DataBlocks[keyName] = newDataNode
		fn.DefaultRoute.DataBlockNames = append(fn.DefaultRoute.DataBlockNames, keyName)
		fn.Stats.modified()
		rfs.BlockHandler.SaveRawBlock(newDataNode, toWrite)
	}

	fn.Stats.modified()
	rfs.BlockHandler.SaveRawBlock(fn.Node, rawBlock(fn))

}

func (fn *FileNode) getBlocksToFree() []BlockNode {
	ret := make([]BlockNode, 10)
	for _, v := range fn.DataBlocks {
		ret = append(ret, v)
	}
	for _, v := range fn.AlternateRoutes {
		ret = append(ret, v)
	}
	return ret
}

// Writes a file, creating if it doesn't exist, overwriting if it does
func (rfs *RootFileSystem) WriteFile(fileName string, contents []byte) error {
	// Find record for this fileName from RootFileSystem
	// After splitting on /
	fn, err := rfs.retrieveFn(fileName, true)

	if err == nil {
		rfs.BlockHandler.FreeBlocks(fn.getBlocksToFree())

		rfs.saveNewData(fn, contents)

		return nil
	} else {
		rfs.deliverMessage("Could not get file node")
		// Something went wrong, what to do? (probably propogate the error)
		return err
	}

}

// Return the stats of the passed file
func (rfs *RootFileSystem) StatFile(fileName string) (*FileNode, error) {
	fn, err := rfs.retrieveFn(fileName, false)

	if err == nil {
		return fn, nil
	} else {
		return nil, err
	}
}

func (rfs *RootFileSystem) retrieveFn(fileName string, createNew bool) (*FileNode, error) {
	parts := strings.Split(fileName, "/")
	dn, _ := rfs.ChangeCache.GetDirectoryNode(rfs.SuperBlock.RootDirectory)
	return dn.findNode(parts[1:], rfs, createNew)
}

// Read all of the contents of the given file
func (rfs *RootFileSystem) ReadFile(fileName string) ([]byte, error) {
	// Traverse the directory node system to find the BlockNode for the FileNode
	// Load that up, and read from the Blocks, appending to a single bytebuffer and then return that
	// If the ContinuationNode is set, load that one and carry on there

	fn, err := rfs.retrieveFn(fileName, false)

	if err == nil {
		buffer := new(bytes.Buffer)
		for _, i := range fn.DefaultRoute.DataBlockNames {
			data := rfs.BlockHandler.GetRawBlock(fn.DataBlocks[i])
			buffer.Write(data)
		}
		return buffer.Bytes(), nil
	} else {
		return nil, err
	}
}
