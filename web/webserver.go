// Web server for a pmfs system - used to view what's in the filesystem and some common actions on
// data (so it acts also as an API end point, for restful, for AJAX, etc.)
package web

import (
	"encoding/json"
	"net/http"

	"fmt"

	"github.com/amkimian/pmfs/config"
	"github.com/amkimian/pmfs/fs"
)

type ApiRequest struct {
	processor func(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem)
}

var requests = map[string]ApiRequest{
	"get":  ApiRequest{getFunc},
	"stat": ApiRequest{statFunc},
}

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

var filesys *fs.RootFileSystem

// Start the web server - for the static pages and also the API end points
func StartServer(rfs *fs.RootFileSystem) {
	filesys = rfs
	http.HandleFunc("/", handler)
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	serveString := fmt.Sprintf(":%d", config.PORT)
	fmt.Printf("Serving on %s\n", serveString)
	http.ListenAndServe(serveString, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("URL Path is %s\n", r.URL.Path)
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Form values is %v\n", r.Form)
	// cmd defines the command, we route based on this
	_, ok := r.Form["cmd"]
	if !ok {
		r.Form["cmd"] = []string{"get"}
	}

	var switcher ApiRequest
	switcher, ok = requests[r.Form["cmd"][0]]
	if ok {
		switcher.processor(w, r, filesys)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	//assetName := config.WEBROOT + r.URL.Path
	//extension := mime.TypeByExtension(filepath.Ext(assetName))
	//data, _ := ioutil.ReadFile(assetName)
	//w.Header().Set("Content-Type", extension)
	//fmt.Fprintf(w, "%s", data)
}
