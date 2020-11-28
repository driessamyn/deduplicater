package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// end-to-end tests, using real files
type e2eTestSuite struct {
	suite.Suite

	testDir  string
	indexDir string
	moveDir  string
}

func TestE2eTestSuite(t *testing.T) {
	suite.Run(t, new(e2eTestSuite))
}

func (suite *e2eTestSuite) SetupTest() {
	var err error
	suite.testDir, err = ioutil.TempDir("", "test_")
	if err != nil {
		log.Fatal(err)
	}

	suite.indexDir, err = ioutil.TempDir("", "index_")
	if err != nil {
		log.Fatal(err)
	}

	suite.moveDir, err = ioutil.TempDir("", "move_")
	if err != nil {
		log.Fatal(err)
	}

	var currentDir string
	currentDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = copyTree(filepath.Join(currentDir, "../../test"), suite.testDir)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *e2eTestSuite) Test_Main_Move_Md5() {
	defer os.RemoveAll(suite.indexDir)
	defer os.RemoveAll(suite.testDir)

	// index - deduplicater index --md5 -d "/mnt/c/Users/bob/Pictures" -f "/mnt/c/Users/bob/Pictures"
	args := []string{
		"main",
		"index",
		"--md5",
		"-d",
		suite.testDir,
		"-f",
		suite.indexDir,
	}
	run(args)
	assert.FileExists(suite.T(), filepath.Join(suite.indexDir, ".duplicate-index.json"))

	// find --md5 -f "/mnt/c/Users/bob/Pictures"
	args = []string{
		"main",
		"find",
		"--md5",
		"-f",
		suite.indexDir,
		"--move-dir",
		suite.moveDir,
	}
	run(args)

	assert.FileExists(suite.T(), filepath.Join(suite.moveDir, "fred.txt"))
	assert.NoFileExists(suite.T(), filepath.Join(suite.testDir, "fred.txt"))
	assert.FileExists(suite.T(), filepath.Join(suite.testDir, "bob/freddy.txt"))
}

func assertFileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
}

// fixed and simplified version of https://github.com/termie/go-shutil/blob/master/shutil.go
func copyTree(src, dst string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcFileInfo.Mode())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		entryFileInfo, err := os.Lstat(srcPath)
		if err != nil {
			return err
		}

		if entryFileInfo.IsDir() {
			err = copyTree(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Do the actual copy
			fsrc, err := os.Open(srcPath)
			if err != nil {
				return err
			}
			defer fsrc.Close()

			fdst, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer fdst.Close()

			_, err = io.Copy(fdst, fsrc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
