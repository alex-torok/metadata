package metadata

import (
	"go.starlark.net/starlark"
)

type ParseResult struct {
	file    MetadataFile
	entries []Entry
}

func ParseAll(files []MetadataFile) ([]ParseResult, error) {
	parsed := make([]ParseResult, 0)
	for _, file := range files {
		p, err := ParseOne(file)
		if err != nil {
			return parsed, err
		}
		parsed = append(parsed, p)
	}
	return parsed, nil
}

func ParseOne(file MetadataFile) (ParseResult, error) {
	fileContents, err := file.Contents()
	if err != nil {
		return ParseResult{}, err
	}

	threadName := file.path
	thread := &starlark.Thread{Name: threadName}

	predeclared := starlark.StringDict{
		"metadata": starlark.NewBuiltin("metadata", metadata_starlark_func),
	}

	_, execErr := starlark.ExecFile(thread, file.path, fileContents, predeclared)
	if execErr != nil {
		return ParseResult{}, err
	}

	return ParseResult{
		file:    file,
		entries: globalMetadataStore.get(threadName),
	}, nil
}

func metadata_starlark_func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	var key string
	var value int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "value", &value); err != nil {
		//TODO: Add some way to show the file name in this error?
		return nil, err
	}

	entry := Entry{
		key:   key,
		value: value,
	}

	globalMetadataStore.addEntry(thread.Name, entry)

	return starlark.None, nil
}

var globalMetadataStore metadataStore

func init() {
	globalMetadataStore = metadataStore{
		store: make(map[string][]Entry),
	}
}

//TODO: Make thread safe
type metadataStore struct {
	store map[string][]Entry
}

func (m *metadataStore) addEntry(path string, entry Entry) {
	if val, ok := m.store[path]; ok {
		val = append(val, entry)
		m.store[path] = val
	} else {
		val := make([]Entry, 1)
		val[0] = entry
		m.store[path] = val
	}
}

func (m *metadataStore) get(path string) []Entry {
	return m.store[path]
}
