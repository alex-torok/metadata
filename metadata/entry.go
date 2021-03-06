package metadata

import "go.starlark.net/starlark"

type Entry struct {
	key   string
	value starlark.Value
}
