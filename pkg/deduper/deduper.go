package deduper

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Deduper interface {
	Indexer
	Finder
	IsDirExist(target string) error
	MoveDuplicates(files [][]string, target string) error
}

type deduperImp struct {
	indexPath string
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
		indexPath,
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

func (d deduperImp) IsDirExist(target string) error {
	src, err := os.Stat(target)

	if os.IsNotExist(err) {
		return fmt.Errorf("Directory %v does not exist", target)
	}

	if !src.Mode().IsDir() {
		return fmt.Errorf("%v is not a directory", target)
	}

	return err
}

func (d deduperImp) MoveDuplicates(dupes [][]string, target string) error {
	for _, files := range dupes {
		// keeping the first one, moving the rest
		for _, file := range files[1:] {
			newPath := filepath.Join(target, file[len(d.indexPath):])
			newPathDir := filepath.Dir(newPath)
			// create dir if needed
			if _, err := os.Stat(newPathDir); os.IsNotExist(err) {
				fmt.Printf("Creating target directory %v\n", newPathDir)
				os.MkdirAll(newPathDir, os.ModePerm)
			}

			fmt.Printf("Moving %v to %v\n", file, newPath)
			err := os.Rename(file, newPath)
			if nil != err {
				return fmt.Errorf("error moving %v to %v: %w\n", file, newPath, err)
			}
		}
	}

	return nil
}
