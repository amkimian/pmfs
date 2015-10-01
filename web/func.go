package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amkimian/pmfs/fs"
)

func statFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		w.WriteHeader(http.StatusNotFound)
	} else if dirNode != nil {
		var b []byte
		b, err = json.MarshalIndent(dirNode, "", "    ")
		fmt.Fprintf(w, "%v", string(b))
	} else if fileNode != nil {
		var b []byte
		b, err = json.MarshalIndent(fileNode, "", "    ")
		fmt.Fprintf(w, "%v", string(b))
	}
}

func getFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	// This can be two things
	// 1 get of a file, so dump the contents
	// 2 get of a folder, so construct some nice json

	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path)

	if err != nil {
		fmt.Fprintf(w, "%v", err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		if fileNode != nil {
			var x []byte
			x, err = filesys.ReadFile(r.URL.Path)
			fmt.Fprintf(w, "%v", string(x))
		} else {
			// Need to get file directory structure as a json object
			dirStructure := getDirStructure(r.URL.Path, dirNode, filesys)
			var b []byte
			b, err = json.MarshalIndent(dirStructure, "", "    ")
			fmt.Fprintf(w, "%v", string(b))
		}
	}
}

// Retrieve a specific version of a *file* based node
// The version tag must be in the filenodes AlternateRoutes, you can get the tag
// names from a stat call
func verGetFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	arr, err := filesys.ReadFileTag(r.URL.Path, r.Form["tag"][0])
	if err == nil {
		fmt.Fprintf(w, "%v", string(arr))
	}
}

// Delete this path, and all of its versions
func deleteFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	err := filesys.DeleteFile(r.URL.Path)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Removed %s", r.URL.Path)
	}
}

// Add a new file, with optional content, optional mime type
func addFileFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	filesys.WriteFile(r.URL.Path, []byte(r.Form["data"][0]))
	getFunc(w, r, filesys)
}

// Append data to a file, with an optional block name (for series files and the like). If the block name is specified
// it must not be present already (?) or it overwrites
func appendFileFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	filesys.AppendFile(r.URL.Path, []byte(r.Form["data"][0]))
	getFunc(w, r, filesys)
}

// Append a line to the data of a file, creating a new version. The data goes into a new block (with a CR added before)
// and a new version created using this block
func appendLineFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	filesys.AppendFile(r.URL.Path, []byte("\n"+r.Form["data"][0]))
	getFunc(w, r, filesys)
}
