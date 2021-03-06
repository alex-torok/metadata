package metadata

import "go.starlark.net/starlark"

type Entry struct {
	key   string
	value starlark.Value

	// files that this metadata entry applies to. If empty, apply to all files
	filesThisAppliesTo []string
}
