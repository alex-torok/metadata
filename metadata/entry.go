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
	fileMatchSet StringSet
}
