package web

import "github.com/amkimian/pmfs/fs"

type DirectoryStructure struct {
	FullPath string
	Files    []FileInfo
	Folders  []DirectoryInfo
}

type FileInfo struct {
	Name  string
	Stats fs.FileStats
	Type  fs.FileType
}

type DirectoryInfo struct {
	Name  string
	Stats fs.FileStats
}

func getDirStructure(fullName string, dirNode *fs.DirectoryNode, fs *fs.RootFileSystem) DirectoryStructure {
	ret := DirectoryStructure{}
	ret.FullPath = fullName
	ret.Files = make([]FileInfo, 0)
	ret.Folders = make([]DirectoryInfo, 0)
	for k, v := range dirNode.Files {
		f := FileInfo{}
		fNode, err := fs.RetrieveFileNode(v)
		if err == nil {
			f.Name = k
			f.Stats = fNode.Stats
			f.Type = fNode.Type
			ret.Files = append(ret.Files, f)
		}
	}

	for k, v := range dirNode.Folders {
		f := DirectoryInfo{}
		fNode, err := fs.RetrieveDirectoryNode(v)
		if err == nil {
			f.Name = k
			f.Stats = fNode.Stats
			ret.Folders = append(ret.Folders, f)
		}
	}

	return ret
}
