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
	f.WriteFile("/fred/alan", []byte("Hello world"))

	names, _ := f.ListDirectory("/fred")
	for y := range names {
		fmt.Println(names[y])
	}

	x, err := f.ReadFile("/fred/alan")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Data is %v\n", string(x))
	}

	stats, e2 := f.StatFile("/fred/alan")
	if e2 != nil {
		fmt.Println(e2)
	} else {
		fmt.Printf("Created : %v\nModified : %v\nAccessed : %v\n", stats.Created, stats.Modified, stats.Accessed)
	}
}
