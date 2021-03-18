package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestValueToJson(t *testing.T) {

	makeTuple := func(vals ...starlark.Value) starlark.Tuple {
		return vals
	}

	makeList := func(vals ...starlark.Value) starlark.Value {
		return starlark.NewList(vals)
	}

	tests := []struct {
		name    string
		value   starlark.Value
		want    string
		wantErr bool
	}{
		{"None Type", starlark.None, "null", false},
		{"integer", starlark.MakeInt(1234), "1234", false},
		{"float", starlark.Float(123.456), "123.456", false},
		{"string", starlark.String("abcdefg"), `"abcdefg"`, false},
		{"bool true", starlark.Bool(true), "true", false},
		{"bool false", starlark.Bool(false), "false", false},
		{"dict", func() starlark.Value {
			s := starlark.NewDict(0)
			require.NoError(t, s.SetKey(starlark.String("abc"), starlark.MakeInt(5)))
			require.NoError(t, s.SetKey(starlark.String("list"), makeList(starlark.MakeInt(5))))
			require.NoError(t, s.SetKey(starlark.String("none"), starlark.None))
			return s
		}(), `{"abc":5,"list":[5],"none":null}`, false},
		{"set", func() starlark.Value {
			s := starlark.NewSet(0)
			require.NoError(t, s.Insert(starlark.MakeInt(5)))
			require.NoError(t, s.Insert(starlark.MakeInt(10)))
			require.NoError(t, s.Insert(starlark.MakeInt(15)))
			return s
		}(), "[5,10,15]", false},
		{"int tuple",
			makeTuple(
				starlark.MakeInt(1),
				starlark.MakeInt(2),
				starlark.MakeInt(42)), "[1,2,42]", false},
		{"int list",
			makeList(
				starlark.MakeInt(1),
				starlark.MakeInt(2),
				starlark.MakeInt(42)), "[1,2,42]", false},
		{"mixed list",
			makeList(
				starlark.Bool(false),
				starlark.String("abcd"),
				starlark.MakeInt(42)), `[false,"abcd",42]`, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValueToJson(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValueToJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
