package git

import (
	"os"
	"testing"
	//"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
)

type PackfileTestCase struct {
	Packfile        []byte
	ExpectedObjects map[string]bool
}

func runCase(label string, tc PackfileTestCase, t *testing.T) {
	// Create a new gitdir/client for each case so that we don't have stuff left
	// over from the previous one.
	// We don't care about the working directory, because Unpack doesn't use it.
	gitdir, err := ioutil.TempDir("", "gittest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gitdir)

	c, err := NewClient(gitdir, "")
	if err != nil {
		t.Fatal(err)
	}

	shas, err := UnpackObjects(c, UnpackObjectsOptions{}, bytes.NewReader(tc.Packfile))
	if err != nil {
		t.Fatal(err)
	}
	if g := len(shas); g != len(tc.ExpectedObjects) {
		t.Errorf("%s: Unexpected number of objects: got %v want %v", label, g, len(tc.ExpectedObjects))
		return
	}

	// Keep track of everything that we found while iterating through in a
	// map, so that we can easily compare that we got everything we expected
	// after.
	foundMap := make(map[string]bool)

	// Make sure everything we got was expected.
	for i, sha := range shas {
		foundMap[sha.String()] = true
		contains, ok := tc.ExpectedObjects[sha.String()]
		if !contains || !ok {
			t.Errorf("%s Unexpected SHA at index %d: %s", label, i, sha.String())
		}
	}

	// Make sure we got everything that was expected (nothing was missed)
	for sha, _ := range tc.ExpectedObjects {
		if contains, ok := foundMap[sha]; !contains || !ok {
			t.Errorf("%s: Did not get expected SHA1: %s", label, sha)
		}
	}
}

func TestPackfileUnpack(t *testing.T) {
	tests := []PackfileTestCase{
		// This is a small packfile generated with the official
		// git client by taking a blob with a bunch of lines,
		// hashing it, then deleting a couple and hashing that
		// and packing them into a packfile.
		// It should be close to the minimal packfile that has
		// a REF_DELTA delta.
		{
			[]byte{0x50, 0x41, 0x43, 0x4b, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0xbc, 0x08, 0x78, 0x9c,
				0x73, 0xe4, 0x72, 0xc4, 0x09, 0x9d, 0xb8, 0x9c, 0xb9, 0x5c, 0xb8, 0x5c, 0xe9, 0x46, 0x03, 0x00,
				0xcc, 0xc9, 0x15, 0x0f, 0x75, 0xbe, 0x22, 0xa5, 0xc7, 0xd7, 0xb2, 0x5c, 0x99, 0x0d, 0x89, 0xd7,
				0xc1, 0x83, 0x82, 0xf0, 0x81, 0x5f, 0x68, 0x3f, 0x17, 0x78, 0x9c, 0xeb, 0x61, 0x2c, 0x9a, 0x50,
				0x04, 0x00, 0x05, 0xad, 0x02, 0x02, 0x75, 0xb8, 0x48, 0x6c, 0x4d, 0xda, 0x43, 0x46, 0x17, 0x0b,
				0x29, 0x86, 0x95, 0x51, 0x0a, 0x29, 0x29, 0x68, 0x86, 0xe4},
			map[string]bool{
				"84dfc6fb0e86cf29049d53041e2d55f863eacfd8": true,
				"be22a5c7d7b25c990d89d7c18382f0815f683f17": true,
			},
		},
		{
			// Same as the above, but with a chain of length 2 by adding a third
			// modified blob.
			[]byte{
				0x50, 0x41, 0x43, 0x4b, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0xbc, 0x08, 0x78, 0x9c,
				0x73, 0xe4, 0x72, 0xc4, 0x09, 0x9d, 0xb8, 0x9c, 0xb9, 0x5c, 0xb8, 0x5c, 0xe9, 0x46, 0x03, 0x00,
				0xcc, 0xc9, 0x15, 0x0f, 0x75, 0xbe, 0x22, 0xa5, 0xc7, 0xd7, 0xb2, 0x5c, 0x99, 0x0d, 0x89, 0xd7,
				0xc1, 0x83, 0x82, 0xf0, 0x81, 0x5f, 0x68, 0x3f, 0x17, 0x78, 0x9c, 0xeb, 0x61, 0x2c, 0x9a, 0x50,
				0x04, 0x00, 0x05, 0xad, 0x02, 0x02, 0x75, 0x84, 0xdf, 0xc6, 0xfb, 0x0e, 0x86, 0xcf, 0x29, 0x04,
				0x9d, 0x53, 0x04, 0x1e, 0x2d, 0x55, 0xf8, 0x63, 0xea, 0xcf, 0xd8, 0x78, 0x9c, 0x2b, 0x4a, 0x9a,
				0x28, 0x90, 0x04, 0x00, 0x05, 0xfc, 0x01, 0xd8, 0x2d, 0xec, 0xe2, 0xa0, 0x76, 0x47, 0xdd, 0xad,
				0xd9, 0xae, 0xb3, 0x07, 0x4f, 0x8d, 0x9e, 0x62, 0x1b, 0xec, 0x69, 0x79,
			},
			map[string]bool{
				"84dfc6fb0e86cf29049d53041e2d55f863eacfd8": true,
				"be22a5c7d7b25c990d89d7c18382f0815f683f17": true,
				"bbd835f67c0ef19084d9b97e9219c1b38e66bd80": true,
			},
		},

		{
			// Same as the above, but using OFS_DELTA instead of
			// REF_DELTA
			[]byte{
				0x50, 0x41, 0x43, 0x4b, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0xbc, 0x08, 0x78, 0x9c,
				0x73, 0xe4, 0x72, 0xc4, 0x09, 0x9d, 0xb8, 0x9c, 0xb9, 0x5c, 0xb8, 0x5c, 0xe9, 0x46, 0x03, 0x00,
				0xcc, 0xc9, 0x15, 0x0f, 0x65, 0x18, 0x78, 0x9c, 0xeb, 0x61, 0x2c, 0x9a, 0x50, 0x04, 0x00, 0x05,
				0xad, 0x02, 0x02, 0x25, 0x15, 0xc5, 0xe5, 0xae, 0xc7, 0x2b, 0x3a, 0xc9, 0x80, 0xfc, 0x8b, 0x7f,
				0x61, 0xc8, 0xd0, 0x6d, 0xf0, 0x62, 0xf2,
			},
			map[string]bool{
				"84dfc6fb0e86cf29049d53041e2d55f863eacfd8": true,
				"be22a5c7d7b25c990d89d7c18382f0815f683f17": true,
			},
		},
		{
			// OFS_DELTA with a chain of length > 1
			[]byte{0x50, 0x41, 0x43, 0x4b, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0xbc, 0x08, 0x78, 0x9c,
				0x73, 0xe4, 0x72, 0xc4, 0x09, 0x9d, 0xb8, 0x9c, 0xb9, 0x5c, 0xb8, 0x5c, 0xe9, 0x46, 0x03, 0x00,
				0xcc, 0xc9, 0x15, 0x0f, 0x65, 0x18, 0x78, 0x9c, 0xeb, 0x61, 0x2c, 0x9a, 0x50, 0x04, 0x00, 0x05,
				0xad, 0x02, 0x02, 0x65, 0x0f, 0x78, 0x9c, 0x2b, 0x4a, 0x9a, 0x28, 0x90, 0x04, 0x00, 0x05, 0xfc,
				0x01, 0xd8, 0x75, 0xcc, 0x90, 0x92, 0xc3, 0xd9, 0x93, 0xba, 0xcf, 0xe4, 0x1d, 0x7c, 0xed, 0x5d,
				0x8f, 0x46, 0xdf, 0xc2, 0x19, 0x0f,
			},
			map[string]bool{
				"84dfc6fb0e86cf29049d53041e2d55f863eacfd8": true,
				"be22a5c7d7b25c990d89d7c18382f0815f683f17": true,
				"bbd835f67c0ef19084d9b97e9219c1b38e66bd80": true,
			},
		},
	}
	for i, tc := range tests {
		runCase(fmt.Sprintf("Test %d", i), tc, t)
	}
}
