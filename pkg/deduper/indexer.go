package deduper

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/spf13/afero"
)

const INDEX_NAME = ".duplicate-index.json"

type Indexer interface {
	Create(dir string) error
	Load() error
}

type indexerImp struct {
	fs        afero.Fs
	indexPath string
	index     *Index
	fileWalker
	fileHasher
	saver
	loader
}

func newIndexer(fs afero.Fs, indexPath string, index *Index, md5 bool, imageHash bool) Indexer {
	return &indexerImp{
		fs,
		indexPath,
		index,
		&fileSystemWalker{fs},
		newCompositeHasher(fs, md5, imageHash),
		&indexSaver{
			index,
			indexPath,
			fs,
		},
		&indexLoader{
			index,
			indexPath,
			fs,
		},
	}
}

type compositeHasher struct {
	hashers []fileHasher
}

func (hasher compositeHasher) hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func()) {
	for _, h := range hasher.hashers {
		h.hash(filePath, fun, errorFunc, func() {})
	}

	completeFun()
}

func newCompositeHasher(fs afero.Fs, md5 bool, imageHash bool) fileHasher {
	hashers := []fileHasher{}
	if md5 {
		hashers = append(hashers, &mdFiver{fs})
	}
	if imageHash {
		hashers = append(hashers, &imageHasher{fs})
	}

	return &compositeHasher{hashers}
}

func (i indexerImp) Create(dir string) error {
	start := time.Now()

	// find all files
	var wg sync.WaitGroup
	doneChannel := make(chan bool)
	errorChannel := make(chan error)
	fileCount := 0
	i.walk(dir, func(filePath string) {
		fileCount++
		// using routines to create md5 hashes of the files and store in index when done.
		wg.Add(1)
		go i.hash(filePath,
			func(f IndexedFile) {
				i.index.updateIndex(f)
			}, func(filePath string, err error) {
				errorChannel <- fmt.Errorf("error hashing file %v: %w\n", filePath, err)
			}, func() {
				defer wg.Done()
			})
	})

	// signal for done
	go func() {
		wg.Wait()
		close(doneChannel)
	}()

	//w wait for everything to finish or an error happens
	updateTimer := time.Now()
	for {
		select {
		case err := <-errorChannel:
			// give up when we encounter an error
			return err
		case <-doneChannel:
			fmt.Printf("Done indexing %v files in %v\n", len(i.index.ind), time.Since(start))
			// save index
			if err := i.save(); nil != err {
				return err
			}

			return nil
		default:
			if time.Since(updateTimer).Seconds() > 5 {
				fmt.Printf("Indexed %v/%v files in %v\n", len(i.index.ind), fileCount, time.Since(start))
				updateTimer = time.Now()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (i *Index) updateIndex(f IndexedFile) {
	i.mu.Lock()
	if indexedKey, found := i.iMap[f.Path]; found {
		i.ind[indexedKey].merge(f)
	} else {
		i.ind = append(i.ind, f)
		i.iMap[f.Path] = len(i.ind) - 1
	}
	i.mu.Unlock()
}

type loader interface {
	Load() error
}

type indexLoader struct {
	*Index
	indexPath string
	afero.Fs
}

func (i indexLoader) Load() error {
	fp := filepath.Join(i.indexPath, INDEX_NAME)
	jsonFile, err := i.Open(fp)
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

	i.ind = ind
	i.iMap = iMap

	return nil
}

type saver interface {
	save() error
}

type indexSaver struct {
	*Index
	indexPath string
	afero.Fs
}

func (i indexSaver) save() error {
	file, err := json.MarshalIndent(i.ind, "", " ")
	if nil != err {
		return fmt.Errorf("error creating index file: %w\n", err)
	}
	fp := filepath.Join(i.indexPath, INDEX_NAME)
	err = afero.WriteFile(i.Fs, fp, file, 0644)
	if nil != err {
		return fmt.Errorf("error saving index file to %v: %w\n", fp, err)
	}

	return err
}

type fileHasher interface {
	hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func())
}

type mdFiver struct {
	fs afero.Fs
}

func (fiver mdFiver) hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func()) {
	// open file (and close it when done)
	f, err := fiver.fs.Open(filePath)
	if err != nil {
		errorFunc(filePath, err)
		return
	}
	defer f.Close()

	// hash of the file
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		errorFunc(filePath, err)
		return
	}

	// do stuff
	fun(IndexedFile{
		Path:        filePath,
		Md5Checksum: h.Sum(nil),
	})

	completeFun()
}

type imageHasher struct {
	fs afero.Fs
}

func (hasher imageHasher) hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func()) {
	// open file (and close it when done)
	f, err := hasher.fs.Open(filePath)
	if err != nil {
		errorFunc(filePath, err)
		completeFun()
		return
	}
	defer f.Close()

	// assume jpeg for now
	jpg, err := jpeg.Decode(f)
	if _, isFormatError := err.(jpeg.FormatError); isFormatError {
		fmt.Printf("Skipping '%s', only supporting jpeg images.\n", filePath)
	} else if nil != err {
		// todo figure out error when not jpg and try something else
		errorFunc(filePath, err)
	} else {
		// hash of the file
		if h, err := goimagehash.DifferenceHash(jpg); nil != err {
			errorFunc(filePath, err)
		} else {
			fun(IndexedFile{
				Path: filePath,
				ImageHash: ImageHash{
					Kind: int(h.GetKind()),
					Hash: h.GetHash(),
				},
			})
		}
	}

	completeFun()
}

type fileWalker interface {
	walk(dir string, fun func(string)) error
}

type fileSystemWalker struct {
	fs afero.Fs
}

func (fw fileSystemWalker) walk(dir string, fun func(string)) error {
	// find all files
	err := afero.Walk(fw.fs, dir, func(path string, info os.FileInfo, err error) error {
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
