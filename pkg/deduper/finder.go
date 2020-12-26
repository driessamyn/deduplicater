package deduper

import (
	"bytes"
	"errors"

	"github.com/corona10/goimagehash"
)

type Finder interface {
	// return list of duplicate (list of list)
	Find() ([][]string, error)
}

type CompositeFinder struct {
	md5       *md5Finder
	imageHash *imageHashFinder
}

func newCompositeFinder(md5 bool, imageHash bool, index *Index) Finder {
	var md5F *md5Finder
	var imageHashF *imageHashFinder
	if md5 {
		md5F = &md5Finder{index}
	}

	if imageHash {
		imageHashF = &imageHashFinder{index}
	}

	return &CompositeFinder{
		md5F,
		imageHashF,
	}
}

func (finder CompositeFinder) Find() ([][]string, error) {
	if nil == finder.md5 && nil == finder.imageHash {
		return nil, errors.New("Finder type must be specified (md5 or imagehash)")
	}

	if nil != finder.md5 && nil != finder.imageHash {
		return nil, errors.New("Finder only supports 1 type of hash at a time (md5 or imagehash)")
	}

	if nil != finder.md5 {
		return finder.md5.Find()
	}

	if nil != finder.imageHash {
		return finder.imageHash.Find()
	}

	panic("Oops")
}

type md5Finder struct {
	index *Index
}

func (finder md5Finder) Find() ([][]string, error) {
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

	return all, nil
}

type imageHashFinder struct {
	index *Index
}

func (finder imageHashFinder) Find() ([][]string, error) {
	dupes := make(map[uint64][]string)
	skip := make(map[uint64]bool)
	cnt := 0
	for i, v := range finder.index.ind {
		key := v.ImageHash.Hash
		if _, found := skip[key]; found {
			// already considered this duplicate
			continue
		}

		hash := goimagehash.NewImageHash(v.ImageHash.Hash, goimagehash.Kind(v.ImageHash.Kind))
		for _, vv := range finder.index.ind[i+1:] {
			hash2 := goimagehash.NewImageHash(vv.ImageHash.Hash, goimagehash.Kind(vv.ImageHash.Kind))
			distance, err := hash.Distance(hash2)

			if nil != err {
				return nil, err
			}

			if distance == 0 {
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

	return all, nil
}
