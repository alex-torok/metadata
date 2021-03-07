package metadata

import (
	"fmt"
	"regexp"
	"strings"
)

type Glob struct {
	re *regexp.Regexp
}

func (g Glob) Match(str string) bool {
	return g.re.MatchString(str)
}

func NewGlob(pattern string) (*Glob, error) {
	builder := strings.Builder{}

	handleDoubleStars(&builder, pattern)

	regexPattern := fmt.Sprintf("^%s$", builder.String())
	fmt.Println(regexPattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}
	return &Glob{
		re: re,
	}, nil
}

func handleDoubleStars(builder *strings.Builder, pattern string) {
	components := strings.Split(pattern, "**")
	for _, c := range components[:len(components)-1] {
		handleSingleStars(builder, c)
		// Anything
		builder.WriteString(".*")
	}
	handleSingleStars(builder, components[len(components)-1])
}

func handleSingleStars(builder *strings.Builder, pattern string) {
	components := strings.Split(pattern, "*")
	for _, c := range components[:len(components)-1] {
		builder.WriteString(regexp.QuoteMeta(c))
		// Anything that isn't a path separator
		builder.WriteString("[^/]*")
	}
	builder.WriteString(regexp.QuoteMeta(components[len(components)-1]))
}
