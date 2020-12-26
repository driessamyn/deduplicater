package deduper

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func Test_Update_Index_New(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{
			"foo": 0,
		},
		[]IndexedFile{
			{
				Path:        "foo",
				Md5Checksum: []byte("foo-md5"),
			},
		},
	}

	indexedFile := IndexedFile{
		Path:        "bar",
		Md5Checksum: []byte("bar-md5"),
	}

	index.updateIndex(indexedFile)

	assert.Equal(t, 0, index.iMap["foo"])
	assert.Equal(t, 1, index.iMap["bar"])

	assert.Equal(t, "foo", index.ind[0].Path)
	assert.Equal(t, "bar", index.ind[1].Path)
	assert.Equal(t, []byte("bar-md5"), index.ind[1].Md5Checksum)
}

func Test_Update_Index_Update(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{
			"foo": 0,
		},
		[]IndexedFile{
			{
				Path:        "foo",
				Md5Checksum: []byte("foo-md5"),
			},
		},
	}

	indexedFile := IndexedFile{
		Path:        "foo",
		Md5Checksum: []byte("bar-md5"),
	}

	index.updateIndex(indexedFile)

	assert.Equal(t, 0, index.iMap["foo"])

	assert.Equal(t, "foo", index.ind[0].Path)
	assert.Equal(t, []byte("bar-md5"), index.ind[0].Md5Checksum)
}

// Using memory fs rather than mocks for ease

func Test_FileWalker_Walk_Ok(t *testing.T) {
	fs := afero.NewMemMapFs()
	walker := fileSystemWalker{fs}
	if err := afero.WriteFile(fs, "hello/foo/bar.txt", []byte("content: bar"), 0644); nil != err {
		fmt.Errorf("failed to create test file %v: %w", "foo/bar.txt", err)
	}

	found := false
	err := walker.walk("hello", func(s string) {
		assert.Equal(t, "hello/foo/bar.txt", s)
		found = true
	})

	assert.True(t, found)
	assert.NoError(t, err)
}

func Test_FileWalker_Walk_Error(t *testing.T) {
	fs := afero.NewMemMapFs()
	walker := fileSystemWalker{fs}
	if err := afero.WriteFile(fs, "hello/foo/bar.txt", []byte("content: bar"), 0644); nil != err {
		fmt.Errorf("failed to create test file %v: %w", "hello/foo/bar", err)
	}

	err := walker.walk("not-valid", func(s string) {})

	assert.Error(t, err)
}

func Test_MdFiver_Hash_Ok(t *testing.T) {
	fs := afero.NewMemMapFs()
	hasher := mdFiver{fs}
	if err := afero.WriteFile(fs, "bar.txt", []byte("content: bar"), 0644); nil != err {
		fmt.Errorf("failed to create test file %v: %w", "bar.txt", err)
	}

	complete := false
	hasher.hash("bar.txt", func(f IndexedFile) {
		assert.Equal(t, "bar.txt", f.Path)
		assert.Equal(t, []byte{0x96, 0x9c, 0xa5, 0x2e, 0x55, 0x1d, 0x80, 0x92, 0x66, 0xc6, 0x85, 0xf7, 0x4d, 0x53, 0x11, 0xd}, f.Md5Checksum)

		complete = true
	}, func(filePath string, err error) {
		assert.Fail(t, "error not expected")
	}, func() {})

	assert.True(t, complete)
}

func Test_MdFiver_Hash_No_file(t *testing.T) {
	fs := afero.NewMemMapFs()
	hasher := mdFiver{fs}

	hasher.hash("bar.txt", func(f IndexedFile) {
		assert.Fail(t, "Should not complete")
	}, func(filePath string, err error) {
		assert.Equal(t, "bar.txt", filePath)
		assert.Error(t, err)
	}, func() {})
}

func Test_ImageFiver_Hash_Ok(t *testing.T) {
	fs := afero.NewMemMapFs()
	hasher := imageHasher{fs}

	const fileName = "test.jpg"
	// HACK: bit of a hack with loading img from disk
	dat, _ := ioutil.ReadFile("../../test/cat1.jpg")

	if err := afero.WriteFile(fs, fileName, dat, 0644); nil != err {
		fmt.Errorf("failed to create test file %v: %w", fileName, err)
	}

	complete := false
	hasher.hash(fileName, func(f IndexedFile) {
		assert.Equal(t, fileName, f.Path)
		assert.Equal(t, 3, f.ImageHash.Kind)
		assert.Equal(t, uint64(0xc0a0b0f0f0f8c0c0), f.ImageHash.Hash, func() {})

		complete = true
	}, func(filePath string, err error) {
		assert.Fail(t, "error not expected")
	}, func() {})

	assert.True(t, complete)
}

func Test_ImageFiver_Hash_Wrong_Filetype(t *testing.T) {
	fs := afero.NewMemMapFs()
	hasher := imageHasher{fs}

	const fileName = "bar.txt"
	if err := afero.WriteFile(fs, fileName, []byte("content: bar"), 0644); nil != err {
		fmt.Errorf("failed to create test file %v: %w", fileName, err)
	}

	errorProcessed := false
	processed := false
	complete := false
	hasher.hash(fileName, func(f IndexedFile) {
		processed = true
	}, func(filePath string, err error) {
		fmt.Println(err.Error())
		errorProcessed = true
	}, func() {
		complete = true
	})

	// just skipping
	assert.False(t, processed)
	assert.False(t, errorProcessed)
	assert.True(t, complete)
}

func Test_ImageFiver_Hash_No_file(t *testing.T) {
	fs := afero.NewMemMapFs()
	hasher := imageHasher{fs}

	hasher.hash("bar.jpg", func(f IndexedFile) {
		assert.Fail(t, "Should not complete")
	}, func(filePath string, err error) {
		assert.Equal(t, "bar.jpg", filePath)
		assert.Error(t, err)
	}, func() {})
}

type mockHasher1 struct {
	hashCount int
}

func (hasher *mockHasher1) hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func()) {
	hasher.hashCount++
}

type mockHasher2 struct {
	hashCount int
}

func (hasher *mockHasher2) hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func()) {
	hasher.hashCount++
}

func Test_Composite_Hasher_All(t *testing.T) {
	hasher1 := &mockHasher1{}
	hasher2 := &mockHasher2{}
	hasher := &compositeHasher{[]fileHasher{hasher1, hasher2}}

	hasher.hash("test.txt", func(f IndexedFile) {}, func(filePath string, err error) {}, func() {})

	assert.Equal(t, 1, hasher1.hashCount)
	assert.Equal(t, 1, hasher2.hashCount)
}

func Test_Saver_Ok(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{
			"foo": 0,
			"bar": 1,
		},
		[]IndexedFile{
			{
				Path:        "foo",
				Md5Checksum: []byte("foo-md5"),
				ImageHash: ImageHash{
					Kind: 0,
					Hash: 1234567890,
				},
			},
			{
				Path:        "bar",
				Md5Checksum: []byte("bar-md5"),
				ImageHash: ImageHash{
					Kind: 1,
					Hash: 91234567890,
				},
			},
		},
	}

	filePath := "index"
	fs := afero.NewMemMapFs()

	saver := indexSaver{index, filePath, fs}

	err := saver.save()
	assert.NoError(t, err)

	jsonFile, _ := fs.Open("index/" + INDEX_NAME)
	byteValue, _ := afero.ReadAll(jsonFile)
	actual := string(byteValue)

	assert.JSONEq(t, `[
 { "Path": "foo", "Md5Checksum": "Zm9vLW1kNQ==", "ImageHash": { "Kind": 0, "Hash": 1234567890} }, 
 { "Path": "bar",  "Md5Checksum": "YmFyLW1kNQ==", "ImageHash": { "Kind": 1, "Hash": 91234567890} }]`,
		actual)
}

func Test_Loader_Ok(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{
			"foo": 0,
		},
		[]IndexedFile{
			{
				Path:        "foo",
				Md5Checksum: []byte("foo-md5"),
				ImageHash: ImageHash{
					Kind: 0,
					Hash: 1234567890,
				},
			},
		},
	}

	filePath := "index"
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "index/"+INDEX_NAME, []byte("[{ \"Path\": \"bar\",  \"Md5Checksum\": \"YmFyLW1kNQ==\", \"ImageHash\": { \"Kind\": 3, \"Hash\": 9876543210} }]"), 0644)

	saver := indexLoader{index, filePath, fs}

	err := saver.Load()
	assert.NoError(t, err)
	assert.Equal(t, 0, index.iMap["bar"])
	assert.Equal(t, "bar", index.ind[0].Path)
	assert.Equal(t, "bar-md5", string(index.ind[0].Md5Checksum))
	assert.Equal(t, 3, index.ind[0].ImageHash.Kind)
	assert.True(t, 9876543210 == index.ind[0].ImageHash.Hash)
}

func Test_Loader_Invalid_Json(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{},
		[]IndexedFile{},
	}

	filePath := "index"
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "index/"+INDEX_NAME, []byte("[INVALID]"), 0644)

	saver := indexLoader{index, filePath, fs}

	err := saver.Load()
	assert.Error(t, err)
}

func Test_Loader_No_File(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{},
		[]IndexedFile{},
	}

	filePath := "index"
	fs := afero.NewMemMapFs()

	saver := indexLoader{index, filePath, fs}

	err := saver.Load()
	assert.Error(t, err)
}

type mockFileSystemWalker struct{}

var walkerMock func(dir string, fun func(string)) error

func (m mockFileSystemWalker) walk(dir string, fun func(string)) error {
	return walkerMock(dir, fun)
}

type mockFileHasher struct{}

var hasherMock func(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func())

func (m mockFileHasher) hash(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFun func()) {
	hasherMock(filePath, fun, errorFunc, completeFun)
}

type mockSaver struct{}

var saverMock func() error

func (m mockSaver) save() error {
	return saverMock()
}

type mockLoader struct{}

var loaderMock func() error

func (m mockLoader) Load() error {
	return loaderMock()
}

type IndexerTestSuite struct {
	suite.Suite
	*Index
	path   string
	hash   []byte
	walker *mockFileSystemWalker
	hasher *mockFileHasher
	saver  *mockSaver
	loader *mockLoader
	Indexer
}

func Test_CreateTestSuite(t *testing.T) {
	suite.Run(t, new(IndexerTestSuite))
}

func (suite *IndexerTestSuite) SetupTest() {
	fs := afero.NewMemMapFs()
	indexPath := "index"
	suite.Index = &Index{
		sync.Mutex{},
		map[string]int{},
		[]IndexedFile{},
	}

	suite.path = "foo.txt"
	suite.hash = []byte("foo")

	suite.walker = &mockFileSystemWalker{}
	walkerMock = func(dir string, fun func(string)) error {
		fun(suite.path)
		return nil
	}
	suite.hasher = &mockFileHasher{}

	hasherMock = func(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFunc func()) {
		fun(*&IndexedFile{
			Path:        suite.path,
			Md5Checksum: suite.hash,
		})
		completeFunc()
	}

	suite.saver = &mockSaver{}

	saverMock = func() error {
		return nil
	}

	suite.loader = &mockLoader{}

	suite.Indexer = &indexerImp{
		fs,
		indexPath,
		suite.Index,
		suite.walker,
		suite.hasher,
		suite.saver,
		suite.loader,
	}
}

func (suite *IndexerTestSuite) Test_Create_Ok() {
	isSaved := false

	saverMock = func() error {
		isSaved = true
		return nil
	}

	suite.Indexer.Create("dir")

	assert.True(suite.T(), isSaved, "Expected save to be called")
	assert.Equal(suite.T(), 0, suite.iMap[suite.path])
	assert.Equal(suite.T(), suite.path, suite.ind[0].Path)
	assert.Equal(suite.T(), suite.hash, suite.ind[0].Md5Checksum)
}

func (suite *IndexerTestSuite) Test_Create_Hash_Error() {
	path := "foo"
	raisedError := errors.New("Hashing failed")

	hasherMock = func(filePath string, fun func(f IndexedFile), errorFunc func(filePath string, err error), completeFunc func()) {
		errorFunc(path, raisedError)
		completeFunc()
	}

	err := suite.Indexer.Create("dir")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), path)
	assert.Equal(suite.T(), raisedError, errors.Unwrap(err))
}

func (suite *IndexerTestSuite) Test_Create_Save_Error() {
	raisedError := errors.New("Saving failed")

	saverMock = func() error {
		return raisedError
	}

	err := suite.Indexer.Create("dir")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err, raisedError)
}
