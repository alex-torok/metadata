package metadata

import (
	"errors"
	"fmt"
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
		"meta":     starlark.NewBuiltin("meta", meta_new_starlark_func),
		"metadata": starlark.NewBuiltin("metadata", metadata_starlark_func),
		"glob":     starlark.NewBuiltin("glob", glob_starlark_func),
	}

	globals, execErr := starlark.ExecFile(thread, threadName, fileContents, predeclared)
	result = &execFileResult{globals, execErr}

	p.cache[path] = result

	return result.globals, result.err
}

func meta_new_starlark_func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	var verticalMergeFunc starlark.Callable
	var horizontalMergeFunc starlark.Callable
	var key string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"vertical_merge?", &verticalMergeFunc,
		"horizontal_merge?", &horizontalMergeFunc,
		"key", &key,
	); err != nil {
		//TODO: Add some way to show the file name in this error?
		return nil, err
	}

	returnFunc := starlark.NewBuiltin("metadata", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		var value starlark.Value
		var filesArg starlark.Value
		if err := starlark.UnpackArgs(b.Name(), args, kwargs,
			"value", &value,
			"files?", &filesArg,
		); err != nil {
			return nil, err
		}

		fileMatchSet, err := handleFilesArg(filesArg, dirOfRelativePath(thread.Name))
		if err != nil {
			return nil, err
		}

		entry := Entry{
			key:               key,
			value:             value,
			fileMatchSet:      fileMatchSet,
			mergeVertically:   newVerticalMerger(verticalMergeFunc),
			mergeHorizontally: newHorizontalMerger(horizontalMergeFunc),
		}

		globalMetadataStore.addEntry(thread.Name, entry)
		return starlark.None, nil
	})

	return returnFunc, nil
}

func glob_starlark_func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	var pattern string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"pattern", &pattern,
	); err != nil {
		//TODO: Add some way to show the file name in this error?
		return nil, err
	}
	glob, err := NewGlobRelativeTo(pattern, dirOfRelativePath(thread.Name))
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

func handleFilesArg(filesArg starlark.Value, relativeDir string) (*FileMatchSet, error) {
	exactMatchSet := make(StringSet)
	globList := make([]*Glob, 0)
	if filesArg != nil {
		if filesArg.Type() != "list" {
			return nil, errors.New("files must be of list type")
		}

		asList := filesArg.(*starlark.List)
		for i := 0; i < asList.Len(); i++ {
			val := asList.Index(i)
			switch val := val.(type) {
			// TODO: this is a bit sloppy to get the full path to the file
			// relative to the current metadata file.
			// There be dragons if:
			//   1. Someone `load`s another METADATA file
			//   2. we have lots of file patterns, this will use lots of memory
			case starlark.String:
				// TODO: ban relative pathing up like ./../some_other
				exactMatchSet.Add(filepath.Join(relativeDir, val.GoString()))
			case *StarlarkGlob:
				// TODO: fix pattern to match relative to metadata file
				globList = append(globList, val.impl)
			default:
				return nil, errors.New("Only string and glob types are allowed for the files arg")
			}

		}
	}

	return &FileMatchSet{
		exactMatches:   exactMatchSet,
		patternMatches: globList,
	}, nil
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

func newVerticalMerger(vertMergeFunc starlark.Callable) VerticalMergeFunc {
	return func(upper, lower starlark.Value) (starlark.Value, error) {
		if vertMergeFunc == nil {
			return nil, fmt.Errorf("Cannot merge vertically. No vertical merge function defined for this metadata type")
		}

		thread := &starlark.Thread{
			Name: "Vertically Merging",
		}
		args := []starlark.Value{upper, lower}
		res, err := starlark.Call(thread, vertMergeFunc, args, []starlark.Tuple{})
		if err != nil {
			return nil, fmt.Errorf("Could not vertically merge upper(%v) and lower(%v): %v", upper, lower, err)
		}

		return res, nil
	}
}

func newHorizontalMerger(horizMergeFunc starlark.Callable) HorizontalMergeFunc {
	return func(left, right starlark.Value) (starlark.Value, error) {
		if horizMergeFunc == nil {
			return nil, fmt.Errorf("Cannot merge horizontally. No horizontal merge function defined for this metadata type")
		}

		thread := &starlark.Thread{
			Name: "Vertically Merging",
		}
		args := []starlark.Value{left, right}
		res, err := starlark.Call(thread, horizMergeFunc, args, []starlark.Tuple{})
		if err != nil {
			return nil, fmt.Errorf("Could not horizontally merge left(%v) and right(%v): %v", left, right, err)
		}

		return res, nil
	}
}
