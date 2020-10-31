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
	skip := make(map[string]bool)
	cnt := 0
	// todo: optimise further to no compare items already compared.
	for i, v := range finder.index.ind {
		key := string(v.Md5Checksum)
		if _, found := skip[key]; found {
			// already considered this duplicate
			continue
		}
		for _, vv := range finder.index.ind[i+1:] {
			if bytes.Compare(v.Md5Checksum, vv.Md5Checksum) == 0 {
				skip[key] = true
				if _, ok := dupes[key]; !ok {
					dupes[key] = []string{v.Path}
				} else {
					dupes[key] = append(dupes[key], v.Path)
				}
				cnt++
			}
		}
	}

	all := make([][]string, len(dupes))
	i := 0
	for _, v := range dupes {
		all[i] = v
		i++
	}

	return all
}
