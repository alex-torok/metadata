package metadata

import (
	"fmt"

	"go.starlark.net/starlark"
)

func ParseAll(files []MetadataFile) ([]ParsedMetadataFile, error) {
	parsed := make([]ParsedMetadataFile, 0)
	for _, file := range files {
		p, err := ParseOne(file)
		if err != nil {
			return parsed, err
		}
		parsed = append(parsed, p)
	}
	return parsed, nil
}

func ParseOne(file MetadataFile) (ParsedMetadataFile, error) {
	fileContents, err := file.Contents()
	if err != nil {
		return ParsedMetadataFile{}, err
	}

	thread := &starlark.Thread{Name: file.path}

	predeclared := starlark.StringDict{
		"metadata": starlark.NewBuiltin("metadata", metadata_starlark_func),
	}

	_, execErr := starlark.ExecFile(thread, file.path, fileContents, predeclared)
	if execErr != nil {
		return ParsedMetadataFile{}, err
	}

	return ParsedMetadataFile{
		entries: globalMetadataStore.get(file.path),
	}, nil
}

func metadata_starlark_func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	fileBeingParsed := thread.Name
	// frame := thread.CallFrame(1)

	// pos := frame.Pos
	// fmt.Printf("metadata called from %d:%d %s\n", pos.Line, pos.Col, frame.Name)

	var key string
	var value int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "value", &value); err != nil {
		//TODO: Add some way to show the file name in this error?
		return nil, err
	}

	fmt.Printf("metadata called with key=%s, value=%d\n", key, value)
	entry := Entry{
		key:   key,
		value: value,
	}

	globalMetadataStore.addEntry(fileBeingParsed, entry)

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
