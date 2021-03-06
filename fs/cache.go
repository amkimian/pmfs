package fs

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// The cache is basically a deferred write cache. When changes are happening to a file system the fs will put these changes here
// and the cache will act on the activity in a deferred way.
// When reading from the filestore the cache is accessed first, with the assumption that if it's not in the cache here it can be read from
// the underlying filesystem (and then maybe placed into the cache)

type CacheAction int

const (
	NONE CacheAction = iota
	UPDATE
	DELETE
)

type CacheEntry struct {
	Node   BlockNode   // the id of this entry
	dirty  bool        // whether this entry has been written or sent back to the filesystem
	action CacheAction // what to do with this entry
	entry  interface{} // what we need to write or read
}

type Cache struct {
	EntryMap map[BlockNode]*CacheEntry
	Fs       *RootFileSystem
	c        chan BlockNode
	rwmutex  sync.RWMutex
}

func (c *Cache) Init(fs *RootFileSystem) {
	c.EntryMap = make(map[BlockNode]*CacheEntry)
	c.Fs = fs
	c.c = make(chan BlockNode)
	go cacheManager(c)
}

func (c *Cache) pushEntry(node BlockNode) {
	c.c <- node
}

// A goroutine for managing the output of the cache. We receive
// messages which determine what we should be doing, we do them and
// then update the cache. Periodically we also clean the cache out of non-dirty
// assets
func cacheManager(cache *Cache) {
	timer := time.Tick(10 * time.Second)
	for {
		select {
		case id := <-cache.c:
			// Peform the activity for the entry in the cache
			entry, ok := cache.EntryMap[id]
			if ok {
				//cache.rwmutex.Lock()
				if entry.action == UPDATE {
					vals := rawBlock(entry.entry)
					cache.Fs.deliverMessage(fmt.Sprintf("Cache save to fs, size is %d, type is %v", len(vals), reflect.TypeOf(entry.entry)))
					cache.Fs.BlockHandler.SaveRawBlock(id, rawBlock(entry.entry))
					entry.dirty = false
					cache.EntryMap[id] = entry
				} else if entry.action == DELETE {
					cache.Fs.deliverMessage("Cache delete from fs")
					blocks := make([]BlockNode, 1)
					blocks[0] = entry.Node
					cache.Fs.BlockHandler.FreeBlocks(blocks)
					entry.dirty = false
					cache.EntryMap[id] = entry
				}
				//cache.rwmutex.Unlock()
			}
		case <-timer:
			// Do clean up work
			cache.Fs.deliverMessage("Cache cleanup")
			cache.rwmutex.Lock()
			for n := range cache.EntryMap {
				if !cache.EntryMap[n].dirty {
					cache.Fs.deliverMessage("Found something to remove")
					delete(cache.EntryMap, n)
				}
			}
			cache.rwmutex.Unlock()
		}
	}
}

func (c *Cache) GetSearchIndex() *SearchIndex {
	entry, ok := c.EntryMap[c.Fs.SuperBlock.SearchIndexNode]
	var si *SearchIndex
	if !ok {
		si := c.Fs.getSearchIndex()
		newEntry := CacheEntry{c.Fs.SuperBlock.SearchIndexNode, false, NONE, si}
		c.rwmutex.Lock()
		c.EntryMap[c.Fs.SuperBlock.SearchIndexNode] = &newEntry
		c.rwmutex.Unlock()
		return si
	} else {
		c.Fs.deliverMessage("Search search index from cache")
		si = entry.entry.(*SearchIndex)
		return si
	}
}

func (c *Cache) GetSearchTree(nodeId BlockNode) (*SearchTree, error) {
	entry, ok := c.EntryMap[nodeId]
	var searchTree *SearchTree
	if !ok {
		c.Fs.deliverMessage("Put search tree in cache")
		rawData := c.Fs.BlockHandler.GetRawBlock(nodeId)
		searchTree := getSearchTree(rawData)
		newEntry := CacheEntry{nodeId, false, NONE, searchTree}
		c.rwmutex.Lock()
		c.EntryMap[nodeId] = &newEntry
		c.rwmutex.Unlock()
		return searchTree, nil
	} else {
		c.Fs.deliverMessage("Serve searchTree from cache")
		if entry.action != DELETE {
			searchTree = entry.entry.(*SearchTree)
			return searchTree, nil
		} else {
			return nil, errors.New("No search tree found, was deleted in cache")
		}
	}

}

// Retrieve a file node from either the cache or the FileSystem
func (c *Cache) GetFileNode(nodeId BlockNode) (*FileNode, error) {
	entry, ok := c.EntryMap[nodeId]
	var fileNode *FileNode
	if !ok {
		c.Fs.deliverMessage("Put file in cache")
		rawData := c.Fs.BlockHandler.GetRawBlock(nodeId)
		fn := getFileNode(rawData)
		newEntry := CacheEntry{nodeId, false, NONE, fn}
		c.rwmutex.Lock()
		c.EntryMap[nodeId] = &newEntry
		c.rwmutex.Unlock()
		return fn, nil
	} else {
		c.Fs.deliverMessage("Serve file from cache")
		if entry.action != DELETE {
			fileNode = entry.entry.(*FileNode)
			return fileNode, nil
		} else {
			return nil, errors.New("No file found, was deleted in cache")
		}
	}
}

// Retrieve a directory node from either the cache or the FileSystem
func (c *Cache) GetDirectoryNode(nodeId BlockNode) (*DirectoryNode, error) {
	entry, ok := c.EntryMap[nodeId]
	var dirNode *DirectoryNode
	if !ok {
		c.Fs.deliverMessage("Put dir in cache")
		rawData := c.Fs.BlockHandler.GetRawBlock(nodeId)
		dn := getDirectoryNode(rawData)
		newEntry := CacheEntry{nodeId, false, NONE, dn}
		c.rwmutex.Lock()
		c.EntryMap[nodeId] = &newEntry
		c.rwmutex.Unlock()
		return dn, nil
	} else {
		c.Fs.deliverMessage("Serve dir from cache")
		dirNode = entry.entry.(*DirectoryNode)
		return dirNode, nil
	}
}

func (c *Cache) SaveSearchIndex(searchIndex *SearchIndex) error {
	entry, ok := c.EntryMap[searchIndex.Node]
	c.rwmutex.Lock()
	if !ok {
		newEntry := CacheEntry{searchIndex.Node, true, UPDATE, searchIndex}
		c.EntryMap[searchIndex.Node] = &newEntry
	} else {
		entry.dirty = true
		entry.action = UPDATE
		entry.entry = searchIndex
		c.EntryMap[searchIndex.Node] = entry
	}
	c.pushEntry(searchIndex.Node)
	c.rwmutex.Unlock()
	return nil
}

func (c *Cache) SaveDirectoryNode(dirNode *DirectoryNode) error {
	entry, ok := c.EntryMap[dirNode.Node]
	c.rwmutex.Lock()
	if !ok {
		newEntry := CacheEntry{dirNode.Node, true, UPDATE, dirNode}
		c.EntryMap[dirNode.Node] = &newEntry
	} else {
		entry.dirty = true
		entry.action = UPDATE
		entry.entry = dirNode
		c.EntryMap[dirNode.Node] = entry
	}
	c.pushEntry(dirNode.Node)
	c.rwmutex.Unlock()
	return nil
}

func (c *Cache) DeleteFileNode(fileNode *FileNode) {
	entry, ok := c.EntryMap[fileNode.Node]
	c.rwmutex.Lock()

	if !ok {
		newEntry := CacheEntry{fileNode.Node, true, DELETE, fileNode}
		c.EntryMap[fileNode.Node] = &newEntry
	} else {
		entry.dirty = true
		entry.action = DELETE
		entry.entry = fileNode
		c.EntryMap[fileNode.Node] = entry
	}
	c.pushEntry(fileNode.Node)
	c.rwmutex.Unlock()
}

func (c *Cache) SaveSearchTree(searchTree *SearchTree) error {
	entry, ok := c.EntryMap[searchTree.Node]
	c.rwmutex.Lock()
	if !ok {
		newEntry := CacheEntry{searchTree.Node, true, UPDATE, searchTree}
		c.EntryMap[searchTree.Node] = &newEntry
	} else {
		entry.dirty = true
		entry.action = UPDATE
		entry.entry = searchTree
		c.EntryMap[searchTree.Node] = entry
	}
	c.pushEntry(searchTree.Node)
	c.rwmutex.Unlock()
	return nil
}

func (c *Cache) SaveFileNode(fileNode *FileNode) error {
	entry, ok := c.EntryMap[fileNode.Node]
	c.rwmutex.Lock()
	if !ok {
		newEntry := CacheEntry{fileNode.Node, true, UPDATE, fileNode}
		c.EntryMap[fileNode.Node] = &newEntry
	} else {
		entry.dirty = true
		entry.action = UPDATE
		entry.entry = fileNode
		c.EntryMap[fileNode.Node] = entry
	}
	c.pushEntry(fileNode.Node)
	c.rwmutex.Unlock()
	return nil
}
