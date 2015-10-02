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
	searchNode := rfs.BlockHandler.GetFreeBlockNode(SEARCHINDEX)
	searchIndex := SearchIndex{Node: searchNode, Terms: make(map[string]BlockNode)}

	rfs.BlockHandler.SaveRawBlock(searchNode, rawBlock(searchIndex))

	Root := rfs.BlockHandler.SaveRawBlock(blockNode, rawBlock(rdn))
	sb := SuperBlockNode{SuperBlock, bc, bs, Root, searchNode}

	rfs.BlockHandler.SaveRawBlock(SuperBlock, rawBlock(sb))
	rfs.SuperBlock = sb
}

func (rfs *RootFileSystem) GetFileOrDirectory(path string, createIfNotExist bool) (*FileNode, *DirectoryNode, error) {
	dn, _ := rfs.ChangeCache.GetDirectoryNode(rfs.SuperBlock.RootDirectory)

	var dnReal *DirectoryNode
	var fnReal *FileNode
	var err error

	if path == "/" {
		dnReal = dn
	} else {
		parts := strings.Split(path, "/")
		dnReal, err = dn.findDirectoryNode(parts[1:], rfs)
		// And now get the names of things and add them to "entries"
		// for now, don't do the continuation
		if err != nil {
			// Must be a file
			fnReal, err = dn.findNode(parts[1:], rfs, createIfNotExist)
		}
	}
	return fnReal, dnReal, err
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

// Delete the contents (that the fileName points to)
func (rfs *RootFileSystem) DeleteFile(fileName string) error {
	parts := strings.Split(fileName, "/")
	dn, _ := rfs.ChangeCache.GetDirectoryNode(rfs.SuperBlock.RootDirectory)
	fn, err := dn.findNode(parts[1:], rfs, false)
	dnReal, err2 := dn.findDirectoryNode(parts[1:len(parts)-1], rfs)
	if err == nil && err2 == nil {
		rfs.deliverMessage("Removing blocks")
		blocks := make([]BlockNode, 0)
		//blocks = append(blocks, fn.Node)
		for _, v := range fn.DataBlocks {
			blocks = append(blocks, v)
		}
		for _, v := range fn.AlternateRoutes {
			blocks = append(blocks, v)
		}

		rfs.BlockHandler.FreeBlocks(blocks)
		// Now update the directory node
		delete(dnReal.Files, parts[len(parts)-1])
		rfs.ChangeCache.DeleteFileNode(fn)
		rfs.ChangeCache.SaveDirectoryNode(dnReal)
		// TODO Also remove from search index
		return nil
	}
	return err
}

func (rfs *RootFileSystem) RetrieveFileNode(id BlockNode) (*FileNode, error) {
	rawBlock := rfs.BlockHandler.GetRawBlock(id)
	return getFileNode(rawBlock), nil
}

func (rfs *RootFileSystem) RetrieveDirectoryNode(id BlockNode) (*DirectoryNode, error) {
	rawBlock := rfs.BlockHandler.GetRawBlock(id)
	return getDirectoryNode(rawBlock), nil
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
				// TODO Change index
			}
		}
		// Now save them
		if !withinNode {
			rfs.ChangeCache.SaveDirectoryNode(sourceNode)
		}
		rfs.ChangeCache.SaveDirectoryNode(targetNode)
		return nil
	}
}

func (rfs *RootFileSystem) GetBlock(fileNode *FileNode, tag string, start string, end string) (*BlockStructure, error) {
	route := fileNode.DefaultRoute.DataBlockNames
	if len(tag) != 0 {
		routeNode, ok := fileNode.AlternateRoutes[tag]
		if !ok {
			return nil, errors.New("No tag found")
		}
		r := getRoute(rfs.BlockHandler.GetRawBlock(routeNode))
		route = r.DataBlockNames
	}
	// Now we need to filter DataBlockNames
	newRoute := make([]string, 0)
	foundStart := (len(start) == 0)
	foundEnd := false
	for entry := range route {
		if !foundStart {
			if route[entry] >= start {
				foundStart = true
			}
		}
		if !foundEnd {
			if len(end) > 0 && route[entry] > end {
				foundEnd = true
			}
		}
		if foundStart && !foundEnd {
			newRoute = append(newRoute, route[entry])
		}
	}

	// Ok now newRoute contains the list of block names
	blockStructure := BlockStructure{}
	blockStructure.Blocks = make([]Block, 0)

	for point := range newRoute {
		rNode, ok := fileNode.DataBlocks[newRoute[point]]
		if !ok {
			return nil, errors.New("Invalid block structure")
		}
		data := rfs.BlockHandler.GetRawBlock(rNode)
		b := Block{}
		b.Key = newRoute[point]
		b.Value = string(data)
		blockStructure.Blocks = append(blockStructure.Blocks, b)
	}
	return &blockStructure, nil
}

// Appends the content to the given file, creating the file if it doesn't exist
func (rfs *RootFileSystem) AppendFile(fileName string, contents []byte) error {
	fn, err := rfs.retrieveFn(fileName, true)

	if err == nil {
		// We need to find the last data block, and append to the data of that block so that it is filled up,
		// then add any remaining data to a new data block
		// NEW - for versioning
		// We always append a new block (or blocks) for this data, the old route will be preserved
		// in the routes information. After appending the blocks, we update the DefaultRoute and copy the
		// DefaultRoute into the new version in the version route information

		rfs.saveNewData(fileName, fn, contents)
	} else {
		return err
	}
	return nil
}

// Retrieves the version tags
func (rfs *RootFileSystem) GetTags(fileName string) ([]string, error) {
	fn, err := rfs.retrieveFn(fileName, false)

	if err == nil {
		ret := make([]string, len(fn.AlternateRoutes), len(fn.AlternateRoutes))
		for k, _ := range fn.AlternateRoutes {
			ret = append(ret, k)
		}
		return ret, nil
	}
	return nil, err
}

// Writes a file, creating if it doesn't exist, overwriting if it does
func (rfs *RootFileSystem) WriteFile(fileName string, contents []byte) error {
	// Find record for this fileName from RootFileSystem
	// After splitting on /
	fn, err := rfs.retrieveFn(fileName, true)

	if err == nil {
		rfs.BlockHandler.FreeBlocks(fn.getBlocksToFree())

		rfs.saveNewData(fileName, fn, contents)

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

// Read all of the contents of the given file
func (rfs *RootFileSystem) ReadFileTag(fileName string, tagName string) ([]byte, error) {
	// Traverse the directory node system to find the BlockNode for the FileNode
	// Load that up, and read from the Blocks, appending to a single bytebuffer and then return that
	// If the ContinuationNode is set, load that one and carry on there

	fn, err := rfs.retrieveFn(fileName, false)

	if err == nil {
		routeBlock, ok := fn.AlternateRoutes[tagName]
		if !ok {
			return nil, errors.New("That tag does not exist")
		}
		route := rfs.getRoute(routeBlock)
		buffer := new(bytes.Buffer)
		for _, i := range route.DataBlockNames {
			data := rfs.BlockHandler.GetRawBlock(fn.DataBlocks[i])
			buffer.Write(data)
		}
		return buffer.Bytes(), nil
	} else {
		return nil, err
	}
}
