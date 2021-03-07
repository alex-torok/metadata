package metadata

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Glob struct {
	re      *regexp.Regexp
	pattern string
}

func (g Glob) Match(str string) bool {
	return g.re.MatchString(str)
}

func NewGlob(pattern string) (*Glob, error) {
	builder := strings.Builder{}

	err := handleDoubleStars(&builder, pattern)
	if err != nil {
		return nil, err
	}

	regexPattern := fmt.Sprintf("^%s$", builder.String())
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}
	return &Glob{
		re:      re,
		pattern: pattern,
	}, nil
}

func handleDoubleStars(builder *strings.Builder, pattern string) error {
	components := strings.Split(pattern, "**")

	err := checkDoubleStarViolations(components)
	if err != nil {
		return err
	}

	for _, c := range components[:len(components)-1] {
		handleSingleStars(builder, c)
		// Anything
		builder.WriteString(".*")
	}
	handleSingleStars(builder, components[len(components)-1])
	return nil
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

func checkDoubleStarViolations(components []string) error {
	if len(components) == 1 {
		return nil
	}
	for i, c := range components {
		isFirst := i == 0
		isLast := i == len(components)-1

		if isFirst && c == "" || isLast && c == "" {
			continue
		}

		if !isFirst && !strings.HasPrefix(c, "/") {
			return errors.New("Invalid ** component - must be at start, end, or between path separators (/**/)")
		}

		if !isLast && !strings.HasSuffix(c, "/") {
			return errors.New("Invalid ** component - must be at start, end, or between path separators (/**/)")
		}
	}
	return nil
}
