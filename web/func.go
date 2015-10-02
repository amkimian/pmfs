package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amkimian/pmfs/fs"
)

// Add a new block to a structured file. The name of the block
// is in the parameter "block", the content is as before in "data"
func blockAddFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path, true)
	if err != nil {
		writeError(w, err)
	} else if dirNode != nil {
		writeError(w, errors.New("Cannot add to a directory"))
	} else {
		// We have a filenode...
		filesys.SaveNewBlock(r.URL.Path, fileNode, r.Form["block"][0], []byte(r.Form["data"][0]), true)
		getFunc(w, r, filesys)
	}
}

func getFormValue(r *http.Request, field string, def string) string {
	val, ok := r.Form[field]
	if !ok {
		return def
	} else {
		return val[0]
	}
}

// term, start, end
func attrFindFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	term := r.Form["term"][0]
	start := r.Form["start"][0]
	end := r.Form["end"][0]
	ret, err := filesys.SearchFindTerms(term, start, end)
	if err != nil {
		writeError(w, err)
	} else {
		var b []byte
		b, err = json.MarshalIndent(ret, "", "    ")
		fmt.Fprintf(w, "%v", string(b))
	}
}

// The block get Func retrieves a block given a range
// Parameters are
// start (optional) the start block (inclusive)
// end (optional) the end block (inclusive)
// tag (optional) the tag of the version to do this for
// Basically the function retrieves the route given the tag version (or default root)
// and then filters the block names given the start and end.
// The return data is a list structure containing the key (block name) and the value (the value of the block)
func blockGetFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path, false)
	if err != nil {
		writeError(w, err)
	} else if dirNode != nil {
		writeError(w, errors.New("Cannot do this to a directory"))
	} else {
		// We have a filenode...
		blockStructure, err := filesys.GetBlock(fileNode, getFormValue(r, "tag", ""), getFormValue(r, "start", ""), getFormValue(r, "end", ""))
		if err != nil {
			writeError(w, err)
		} else {
			var b []byte
			b, err = json.MarshalIndent(blockStructure, "", "    ")
			fmt.Fprintf(w, "%v", string(b))
		}
	}
}

func writeError(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "%v", err)
	w.WriteHeader(http.StatusNotFound)
}

func attrAddFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path, false)
	if err != nil {
		writeError(w, err)
	} else if dirNode != nil {
		dirNode.Attributes[r.Form["key"][0]] = r.Form["value"][0]
		filesys.ChangeCache.SaveDirectoryNode(dirNode)
		statFunc(w, r, filesys)
	} else if fileNode != nil {
		fileNode.Attributes[r.Form["key"][0]] = r.Form["value"][0]
		filesys.ChangeCache.SaveFileNode(fileNode)
		statFunc(w, r, filesys)
	}
}

// TBI
func attrListFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
}

func attrGetFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
}

func statFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path, false)
	if err != nil {
		writeError(w, err)
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

	fileNode, dirNode, err := filesys.GetFileOrDirectory(r.URL.Path, false)

	if err != nil {
		writeError(w, err)
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
	} else {
		writeError(w, err)
	}
}

// Delete this path, and all of its versions
func deleteFunc(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem) {
	err := filesys.DeleteFile(r.URL.Path)
	if err != nil {
		writeError(w, err)
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
