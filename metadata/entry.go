package metadata

import "go.starlark.net/starlark"

type StringSet map[string]struct{}

func (s StringSet) Add(val string) {
	s[val] = struct{}{}
}

func (s StringSet) Contains(val string) bool {
	_, ok := s[val]
	return ok
}

type FileMatchSet struct {
	exactMatches   StringSet
	patternMatches []*Glob
}

func (f FileMatchSet) Matches(val string) bool {
	if f.exactMatches.Contains(val) {
		return true
	}

	for _, p := range f.patternMatches {
		if p.Match(val) {
			return true
		}
	}

	return false
}

func (f FileMatchSet) IsEmpty() bool {
	return len(f.exactMatches) == 0 && len(f.patternMatches) == 0
}

type VerticalMergeFunc func(upper, lower starlark.Value) (starlark.Value, error)
type HorizontalMergeFunc func(upper, lower starlark.Value) (starlark.Value, error)

// TODO: add "applies to file" function
type Entry struct {
	key   string
	value starlark.Value

	// files that this metadata entry applies to. If empty, apply to all files
	// this contains the full path relative to the root of the repo of any files
	// that match
	fileMatchSet      *FileMatchSet
	mergeVertically   VerticalMergeFunc
	mergeHorizontally HorizontalMergeFunc
}
