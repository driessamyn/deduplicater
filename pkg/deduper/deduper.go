package deduper

import (
	"sync"
)

type Deduper interface {
	Indexer
	Finder
}

type deduperImp struct {
	Indexer
	Finder
}

func NewDeduper(indexPath string, md5 bool) Deduper {
	ind := &Index{
		iMap: make(map[string]int),
		ind:  []IndexedFile{},
	}
	// todo: composite finder
	f := &md5Finder{ind}
	return &deduperImp{
		indexerImp{
			indexPath,
			md5,
			// just in memory dictionary for now - maybe need to do something better in the future
			ind,
		},
		f}
}

type IndexedFile struct {
	Path        string
	Md5Checksum []byte
}

func (f IndexedFile) merge(mf IndexedFile) {
	if nil != mf.Md5Checksum {
		f.Md5Checksum = mf.Md5Checksum
	}
}

type Index struct {
	mu   sync.Mutex
	iMap map[string]int
	ind  []IndexedFile
}
