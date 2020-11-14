package deduper

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Find_Md5(t *testing.T) {
	finder := md5Finder{
		&Index{
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
		},
	}

	dupes := finder.Find()

	assert.Equal(t, []string{"foo", "bar"}, dupes[0])
}
