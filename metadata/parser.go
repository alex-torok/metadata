package metadata

import (
	"errors"
	"strings"

	"go.starlark.net/starlark"
)

type ParseResult struct {
	file    MetadataFile
	entries []Entry
}

type execFileResult struct {
	globals starlark.StringDict
	err     error
}

type Parser struct {
	cache map[string]*execFileResult
	repo  *Repo
}

func NewParser(repo *Repo) Parser {
	return Parser{
		cache: make(map[string]*execFileResult),
		repo:  repo,
	}
}

func (p *Parser) ParseAll(files []MetadataFile) ([]ParseResult, error) {
	parsed := make([]ParseResult, 0)
	for _, file := range files {
		p, err := p.ParseOne(file)
		if err != nil {
			return parsed, err
		}
		parsed = append(parsed, p)
	}
	return parsed, nil
}

func (p *Parser) ParseOne(file MetadataFile) (ParseResult, error) {

	_, execErr := p.starlarkLoadFunc(nil, "//"+file.pathRelativeToRoot)
	if execErr != nil {
		return ParseResult{}, execErr
	}

	return ParseResult{
		file:    file,
		entries: globalMetadataStore.get(file.pathRelativeToRoot),
	}, nil
}

func (p *Parser) starlarkLoadFunc(_ *starlark.Thread, module string) (starlark.StringDict, error) {
	if !strings.HasPrefix(module, "//") {
		return nil, errors.New("Cannot load module that does not start with '//'")
	}

	// strip leading "//"
	filepath := module[2:]

	result, ok := p.cache[filepath]
	if result != nil {
		return result.globals, result.err
	}

	// If result is nil, and it was put in the cache, then we've already started trying
	// to load this file.
	if ok {
		return nil, errors.New("Cycle detected in load graph")
	}

	// Start actually loading the module. Mark that load is in progress by adding
	// nil to the cache
	p.cache[filepath] = nil

	fileContents, err := p.repo.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	threadName := filepath
	thread := &starlark.Thread{
		Name: threadName,
		Load: p.starlarkLoadFunc,
	}

	predeclared := starlark.StringDict{
		"metadata": starlark.NewBuiltin("metadata", metadata_starlark_func),
	}

	globals, execErr := starlark.ExecFile(thread, threadName, fileContents, predeclared)
	result = &execFileResult{globals, execErr}

	p.cache[filepath] = result

	return result.globals, result.err
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
