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

type Entry struct {
	key   string
	value starlark.Value

	// files that this metadata entry applies to. If empty, apply to all files
	// this contains the full path relative to the root of the repo of any files
	// that match
	fileMatchSet StringSet
}
