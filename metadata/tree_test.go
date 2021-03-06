package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleMetadataTree(t *testing.T) {
	// TODO: Add test cases.
	fullPath := "../test_data/simple_test_case"
	tree, err := TreeFromDir(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	var value int

	tests := []struct {
		path     string
		key      string
		expected int
	}{
		// from root METADATA file
		{"someFile.txt", "cool factor", 9001},
		{"someFile.txt", "minimum_coverage", 80},
		{"no/dir/file.txt", "minimum_coverage", 80},
		{"one/someFile.txt", "cool factor", 9001},

		// from lower METADATA file
		{"one/someFile.txt", "minimum_coverage", 90},
		{"one/other/deep/file/someFile.txt", "minimum_coverage", 90},
	}

	for _, tt := range tests {
		t.Run(tt.path+":"+tt.key, func(t *testing.T) {
			value, err = tree.GetClosestValue(tt.path, tt.key)
			require.NoError(t, err)
			assert.Equal(t, value, tt.expected)
		})
	}

}
