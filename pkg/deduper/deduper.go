package deduper

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
)

type Deduper interface {
	Indexer
	Finder
	IsDirExist(target string) error
	MoveDuplicates(files [][]string, target string) error
}

type deduperImp struct {
	fs        afero.Fs
	indexPath string
	Indexer
	Finder
}

func NewDeduper(fs afero.Fs, indexPath string, md5 bool) Deduper {
	ind := &Index{
		iMap: make(map[string]int),
		ind:  []IndexedFile{},
	}
	// todo: composite finder
	f := &md5Finder{ind}
	return &deduperImp{
		fs,
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

func (f *IndexedFile) merge(mf IndexedFile) {
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
	exist, err := afero.DirExists(d.fs, target)

	if !exist {
		return fmt.Errorf("Directory %v does not exist", target)
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
			if _, err := d.fs.Stat(newPathDir); os.IsNotExist(err) {
				fmt.Printf("Creating target directory %v\n", newPathDir)
				d.fs.MkdirAll(newPathDir, os.ModePerm)
			}

			fmt.Printf("Moving %v to %v\n", file, newPath)
			err := d.fs.Rename(file, newPath)
			if nil != err {
				return fmt.Errorf("error moving %v to %v: %w\n", file, newPath, err)
			}
		}
	}

	return nil
}
