package indexer

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Indexer interface {
	Create() error
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

type index struct {
	mu   sync.Mutex
	iMap map[string]IndexedFile
}

type indexerImp struct {
	dir       string
	indexPath string
	md5       bool
	index     *index
}

func NewIndexer(dir string, indexPath string, md5 bool) Indexer {
	return indexerImp{
		dir:       dir,
		indexPath: indexPath,
		md5:       md5,
		// just in memory dictionary for now - maybe need to do something better in the future
		index: &index{
			iMap: make(map[string]IndexedFile),
		},
	}
}

func (i indexerImp) updateIndex(f IndexedFile) {
	i.index.mu.Lock()
	if indexedFile, found := i.index.iMap[f.Path]; found {
		indexedFile.merge(f)
	} else {
		i.index.iMap[f.Path] = f
	}
	i.index.mu.Unlock()
	// fmt.Printf("%v(%v)\n", f.Path, f.Md5Checksum)
}

func (i indexerImp) save() error {
	fmt.Printf("index: %v", *i.index)
	file, err := json.MarshalIndent(i.index.iMap, "", " ")
	if nil != err {
		return fmt.Errorf("error creating index file: %w\n", err)
	}
	fp := filepath.Join(i.indexPath, ".duplicate-index.json")
	err = ioutil.WriteFile(fp, file, 0644)
	if nil != err {
		return fmt.Errorf("error saving index file to %v: %w\n", fp, err)
	}

	return err
}

func (i indexerImp) Create() error {
	// find all files
	var wg sync.WaitGroup
	doneChannel := make(chan bool)
	errorChannel := make(chan error)
	findAll(i.dir, func(filePath string) {
		if i.md5 {
			// using routines to create md5 hashes of the files and store in index when done.
			wg.Add(1)
			go md5ChecksumFile(filePath, errorChannel, func(f IndexedFile) {
				defer wg.Done()
				i.updateIndex(f)
			})
		}
	})

	// signal for done
	go func() {
		wg.Wait()
		close(doneChannel)
	}()

	//w wait for everything to finish or an error happens
	select {
	case err := <-errorChannel:
		// give up when we encounter an error
		return err
	case <-doneChannel:
	}

	// save index
	if err := i.save(); nil != err {
		return err
	}

	return nil
}

func md5ChecksumFile(filePath string, errorChannel chan error, fun func(f IndexedFile)) {
	// open file (and close it when done)
	f, err := os.Open(filePath)
	if err != nil {
		errorChannel <- fmt.Errorf("error opening file %v: %w\n", filePath, err)
		return
	}
	defer f.Close()

	// hash of the file
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		errorChannel <- fmt.Errorf("error creating hash for file %v: %w\n", filePath, err)
		return
	}

	// do stuff
	fun(IndexedFile{
		Path:        filePath,
		Md5Checksum: h.Sum(nil),
	})
}

func findAll(dir string, fun func(string)) error {
	// find all files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing a path %q: %w\n", path, err)
		}
		if !info.IsDir() {
			fun(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %q: %w\n", dir, err)
	}

	return nil
}
