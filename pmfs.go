package main

import "fmt"
import "github.com/amkimian/pmfs/fs"
import "github.com/amkimian/pmfs/memory"

// Phoenix Meta File System
// Abstracts the concept of a file system to underlying cloud block storage devices

func main() {
	fmt.Println("Starting PMFS Test")
	// Mount a memory root file system
	// Add some content, read it back
	var f fs.RootFileSystem
	var mh memory.MemoryFileSystem

	f.Init(&mh, "")
	f.Format(100, 100)
	f.WriteFile("/fred", []byte("Hello world"))
	x := f.ReadFile("/fred")
	fmt.Printf("Data is %v\n", x)
}
