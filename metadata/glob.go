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

	components := strings.Split(pattern, "*")
	for _, c := range components[:len(components)-1] {
		builder.WriteString(regexp.QuoteMeta(c))
		fmt.Println(regexp.QuoteMeta(c))
		// Anything that isn't a path separator
		builder.WriteString("[^/]*")
	}
	builder.WriteString(regexp.QuoteMeta(components[len(components)-1]))

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
