// Web server for a pmfs system - used to view what's in the filesystem and some common actions on
// data (so it acts also as an API end point, for restful, for AJAX, etc.)
package web

import (
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"

	"fmt"

	"github.com/amkimian/pmfs/config"
)

// Start the web server - for the static pages and also the API end points
func StartServer() {
	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	serveString := fmt.Sprintf(":%d", config.PORT)
	http.ListenAndServe(serveString, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	assetName := config.WEBROOT + r.URL.Path
	extension := mime.TypeByExtension(filepath.Ext(assetName))
	data, _ := ioutil.ReadFile(assetName)
	w.Header().Set("Content-Type", extension)
	fmt.Fprintf(w, "%s", data)
}
