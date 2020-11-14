package deduper

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// Using memory fs rather than mocks for ease

func TestFileWalker_Walk_Ok(t *testing.T) {
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
