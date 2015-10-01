package web

import (
	"net/http"

	"github.com/amkimian/pmfs/fs"
)

type ApiRequest struct {
	processor func(w http.ResponseWriter, r *http.Request, filesys *fs.RootFileSystem)
}

var requests = map[string]ApiRequest{
	"get":        ApiRequest{getFunc},
	"stat":       ApiRequest{statFunc},
	"verget":     ApiRequest{verGetFunc},
	"rm":         ApiRequest{deleteFunc},
	"addFile":    ApiRequest{addFileFunc},
	"appendFile": ApiRequest{appendFileFunc},
	"appendLine": ApiRequest{appendLineFunc},
}
