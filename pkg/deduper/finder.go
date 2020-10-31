package deduper

import (
	"bytes"
	"fmt"
)

type Finder interface {
	Find() map[string][]string
}

type md5Finder struct {
	index *Index
}

func (finder md5Finder) Find() map[string][]string {
	dupes := make(map[string][]string)
	for i, v := range finder.index.ind {
		for _, vv := range finder.index.ind[i+1:] {
			if bytes.Compare(v.Md5Checksum, vv.Md5Checksum) != 0 {
				fmt.Printf("Found duplicate: %v - %v", v.Path, vv.Path)
			}
		}
	}

	return dupes
}
