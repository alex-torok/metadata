package metadata

import (
	"errors"
	"path/filepath"
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
		//entries: globalMetadataStore.get(file.pathRelativeToRoot),
	}, nil
}

func (p *Parser) starlarkLoadFunc(_ *starlark.Thread, module string) (starlark.StringDict, error) {
	if !strings.HasPrefix(module, "//") {
		return nil, errors.New("Cannot load module that does not start with '//'")
	}

	// strip leading "//"
	path := module[2:]

	result, ok := p.cache[path]
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
	p.cache[path] = nil

	fileContents, err := p.repo.ReadFile(path)
	if err != nil {
		return nil, err
	}

	threadName := path // filepath.Dir(path)
	// if threadName == "." {
	// 	threadName = ""
	// }

	thread := &starlark.Thread{
		Name: threadName,
		Load: p.starlarkLoadFunc,
	}

	predeclared := starlark.StringDict{
		"metadata": starlark.NewBuiltin("metadata", metadata_starlark_func),
		"glob":     starlark.NewBuiltin("glob", glob_starlark_func),
	}

	globals, execErr := starlark.ExecFile(thread, threadName, fileContents, predeclared)
	result = &execFileResult{globals, execErr}

	p.cache[path] = result

	return result.globals, result.err
}

func glob_starlark_func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	var pattern string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"pattern", &pattern,
	); err != nil {
		//TODO: Add some way to show the file name in this error?
		return nil, err
	}

	glob, err := NewGlob(pattern)
	if err != nil {
		return nil, err
	}

	return &StarlarkGlob{glob}, nil
}

func metadata_starlark_func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	var key string
	var value starlark.Value
	var filesArg starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"key", &key,
		"value", &value,
		"files?", &filesArg,
	); err != nil {
		//TODO: Add some way to show the file name in this error?
		return nil, err
	}

	fileMatchSet, err := handleFilesArg(filesArg, dirOfRelativePath(thread.Name))
	if err != nil {
		return nil, err
	}

	entry := Entry{
		key:          key,
		value:        value,
		fileMatchSet: fileMatchSet,
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

func handleFilesArg(filesArg starlark.Value, relativeDir string) (StringSet, error) {
	fileMatchSet := make(StringSet)
	if filesArg != nil {
		if filesArg.Type() != "list" {
			return nil, errors.New("files must be of list type")
		}

		asList := filesArg.(*starlark.List)
		for i := 0; i < asList.Len(); i++ {
			val := asList.Index(i)
			if val.Type() != "string" {
				return nil, errors.New("Only string types are allowed for the files arg")
			}
			// TODO: ban relative pathing up like ./../some_other

			// TODO: this is a bit sloppy to get the full path to the file
			// relative to the current metadata file.
			// There be dragons if:
			//   1. Someone `load`s another METADATA file
			//   2. we have lots of file patterns, this will use lots of memory
			fileMatchSet.Add(filepath.Join(relativeDir, val.(starlark.String).GoString()))
		}
	}
	return fileMatchSet, nil
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
