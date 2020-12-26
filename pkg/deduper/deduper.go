package deduper

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func NewDeduper(fs afero.Fs, indexPath string, md5 bool, imageHash bool) Deduper {
	ind := &Index{
		iMap: make(map[string]int),
		ind:  []IndexedFile{},
	}

	return &deduperImp{
		fs,
		indexPath,
		newIndexer(
			fs,
			indexPath,
			// just in memory dictionary for now - maybe need to do something better in the future
			ind,
			md5,
			imageHash,
		),
		newCompositeFinder(md5, imageHash, ind)}
}

// duplicate this as we cannot easily serialise the private members, and so to maintain decoupling.
//  TODO: combine md5 and image hash in generic dict of hashes
type ImageHash struct {
	Kind int
	Hash uint64
}

type IndexedFile struct {
	Path        string
	Md5Checksum []byte
	ImageHash   ImageHash
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
		// TODO: let user figure out which ones to delete and which to keep. For now, we keep the one that is nearest to the root
		sort.SliceStable(files, func(i, j int) bool {
			iDepth := strings.Count(files[i], "/")
			jDepth := strings.Count(files[j], "/")
			if iDepth == jDepth {
				// same level -> first alphabetical
				return strings.TrimSuffix(files[i], filepath.Ext(files[i])) < strings.TrimSuffix(files[j], filepath.Ext(files[j]))
			}
			return iDepth < jDepth
		})
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
