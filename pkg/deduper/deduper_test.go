package deduper

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestMerge_if_not_null(t *testing.T) {
	hash := []byte("ABC")
	f1 := IndexedFile{
		Path:        "hello",
		Md5Checksum: []byte("XXX"),
	}
	f2 := IndexedFile{
		Path:        "hello foo",
		Md5Checksum: hash,
	}
	f1.merge(f2)
	assert.Equal(t, hash, f1.Md5Checksum)
}

func TestMerge_do_nothing_when_null(t *testing.T) {
	hash := []byte("ABC")
	f1 := IndexedFile{
		Path:        "hello",
		Md5Checksum: hash,
	}
	f2 := IndexedFile{
		Path:        "hello foo",
		Md5Checksum: nil,
	}
	f1.merge(f2)
	assert.Equal(t, hash, f1.Md5Checksum)
}

// NOTE: not really unit tests with the "in-memory" fs
//  I should mock things out really, but given the code
//  is so FS heavy, this will do.
type MemoryFsTestSuite struct {
	suite.Suite
	fs        afero.Fs
	indexPath string
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(MemoryFsTestSuite))
}

func (suite *MemoryFsTestSuite) SetupTest() {
	suite.fs = afero.NewMemMapFs()
	suite.indexPath = "testDir/pictures"

	if err := suite.fs.MkdirAll(suite.indexPath, 0755); nil != err {
		fmt.Printf("Failed creating %v: %v", suite.indexPath, err)
	}
}

func (suite *MemoryFsTestSuite) TestDirExist_InvalidDir() {
	deduper := NewDeduper(suite.fs, suite.indexPath, true)

	err := deduper.IsDirExist("invalidDir")

	assert.Error(suite.T(), err)
}

func (suite *MemoryFsTestSuite) TestDirExist_ValidDir() {
	deduper := NewDeduper(suite.fs, suite.indexPath, true)

	err := deduper.IsDirExist(suite.indexPath)

	assert.Nil(suite.T(), err)
}

func (suite *MemoryFsTestSuite) TestDirExist_MoveDuplicates_ok() {
	deduper := NewDeduper(suite.fs, suite.indexPath, true)

	target := "testDir/temp"
	dupe1 := []string{"testDir/pictures/foo.txt", "testDir/pictures/bar.txt", "testDir/pictures/fred.txt"}
	dupe2 := []string{"testDir/pictures/hello/world/foo.txt", "testDir/pictures/another/dir/bar.txt"}

	for _, f := range append(dupe1, dupe2...) {
		if err := afero.WriteFile(suite.fs, f, []byte(fmt.Sprintf("content: %s", f)), 0644); nil != err {
			fmt.Errorf("failed to create test file %v: %w", f, err)
		}
	}

	err := deduper.MoveDuplicates([][]string{dupe1, dupe2}, target)

	assert.Nil(suite.T(), err)
	// boohoo, bad unit test, many asserts :(
	moved, _ := afero.Exists(suite.fs, "testDir/temp/bar.txt")
	assert.True(suite.T(), moved)
	moved, _ = afero.Exists(suite.fs, "testDir/temp/fred.txt")
	assert.True(suite.T(), moved)
	moved, _ = afero.Exists(suite.fs, "testDir/temp/another/dir/bar.txt")
	assert.True(suite.T(), moved)

	notmoved, _ := afero.Exists(suite.fs, "testDir/pictures/foo.txt")
	assert.True(suite.T(), notmoved)

	removed, _ := afero.Exists(suite.fs, "testDir/pictures/bar.txt")
	assert.False(suite.T(), removed)
}
