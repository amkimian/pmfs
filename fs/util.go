package fs

import "bytes"

import "encoding/gob"

func rawBlock(sb interface{}) []byte {
	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.
	err := enc.Encode(sb)
	if err == nil {
		return buffer.Bytes()
	}
	return nil
}

func getRoute(contents []byte) *DataRoute {
	buffer := bytes.NewBuffer(contents)
	dec := gob.NewDecoder(buffer)
	var ret DataRoute
	dec.Decode(&ret)
	return &ret
}

func getDirectoryNode(contents []byte) *DirectoryNode {
	buffer := bytes.NewBuffer(contents)

	dec := gob.NewDecoder(buffer)
	var ret DirectoryNode
	dec.Decode(&ret)
	return &ret
}

func getFileNode(contents []byte) *FileNode {
	buffer := bytes.NewBuffer(contents)

	dec := gob.NewDecoder(buffer)
	var ret FileNode
	dec.Decode(&ret)
	return &ret
}
