package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlob_MatchNoStars(t *testing.T) {
	type args struct {
		str string
	}
	glob, err := NewGlob("somefile.txt")
	require.NoError(t, err)

	tests := []struct {
		pattern string
		want    bool
	}{
		{"somefile.txt", true},
		{"other/somefile.txt", false},
		{"omefile.txt", false},
		{"somefile.tx", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			assert.Equal(t, tt.want, glob.Match(tt.pattern))
		})
	}
}

func TestGlob_MatchOneStar(t *testing.T) {
	type args struct {
		str string
	}
	glob, err := NewGlob("*.txt")
	require.NoError(t, err)

	tests := []struct {
		pattern string
		want    bool
	}{
		{"somefile.txt", true},
		{"other/somefile.txt", false},
		{"readme.txt", true},
		{"readmetxt", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			assert.Equal(t, tt.want, glob.Match(tt.pattern))
		})
	}
}

func TestGlob_MatchTwoStar(t *testing.T) {
	type args struct {
		str string
	}
	glob, err := NewGlob("**/file.txt")
	require.NoError(t, err)

	tests := []struct {
		pattern string
		want    bool
	}{
		{"file.txt", false},
		{"other/file.txt", true},
		{"very/deep/file.txt", true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			assert.Equal(t, tt.want, glob.Match(tt.pattern))
		})
	}
}
