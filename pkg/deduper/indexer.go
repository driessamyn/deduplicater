package deduper

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

const INDEX_NAME = ".duplicate-index.json"

type Indexer interface {
	Create(dir string) error
	Load() error
}

type indexerImp struct {
	indexPath string
	md5       bool
	index     *Index
}

func (i indexerImp) Create(dir string) error {
	// find all files
	var wg sync.WaitGroup
	doneChannel := make(chan bool)
	errorChannel := make(chan error)
	findAll(dir, func(filePath string) {
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

func (i indexerImp) Load() error {
	fp := filepath.Join(i.indexPath, INDEX_NAME)
	jsonFile, err := os.Open(fp)
	if nil != err {
		return fmt.Errorf("error loading index file: %w\n", err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if nil != err {
		return fmt.Errorf("error reading index file: %w\n", err)
	}

	var ind []IndexedFile
	err = json.Unmarshal(byteValue, &ind)
	if nil != err {
		return fmt.Errorf("error parsing index file: %w\n", err)
	}

	iMap := make(map[string]int)
	for i, v := range ind {
		iMap[v.Path] = i
	}

	i.index = &Index{
		ind:  ind,
		iMap: iMap,
	}

	return nil
}

func (i indexerImp) updateIndex(f IndexedFile) {
	i.index.mu.Lock()
	if indexedKey, found := i.index.iMap[f.Path]; found {
		i.index.ind[indexedKey].merge(f)
	} else {
		i.index.ind = append(i.index.ind, f)
		i.index.iMap[f.Path] = len(i.index.ind) - 1
	}
	i.index.mu.Unlock()
	// fmt.Printf("%v(%v)\n", f.Path, f.Md5Checksum)
}

func (i indexerImp) save() error {
	file, err := json.MarshalIndent(i.index.ind, "", " ")
	if nil != err {
		return fmt.Errorf("error creating index file: %w\n", err)
	}
	fp := filepath.Join(i.indexPath, INDEX_NAME)
	err = ioutil.WriteFile(fp, file, 0644)
	if nil != err {
		return fmt.Errorf("error saving index file to %v: %w\n", fp, err)
	}

	return err
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
