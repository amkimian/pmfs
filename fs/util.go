package fs

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/petar/GoLLRB/llrb"
)

import "encoding/gob"

func rawBlock(sb interface{}) []byte {
	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.
	fmt.Printf("Raw block on type %v\n", reflect.TypeOf(sb))
	switch sb.(type) {
	case *SearchTree:
		// Need something special here
		fmt.Println("Doing something special with a SearchTree")
		st := sb.(*SearchTree)
		enc.Encode(st.Node)
		st.Tree.AscendGreaterOrEqual(st.Tree.Min(), func(i llrb.Item) bool {
			enc.Encode(i)
			return true
		})
		return buffer.Bytes()
	default:
		err := enc.Encode(sb)
		if err == nil {
			return buffer.Bytes()
		}
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

func getSearchTree(contents []byte) *SearchTree {
	buffer := bytes.NewBuffer(contents)
	dec := gob.NewDecoder(buffer)
	var ret SearchTree
	var node BlockNode
	dec.Decode(&node)
	ret.Node = node
	ret.Tree = llrb.New()
	for {
		var element SearchEntry
		err := dec.Decode(&element)
		if err == nil {
			fmt.Printf("Adding entry %v", element)
			ret.Tree.ReplaceOrInsert(element)
		} else {
			break
		}
	}
	return &ret
}

func getSearchIndex(contents []byte) *SearchIndex {
	buffer := bytes.NewBuffer(contents)
	dec := gob.NewDecoder(buffer)
	var ret SearchIndex
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
