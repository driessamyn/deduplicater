package deduper

import (
	"bytes"
)

type Finder interface {
	// return list of duplicate (list of list)
	Find() [][]string
}

type md5Finder struct {
	index *Index
}

func (finder md5Finder) Find() [][]string {
	dupes := make(map[string][]string)
	cnt := 0
	// todo: optimise further to no compare items already compared.
	for i, v := range finder.index.ind {
		for _, vv := range finder.index.ind[i+1:] {
			if bytes.Compare(v.Md5Checksum, vv.Md5Checksum) != 0 {
				key := string(v.Md5Checksum)
				if _, ok := dupes[key]; !ok {
					dupes[string(v.Md5Checksum)] = []string{v.Path}
				} else {
					dupes[string(v.Md5Checksum)] = append(dupes[string(v.Md5Checksum)], v.Path)
				}
				cnt++
			}
		}
	}

	all := make([][]string, 0, len(dupes))
	i := 0
	for _, v := range dupes {
		all[i] = v
		i++
	}

	return all
}
