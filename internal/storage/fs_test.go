package storage_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cucumber/godog/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_Open_FS(t *testing.T) {
	tests := map[string]struct {
		fs fs.FS

		expData  []byte
		expError error
	}{
		"normal open": {
			fs: fstest.MapFS{
				"testfile": {
					Data: []byte("hello worlds"),
				},
			},
			expData: []byte("hello worlds"),
		},
		"file not found": {
			fs:       fstest.MapFS{},
			expError: errors.New("open testfile: file does not exist"),
		},
		"nil fs falls back on os": {
			expError: errors.New("open testfile: no such file or directory"),
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f, err := (storage.FS{FS: test.fs}).Open("testfile")
			if test.expError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expError.Error())
				return
			}

			assert.NoError(t, err)

			bb := make([]byte, len(test.expData))
			_, _ = f.Read(bb)
			assert.Equal(t, test.expData, bb)
		})
	}
}

func TestStorage_Open_OS(t *testing.T) {
	tests := map[string]struct {
		files    map[string][]byte
		expData  []byte
		expError error
	}{
		"normal open": {
			files: map[string][]byte{
				"testfile": []byte("hello worlds"),
			},
			expData: []byte("hello worlds"),
		},
		"nil fs falls back on os": {
			expError: errors.New("open %baseDir%/testfile: no such file or directory"),
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			baseDir := filepath.Join(os.TempDir(), t.Name(), "godogs")
			err := os.MkdirAll(baseDir+"/a", 0755)
			defer os.RemoveAll(baseDir)

			require.Nil(t, err)

			for name, data := range test.files {
				err := os.WriteFile(filepath.Join(baseDir, name), data, 0644)
				require.NoError(t, err)
			}

			f, err := (storage.FS{}).Open(filepath.Join(baseDir, "testfile"))
			if test.expError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, strings.ReplaceAll(test.expError.Error(), "%baseDir%", baseDir))
				return
			}

			assert.NoError(t, err)

			bb := make([]byte, len(test.expData))
			_, _ = f.Read(bb)
			assert.Equal(t, test.expData, bb)
		})
	}
}
