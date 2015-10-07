package main

import "fmt"
import (
	"github.com/amkimian/pmfs/fs"
	"github.com/amkimian/pmfs/web"
)
import "github.com/amkimian/pmfs/memory"

// Phoenix Meta File System
// Abstracts the concept of a file system to underlying cloud block storage devices

func main() {
	fmt.Println("Starting PMFS")
	// Mount a memory root file system
	// Add some content, read it back
	var f fs.RootFileSystem
	var mh memory.MemoryFileSystem

	f.Init(&mh, "")
	go func() {
		for msg := range f.Notification {
			fmt.Println(msg)
		}
	}()
	f.Format(100, 100)
	f.WriteFile("/test/alan", []byte("Hello world this is a test"))
	f.AppendFile("/test/alan", []byte("\nThis is line 2, part of version 2"))
	f.AppendFile("/test/other", []byte("\nA new file"))
	f.AppendFile("/other/one", []byte("\nHere we go again"))
	f.AppendFile("/other/three", []byte("\nHello world"))

	names, _ := f.ListDirectory("/test")
	for y := range names {
		fmt.Println(names[y])
	}

	x, err := f.ReadFile("/test/alan")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Data is %v\n", string(x))
	}

	fn, e2 := f.StatFile("/test/alan")
	if e2 != nil {
		fmt.Println(e2)
	} else {
		stats := fn.Stats
		fmt.Printf("Created : %v\nModified : %v\nAccessed : %v\n", stats.Created, stats.Modified, stats.Accessed)
	}

	go web.StartAPIServer(&f)
	web.StartWebServer()
}
