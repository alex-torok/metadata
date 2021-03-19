package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestSimpleMetadataTree(t *testing.T) {
	fullPath := "../test_data/simple_test_case"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

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
			value, err := tree.GetClosestValue(tt.path, tt.key)
			require.NoError(t, err)
			assert.Equal(t, value, starlark.MakeInt(tt.expected))
		})
	}
}

func TestImportMetadataTree(t *testing.T) {
	fullPath := "../test_data/import_file"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	tests := []struct {
		path     string
		key      string
		expected int
	}{
		// from root METADATA file via imported funciton
		{"someFile.txt", "minimum_coverage", 90},
	}

	for _, tt := range tests {
		t.Run(tt.path+":"+tt.key, func(t *testing.T) {
			value, err := tree.GetClosestValue(tt.path, tt.key)
			require.NoError(t, err)
			assert.Equal(t, starlark.MakeInt(tt.expected), value)
		})
	}
}

func TestFileListMetadataTree(t *testing.T) {
	fullPath := "../test_data/limit_with_file_list"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	value, err := tree.GetClosestValue("main.py", "minimum_coverage")
	require.NoError(t, err)
	assert.Equal(t, value, starlark.MakeInt(90))

	value, err = tree.GetClosestValue("other.py", "minimum_coverage")
	assert.Nil(t, value)
	assert.Error(t, err)
}

func TestGlobMetadataTree(t *testing.T) {
	fullPath := "../test_data/limit_with_globs"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	value, err := tree.GetClosestValue("main.cc", "minimum_coverage")
	require.NoError(t, err)
	assert.Equal(t, value, starlark.MakeInt(90))

	value, err = tree.GetClosestValue("other.py", "minimum_coverage")
	require.NoError(t, err)
	assert.Equal(t, value, starlark.MakeInt(90))

	value, err = tree.GetClosestValue("one/other.cc", "cool_factor")
	require.NoError(t, err)
	assert.Equal(t, value, starlark.MakeInt(100))

}

func TestVerticalMergeMetadataTree(t *testing.T) {
	fullPath := "../test_data/vertical_merge"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	assertIntValue := func(path, key string, expected int) {
		value, err := tree.GetMergedValue(path, key)
		require.NoError(t, err)
		assert.Equal(t, value, starlark.MakeInt(expected), "Expected %s(%s) to be %d, but got %s", key, path, expected, value.String())
	}

	assertIntValue("main.cc", "minimum_coverage_take_lower", 80)
	assertIntValue("main.cc", "minimum_coverage_take_upper", 80)

	assertIntValue("one/main.cc", "minimum_coverage_take_lower", 90)
	assertIntValue("one/main.cc", "minimum_coverage_take_upper", 80)

	assertIntValue("one/two/main.cc", "minimum_coverage_take_lower", 100)
	assertIntValue("one/two/main.cc", "minimum_coverage_take_upper", 80)

}

func TestHorizontalMergeMetadataTree(t *testing.T) {
	fullPath := "../test_data/horizontal_merge"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	value, err := tree.GetMergedValue("main.cc", "owners")
	require.NoError(t, err)
	assert.Equal(t,
		starlark.NewList([]starlark.Value{
			starlark.String("alice"),
			starlark.String("bob"),
		}),
		value,
	)

	value, err = tree.GetMergedValue("main.py", "owners")
	require.NoError(t, err)
	assert.Equal(t,
		starlark.NewList([]starlark.Value{
			starlark.String("alice"),
			starlark.String("bob"),
			starlark.String("carol"),
		}),
		value,
	)

}

func TestHorizontalAndVerticalMergeMetadataTree(t *testing.T) {
	fullPath := "../test_data/horizontal_and_vertical_merge"
	tree, err := NewEagerTree(fullPath, "METADATA")
	if err != nil {
		require.NoError(t, err, "Unexpected error")
	}

	value, err := tree.GetMergedValue("main.cc", "owners")
	require.NoError(t, err)
	assert.Equal(t,
		starlark.NewList([]starlark.Value{
			starlark.String("alice"),
		}),
		value,
	)

	value, err = tree.GetMergedValue("main.py", "owners")
	require.NoError(t, err)
	assert.Equal(t,
		starlark.NewList([]starlark.Value{
			starlark.String("alice"),
			starlark.String("bob"),
		}),
		value,
	)

	value, err = tree.GetMergedValue("one/main.py", "owners")
	require.NoError(t, err)
	assert.Equal(t,
		starlark.NewList([]starlark.Value{
			starlark.String("carol"),
			starlark.String("alice"),
			starlark.String("bob"),
		}),
		value,
	)
}

// func TestJsonExport(t *testing.T) {
// 	fullPath := "../test_data/json_export"
// 	tree, err := NewEagerTree(fullPath, "METADATA")
// 	if err != nil {
// 		require.NoError(t, err, "Unexpected error")
// 	}

// 	value, err := tree.GetMergedValue("one.txt", "key")
// 	require.NoError(t, err)
// 	assert.Equal(t, "[1,2,3]", ValueToJson(value))
// }
