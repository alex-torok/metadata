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
	entryMap map[string][]Entry
}

type NoMetadataFoundError struct {
	path string
	key  string
}

func (e NoMetadataFoundError) Error() string {
	return fmt.Sprintf("No '%s' metadata found for '%s'", e.key, e.path)
}

// GetMergedValue - get the value of a particular metadata type for a file
// merge the values with any upper values
func (m *MetadataTree) GetMergedValue(filePath string, metadataKey string) (starlark.Value, error) {
	metStack := m.getMetadataStack(filePath, metadataKey)
	if len(metStack) == 0 {
		return nil, NoMetadataFoundError{filePath, metadataKey}
	}
	// TODO: Implement a metadata type store that can hold these functions
	vertMergeFunc := metStack[0].mergeVertically

	valueStack, err := m.getValueStack(filePath, metadataKey)
	if err != nil {
		return nil, err
	} else if len(valueStack) == 0 {
		return nil, NoMetadataFoundError{filePath, metadataKey}
	}

	return mergeVerticalStack(valueStack, vertMergeFunc)
}

func mergeVerticalStack(stack []starlark.Value, mergeFunc VerticalMergeFunc) (starlark.Value, error) {
	lowerValue := stack[len(stack)-1]
	for i := len(stack) - 2; i >= 0; i-- {
		upperValue := stack[i]

		var err error
		lowerValue, err = mergeFunc(upperValue, lowerValue)
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
		if entries, ok := currentTree.entryMap[metadataKey]; ok {
			// TODO: Fix hacky hacky only taking the first entry
			entry := entries[0]
			if entry.isAppliedToFile(filePath) {
				stack = append(stack, entry)
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

func (m *MetadataTree) getValueStack(filePath string, metadataKey string) ([]starlark.Value, error) {
	stack := make([]starlark.Value, 0)

	currentTree := m
	for _, dirPart := range strings.Split(filePath, string(filepath.Separator)) {
		if entries, ok := currentTree.entryMap[metadataKey]; ok {
			val, err := m.resolveSiblingEntries(entries, filePath, metadataKey)
			if err != nil {
				return nil, err
			}
			stack = append(stack, val)
		}
		var nextSubtreeExists bool
		currentTree, nextSubtreeExists = currentTree.subTrees[dirPart]
		if !nextSubtreeExists {
			break
		}
	}

	return stack, nil
}

func (m MetadataTree) resolveSiblingEntries(entries []Entry, filePath string, metadataKey string) (starlark.Value, error) {
	// Find all entries that match the given file
	matchingEntries := make([]Entry, 0)
	for _, entry := range entries {
		if entry.isAppliedToFile(filePath) {
			matchingEntries = append(matchingEntries, entry)
		}
	}

	if len(matchingEntries) == 0 {
		return nil, NoMetadataFoundError{filePath, metadataKey}
	}

	// Merge the siblings
	leftValue := matchingEntries[0].value
	for i := 1; i < len(matchingEntries); i++ {
		right := matchingEntries[i]
		rightValue := right.value

		var err error
		leftValue, err = right.mergeHorizontally(leftValue, rightValue)
		if err != nil {
			return nil, err
		}
	}

	return leftValue, nil
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
		entryMap: make(map[string][]Entry),
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
			if prev, seen_key := tree.entryMap[entry.key]; seen_key {
				tree.entryMap[entry.key] = append(prev, entry)
			} else {
				tree.entryMap[entry.key] = []Entry{entry}
			}
		}
	}

	return rootTree
}
