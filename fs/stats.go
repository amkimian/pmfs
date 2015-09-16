package fs

import "time"

func (stats *FileStats) setNow() {
	stats.Created = time.Now()
	stats.modified()
}

func (stats *FileStats) accessed() {
  stats.Accessed = time.Now()
}

func (stats *FileStats) modified() {
  stats.Modified = time.Now()
  stats.accessed()
}
