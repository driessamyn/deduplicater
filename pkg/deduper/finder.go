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
	for i, v := range finder.index.ind {
		key := string(v.Md5Checksum)
		if _, found := skip[key]; found {
			// already considered this duplicate
			continue
		}
		for _, vv := range finder.index.ind[i+1:] {
			if bytes.Compare(v.Md5Checksum, vv.Md5Checksum) == 0 {
				if _, ok := dupes[key]; !ok {
					skip[key] = true
					dupes[key] = []string{v.Path}
				}

				dupes[key] = append(dupes[key], vv.Path)

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
