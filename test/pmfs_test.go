package pmfs_test

import (
	"bytes"
	"os"
	"testing"
)
import "fmt"
import "github.com/amkimian/pmfs/fs"
import "github.com/amkimian/pmfs/memory"

var f fs.RootFileSystem
var mh memory.MemoryFileSystem

func statFile(name string) {
	stats, e2 := f.StatFile(name)
	if e2 != nil {
		fmt.Println(e2)
	} else {
		fmt.Printf("File %s, Size: %d\nCreated : %v\nModified : %v\nAccessed : %v\n", name, stats.Stats.Size, stats.Stats.Created, stats.Stats.Modified, stats.Stats.Accessed)
	}
}

func contentsFile(name string) (string, error) {
	x, err := f.ReadFile(name)
	if err != nil {
		fmt.Println(err)
		return "", err
	} else {
		//fmt.Printf("Name %s\nData is %v\n", name, string(x))
		return string(x), nil
	}
}

func dir(path string) int {
	names, _ := f.ListDirectory(path)
	fmt.Printf("Directory of %s\n", path)
	for y := range names {
		fmt.Println(names[y])
	}
	return len(names)
}

func TestMain(m *testing.M) {

	f.Init(&mh, "")
	f.Format(100, 100)
	go func() {
		for m := range f.Notification {
			fmt.Println(m)
		}
	}()
	os.Exit(m.Run())
}

func TestAppend(m *testing.T) {
	f.AppendFile("/append/1", []byte("Hello world 1"))
	f.AppendFile("/append/1", []byte(", Hello world 2"))
	v, err := contentsFile("/append/1")
	if err != nil {
		m.Errorf("%v", err)
	}
	if v != "Hello world 1, Hello world 2" {
		fmt.Println(v)
		m.Error(v)
	}
}
func TestSimpleReadWrite(m *testing.T) {
	f.WriteFile("/fred/alan", []byte("Hello world"))

	if 1 != dir("/fred") {
		m.Error("Directory size wrong")
	}

	v, err := contentsFile("/fred/alan")
	if err != nil {
		m.Errorf("%v", err)
	}
	if v != "Hello world" {
		m.Error("Contents not the same")
	}
	//statFile("/fred/alan")
}

func TestSecondWrite(m *testing.T) {
	f.WriteFile("/fred/other", []byte("Another file"))
	dir("/fred")
	contentsFile("/fred/other")
}

func TestReadBoth(m *testing.T) {
	statFile("/fred/other")
	statFile("/fred/alan")
}

func TestLoadsOfWrites(m *testing.T) {
	for i := 0; i < 20; i++ {
		fileName := fmt.Sprintf("/other/%v", i)
		f.WriteFile(fileName, []byte("Hello from me"))
	}
	dir("/other")
}

func TestAddAndDelete(m *testing.T) {
	f.WriteFile("/deleteme/1", []byte("One"))
	f.WriteFile("/deleteme/2", []byte("Two"))
	//fmt.Println("Before delete")
	if 2 != dir("/deleteme") {
		m.Error("File count before delete wrong")
	}
	f.DeleteFile("/deleteme/1")
	//fmt.Println("After removing 1")
	if 1 != dir("/deleteme") {
		m.Error("File count after delete wrong")
	}
}

func TestMultiBlockFile(m *testing.T) {
	buffer := new(bytes.Buffer)
	for i := 0; i < 100; i++ {
		buffer.WriteString("A reasonably long string \n")
	}
	f.WriteFile("/large/1", buffer.Bytes())
	statFile("/large/1")
	contentsFile("/large/1")
}

func TestDump(m *testing.T) {
	f.Dump()
}
