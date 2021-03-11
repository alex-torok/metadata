package metadata

import (
	"fmt"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
)

func TreeFromDir(root, metadataFilename string) (*MetadataTree, error) {
	r := Repo{
		Root:             root,
		MetadataFilename: metadataFilename,
	}

	files, err := r.MetadataFiles()
	if err != nil {
		return nil, err
	}

	parser := NewParser(&r)
	parsed, err := parser.ParseAll(files)
	if err != nil {
		return nil, err
	}

	return NewMetadataTree(parsed), nil
}

// MetadataTree is a tree matching the structure of the filesystem in a repo,
// where the entries in a tree are the metadata entries located in that folder's
// METADATA file
type MetadataTree struct {
	subTrees map[string]*MetadataTree
	entries  []Entry
	entryMap map[string]Entry
}

type NoMetadataFoundError struct {
	path string
	key  string
}

func (e NoMetadataFoundError) Error() string {
	return fmt.Sprintf("No '%s' metadata found for '%s'", e.key, e.path)
}

func (m *MetadataTree) GetMergedValue(filePath string, metadataKey string) (starlark.Value, error) {
	stack := m.getMetadataStack(filePath, metadataKey)
	if len(stack) == 0 {
		return nil, NoMetadataFoundError{filePath, metadataKey}
	}

	lowerValue := stack[len(stack)-1].value
	for i := len(stack) - 2; i >= 0; i-- {
		upper := stack[i]
		upperValue := upper.value

		var err error
		lowerValue, err = upper.mergeVertically(upperValue, lowerValue)
		if err != nil {
			return nil, err
		}
	}
	return lowerValue, nil
}

func (m *MetadataTree) GetClosestValue(filePath string, metadataKey string) (starlark.Value, error) {
	stack := m.getMetadataStack(filePath, metadataKey)
	if len(stack) == 0 {
		return nil, NoMetadataFoundError{filePath, metadataKey}
	}
	return stack[len(stack)-1].value, nil

}

func (m *MetadataTree) getMetadataStack(filePath string, metadataKey string) []Entry {
	stack := make([]Entry, 0)

	currentTree := m
	for _, dirPart := range strings.Split(filePath, string(filepath.Separator)) {
		if val, ok := currentTree.entryMap[metadataKey]; ok {
			if val.fileMatchSet.IsEmpty() || val.fileMatchSet.Matches(filePath) {
				stack = append(stack, val)
			}
		}
		var nextSubtreeExists bool
		currentTree, nextSubtreeExists = currentTree.subTrees[dirPart]
		if !nextSubtreeExists {
			break
		}
	}

	return stack
}

func (m *MetadataTree) get(dirName string) *MetadataTree {
	subTree, exists := m.subTrees[dirName]
	if !exists {
		subTree = newTree()
		m.subTrees[dirName] = subTree
	}
	return subTree
}

func newTree() *MetadataTree {
	return &MetadataTree{
		subTrees: make(map[string]*MetadataTree),
		entries:  make([]Entry, 0),
		entryMap: make(map[string]Entry),
	}
}

func getTree(rootTree *MetadataTree, result ParseResult) *MetadataTree {
	thisTree := rootTree
	dirs := result.file.Dirs()
	for _, dir := range dirs {
		thisTree = thisTree.get(dir)
	}
	return thisTree
}

func NewMetadataTree(results []ParseResult) *MetadataTree {
	rootTree := newTree()
	for _, result := range results {
		tree := getTree(rootTree, result)
		tree.entries = result.entries
		for _, entry := range tree.entries {
			//TODO: This overwrites any duplicated metadata entry key. Implement horizontal flattening.
			tree.entryMap[entry.key] = entry
		}
	}

	return rootTree
}
