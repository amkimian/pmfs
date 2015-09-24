package fs

import (
	"errors"
)

// Finds the parent directory, the one above this one
func (dn *DirectoryNode) findParentDirectoryNode(paths []string, rfs *RootFileSystem, createDirectoryNode bool) (*DirectoryNode, error) {
	if len(paths) < 2 {
		return dn, nil
	}
	nodeId, ok := dn.Folders[paths[0]]
	var dirNode *DirectoryNode
	if !ok {
		if createDirectoryNode {
			dirNode = dn.createSubDirectory(paths[0], rfs)

		} else {
			return nil, errors.New("Parent Folder not found")
		}
	} else {

		dirNode, _ = rfs.ChangeCache.GetDirectoryNode(nodeId)
	}
	if len(paths) == 2 {
		return dirNode, nil
	} else {
		return dirNode.findDirectoryNode(paths[1:], rfs)
	}
}

func (dn *DirectoryNode) findDirectoryNode(paths []string, rfs *RootFileSystem) (*DirectoryNode, error) {
	nodeId, ok := dn.Folders[paths[0]]
	var dirNode *DirectoryNode
	if !ok {
		return nil, errors.New("Folder not found")
	} else {
		var err error
		dirNode, err = rfs.ChangeCache.GetDirectoryNode(nodeId)
		if len(paths) == 1 {
			return dirNode, err
		} else {
			return dirNode.findDirectoryNode(paths[1:], rfs)
		}
	}
}

func (dn *DirectoryNode) createSubDirectory(name string, rfs *RootFileSystem) *DirectoryNode {
	newDnId := rfs.BlockHandler.GetFreeBlockNode(DIRECTORY)
	newDn := &DirectoryNode{Node: newDnId, Folders: make(map[string]BlockNode), Files: make(map[string]BlockNode), Continuation: NilBlock}
	newDn.Stats.setNow()
	rfs.ChangeCache.SaveDirectoryNode(newDn)
	dn.Folders[name] = newDnId
	rfs.ChangeCache.SaveDirectoryNode(dn)
	return newDn
}

func (dn *DirectoryNode) createNewFile(name string, rfs *RootFileSystem) *FileNode {
	nodeId := rfs.BlockHandler.GetFreeBlockNode(FILE)
	fileNode := &FileNode{Node: nodeId, DataBlocks: make(map[string]BlockNode, 0), AlternateRoutes: make(map[string]BlockNode, 0)}
	fileNode.Stats.setNow()
	rfs.ChangeCache.SaveFileNode(fileNode)
	dn.Files[name] = nodeId
	rfs.ChangeCache.SaveDirectoryNode(dn)
	return fileNode
}

// Returns the BlockNode and whether it is a directory or not
func (dn *DirectoryNode) findNode(paths []string, rfs *RootFileSystem, createFileNode bool) (*FileNode, error) {
	if len(paths) == 1 {
		// This should be looking in the Files section and create if not exist (depending on createFileNode)
		nodeId, ok := dn.Files[paths[0]]
		if !ok {
			if createFileNode {
				return dn.createNewFile(paths[0], rfs), nil
			} else {
				return nil, errors.New("File not found")
			}
		} else {
			return rfs.ChangeCache.GetFileNode(nodeId)
		}
	} else {
		// This should look in the directories section and create a new directory node if that does not exist (depending on createFileNode)
		// Then recurse with a subset of the paths
		newDnId, ok := dn.Folders[paths[0]]
		var newDn *DirectoryNode
		if !ok {
			if createFileNode {
				newDn = dn.createSubDirectory(paths[0], rfs)
			} else {
				return nil, errors.New("Directory not found")
			}
		} else {
			newDn, _ = rfs.ChangeCache.GetDirectoryNode(newDnId)
		}

		return newDn.findNode(paths[1:], rfs, createFileNode)
	}
}
