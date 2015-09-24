package fs

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
	EntryMap map[BlockNode]CacheEntry
	Fs       *RootFileSystem
}

func (c *Cache) Init(fs *RootFileSystem) {
	c.EntryMap = make(map[BlockNode]CacheEntry)
	c.Fs = fs
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
		c.EntryMap[nodeId] = newEntry
		return fn, nil
	} else {
		c.Fs.deliverMessage("Serve file from cache")
		fileNode = entry.entry.(*FileNode)
		return fileNode, nil
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
		c.EntryMap[nodeId] = newEntry
		return dn, nil
	} else {
		c.Fs.deliverMessage("Serve dir from cache")
		dirNode = entry.entry.(*DirectoryNode)
		return dirNode, nil
	}
}

func (c *Cache) SaveDirectoryNode(dirNode *DirectoryNode) error {
	entry, ok := c.EntryMap[dirNode.Node]
	if !ok {
		newEntry := CacheEntry{dirNode.Node, true, UPDATE, dirNode}
		c.EntryMap[dirNode.Node] = newEntry
	} else {
		entry.dirty = true
		entry.action = UPDATE
		entry.entry = dirNode
	}
	return nil
}

func (c *Cache) SaveFileNode(fileNode *FileNode) error {
	entry, ok := c.EntryMap[fileNode.Node]
	if !ok {
		newEntry := CacheEntry{fileNode.Node, true, UPDATE, fileNode}
		c.EntryMap[fileNode.Node] = newEntry
	} else {
		entry.dirty = true
		entry.action = UPDATE
		entry.entry = fileNode
	}
	return nil
}
