package fs

import (
	"fmt"
	"sort"
	"strings"
)

func (rfs *RootFileSystem) retrieveFn(fileName string, createNew bool) (*FileNode, error) {
	parts := strings.Split(fileName, "/")
	dn, _ := rfs.ChangeCache.GetDirectoryNode(rfs.SuperBlock.RootDirectory)
	return dn.findNode(parts[1:], rfs, createNew)
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
func safeAppend(target []byte, source []byte, maxSize int) ([]byte, []byte) {
	lt := len(target)
	toCopy := cap(target) - lt
	if toCopy > len(source) {
		toCopy = len(source)
	}
	target = append(target, source[0:toCopy]...)
	return target, source[toCopy:]
}

func getKeyName(val int) string {
	return fmt.Sprintf("%05d", val)
}

func (rfs *RootFileSystem) saveNewData(fn *FileNode, contents []byte) {
	newBlockId := len(fn.DataBlocks) + 1
	keyName := getKeyName(newBlockId)
	rfs.SaveNewBlock(fn, keyName, contents, false)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// This function appends data to this fileNode
func (rfs *RootFileSystem) SaveNewBlock(fn *FileNode, keyName string, contents []byte, sortBlocks bool) {
	fn.Stats.Size = fn.Stats.Size + len(contents)
	fn.Stats.modified()
	for i := 0; i < len(contents); i = i + rfs.SuperBlock.BlockSize {
		var toWrite []byte
		if i+rfs.SuperBlock.BlockSize > len(contents) {
			toWrite = contents[i:]
		} else {
			toWrite = contents[i : i+rfs.SuperBlock.BlockSize]
		}
		newDataNode, ok := fn.DataBlocks[keyName]
		if !ok {
			newDataNode = rfs.BlockHandler.GetFreeDataBlockNode(fn.Node, keyName)
			fn.DataBlocks[keyName] = newDataNode
			fn.DefaultRoute.DataBlockNames = append(fn.DefaultRoute.DataBlockNames, keyName)
			if sortBlocks {
				// We need to sort the Datablock names in the DefaultRoute
				sort.Strings(fn.DefaultRoute.DataBlockNames)
			}
			rfs.BlockHandler.SaveRawBlock(newDataNode, toWrite)
		} else {
			rfs.BlockHandler.SaveRawBlock(newDataNode, toWrite)
		}
	}
	// Now update the version route
	fn.Version++
	newVersionTag := fmt.Sprintf("v%d", fn.Version)
	routeBlockId := rfs.BlockHandler.GetFreeBlockNode(ROUTE)
	// Todo, put in cache
	rfs.BlockHandler.SaveRawBlock(routeBlockId, rawBlock(fn.DefaultRoute))
	fn.AlternateRoutes[newVersionTag] = routeBlockId
	fn.Stats.modified()
	rfs.ChangeCache.SaveFileNode(fn)
}

func getKeys(maps map[string]BlockNode) []string {
	keys := make([]string, 0, len(maps))
	for k := range maps {
		keys = append(keys, k)
	}
	return keys
}

func (rfs *RootFileSystem) getRoute(node BlockNode) *DataRoute {
	rawData := rfs.BlockHandler.GetRawBlock(node)
	return getRoute(rawData)
}

func (rfs *RootFileSystem) deliverMessage(msg string) {
	rfs.Notification <- msg
}
