package metadata

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
)

type StarlarkGlob struct {
	impl *Glob
}

// starlark.Value methods
func (s *StarlarkGlob) String() string        { return fmt.Sprintf("meta.glob(%s)", s.impl.pattern) }
func (s *StarlarkGlob) Type() string          { return "meta.glob" }
func (s *StarlarkGlob) Freeze()               {}
func (s *StarlarkGlob) Truth() starlark.Bool  { return starlark.True }
func (s *StarlarkGlob) Hash() (uint32, error) { return 0, errors.New("not hashable") }
