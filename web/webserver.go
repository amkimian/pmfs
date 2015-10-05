// Web server for a pmfs system - used to view what's in the filesystem and some common actions on
// data (so it acts also as an API end point, for restful, for AJAX, etc.)
package web

import (
	"net/http"

	"fmt"

	"github.com/amkimian/pmfs/config"
	"github.com/amkimian/pmfs/fs"
)

var filesys *fs.RootFileSystem

// Start the web server - for the static pages and also the API end points
func StartAPIServer(rfs *fs.RootFileSystem) {
	filesys = rfs
	http.Handle("/", APIHandler{})
	serveString := fmt.Sprintf(":%d", config.PORT)
	fmt.Printf("Serving on %s\n", serveString)
	http.ListenAndServe(serveString, nil)
}

func StartWebServer() {
	fmt.Println("Serving web on 8080")
	http.ListenAndServe(":8080", http.FileServer(http.Dir("static/")))
}

type APIHandler struct {
}

func (apiHandler APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("URL Path is %s\n", r.URL.Path)
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
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
