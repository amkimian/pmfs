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

// Retrieve a directory node from either the cache or the FileSystem
func (c *Cache) GetDirectoryNode(nodeId BlockNode) (*DirectoryNode, error) {
	entry, ok := c.EntryMap[nodeId]
	var dirNode *DirectoryNode
	if !ok {
		rawData := c.Fs.BlockHandler.GetRawBlock(nodeId)
		dn := getDirectoryNode(rawData)
		newEntry := CacheEntry{nodeId, false, NONE, dn}
		c.EntryMap[nodeId] = newEntry
		return dn, nil
	} else {
		dirNode = entry.entry.(*DirectoryNode)
		return dirNode, nil
	}
}
