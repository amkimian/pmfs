package fs

import (
	"errors"
)

// Finds the parent directory, the one above this one
func (dn *DirectoryNode) findParentDirectoryNode(paths []string, handler BlockHandler, createDirectoryNode bool) (*DirectoryNode, error) {
	if len(paths) < 2 {
		return dn, nil
	}
	nodeId, ok := dn.Folders[paths[0]]
	var dirNode *DirectoryNode
	if !ok {
		if createDirectoryNode {
			dirNode = dn.createSubDirectory(paths[0], handler)

		} else {
			return nil, errors.New("Parent Folder not found")
		}
	} else {
		dirNode = getDirectoryNode(handler.GetRawBlock(nodeId))
	}
	if len(paths) == 2 {
		return dirNode, nil
	} else {
		return dirNode.findDirectoryNode(paths[1:], handler)
	}
}

func (dn *DirectoryNode) findDirectoryNode(paths []string, handler BlockHandler) (*DirectoryNode, error) {
	nodeId, ok := dn.Folders[paths[0]]
	var dirNode *DirectoryNode
	if !ok {
		return nil, errors.New("Folder not found")
	} else {
		dirNode = getDirectoryNode(handler.GetRawBlock(nodeId))
		if len(paths) == 1 {
			return dirNode, nil
		} else {
			return dirNode.findDirectoryNode(paths[1:], handler)
		}
	}
}

func (dn *DirectoryNode) createSubDirectory(name string, handler BlockHandler) *DirectoryNode {
	newDnId := handler.GetFreeBlockNode(DIRECTORY)
	newDn := &DirectoryNode{Node: newDnId, Folders: make(map[string]BlockNode), Files: make(map[string]BlockNode), Continuation: NilBlock}
	newDn.Stats.setNow()
	handler.SaveRawBlock(newDnId, rawBlock(newDn))
	dn.Folders[name] = newDnId
	handler.SaveRawBlock(dn.Node, rawBlock(dn))
	return newDn
}

func (dn *DirectoryNode) createNewFile(name string, handler BlockHandler) *FileNode {
	nodeId := handler.GetFreeBlockNode(FILE)
	fileNode := &FileNode{Node: nodeId, Blocks: make([]BlockNode, 0, 20), Continuation: NilBlock}
	fileNode.Stats.setNow()
	handler.SaveRawBlock(nodeId, rawBlock(fileNode))
	dn.Files[name] = nodeId
	handler.SaveRawBlock(dn.Node, rawBlock(dn))
	return fileNode
}

// Returns the BlockNode and whether it is a directory or not
func (dn *DirectoryNode) findNode(paths []string, handler BlockHandler, createFileNode bool) (*FileNode, error) {
	if len(paths) == 1 {
		// This should be looking in the Files section and create if not exist (depending on createFileNode)
		nodeId, ok := dn.Files[paths[0]]
		if !ok {
			if createFileNode {
				return dn.createNewFile(paths[0], handler), nil
			} else {
				return nil, errors.New("File not found")
			}
		} else {
			return getFileNode(handler.GetRawBlock(nodeId)), nil
		}
	} else {
		// This should look in the directories section and create a new directory node if that does not exist (depending on createFileNode)
		// Then recurse with a subset of the paths
		newDnId, ok := dn.Folders[paths[0]]
		var newDn *DirectoryNode
		if !ok {
			if createFileNode {
				newDn = dn.createSubDirectory(paths[0], handler)
			} else {
				return nil, errors.New("Directory not found")
			}
		} else {
			newDn = getDirectoryNode(handler.GetRawBlock(newDnId))
		}

		return newDn.findNode(paths[1:], handler, createFileNode)
	}
}
