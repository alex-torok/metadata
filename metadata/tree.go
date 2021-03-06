package metadata

type MetadataTree struct {
	subTrees map[string]*MetadataTree
	entries  []Entry
}

func (m MetadataTree) get(dirName string) *MetadataTree {
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
	}

	return rootTree
}
