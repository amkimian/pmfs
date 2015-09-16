package fs

import (
	"errors"
	"fmt"
)

func (dn *DirectoryNode) findDirectoryNode(paths []string, handler BlockHandler) (*DirectoryNode, error) {
	fmt.Printf("Search %v \n", paths)
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

// Returns the BlockNode and whether it is a directory or not
func (dn *DirectoryNode) findNode(paths []string, handler BlockHandler, createFileNode bool) (*FileNode, error) {
	if len(paths) == 1 {
		// This should be looking in the Files section and create if not exist (depending on createFileNode)
		nodeId, ok := dn.Files[paths[0]]
		var fileNode *FileNode
		if !ok {
			if createFileNode {
				nodeId = handler.GetFreeBlockNode(FILE)
				fileNode = &FileNode{Node: nodeId, Blocks: make([]BlockNode, 20), Continuation: NilBlock}
				fileNode.Stats.setNow()
				handler.SaveRawBlock(nodeId, rawBlock(fileNode))
				dn.Files[paths[0]] = nodeId
				handler.SaveRawBlock(dn.Node, rawBlock(dn))

				return fileNode, nil
			} else {
				return nil, errors.New("File not found")
			}
		} else {
			fileNode = getFileNode(handler.GetRawBlock(nodeId))
		}
		return fileNode, nil
	} else {
		// This should look in the directories section and create a new directory node if that does not exist (depending on createFileNode)
		// Then recurse with a subset of the paths
		newDnId, ok := dn.Folders[paths[0]]
		var newDn *DirectoryNode
		if !ok {
			if createFileNode {
				// Create new DirectoryNode, write that, update this DirectoryNode, write that
				newDnId = handler.GetFreeBlockNode(DIRECTORY)
				newDn = &DirectoryNode{Node: newDnId, Folders: make(map[string]BlockNode), Files: make(map[string]BlockNode), Continuation: NilBlock}
				newDn.Stats.setNow()
				handler.SaveRawBlock(newDnId, rawBlock(newDn))
				dn.Folders[paths[0]] = newDnId
				handler.SaveRawBlock(dn.Node, rawBlock(dn))

			} else {
				return nil, errors.New("Directory not found")
			}
		} else {
			newDn = getDirectoryNode(handler.GetRawBlock(newDnId))
		}

		return newDn.findNode(paths[1:], handler, createFileNode)
	}
}
