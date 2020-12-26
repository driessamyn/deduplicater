package deduper

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_No_Finders(t *testing.T) {
	finder := newCompositeFinder(false, false, &Index{})

	_, err := finder.Find()

	assert.Error(t, err)
}

func Test_Multiple_Finders(t *testing.T) {
	finder := newCompositeFinder(true, true, &Index{})

	_, err := finder.Find()

	assert.Error(t, err)
}

func Test_Find_Md5(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{},
		[]IndexedFile{
			{
				Path:        "foo",
				Md5Checksum: []byte("foo-md5"),
			},
			{
				Path:        "bar",
				Md5Checksum: []byte("foo-md5"),
			},
		},
	}
	finder := newCompositeFinder(true, false, index)

	dupes, _ := finder.Find()

	assert.Equal(t, []string{"foo", "bar"}, dupes[0])
}

func Test_Find_ImageHash(t *testing.T) {
	index := &Index{
		sync.Mutex{},
		map[string]int{},
		[]IndexedFile{
			{
				Path: "foo",
				ImageHash: ImageHash{
					Kind: 3,
					Hash: uint64(0xc0a0b0f0f0f8c0c0),
				},
			},
			{
				Path: "bar",
				ImageHash: ImageHash{
					Kind: 3,
					Hash: uint64(0xc0a0b0f0f0f8c0c0),
				},
			},
		},
	}
	finder := newCompositeFinder(false, true, index)

	dupes, _ := finder.Find()

	assert.Equal(t, []string{"foo", "bar"}, dupes[0])
}
