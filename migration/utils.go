package migration

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type optionSet struct {
	elements map[string]struct{}
}

func (s *optionSet) Add(elements ...string) {
	for _, element := range elements {
		s.elements[element] = struct{}{}
	}
}

func (s *optionSet) Size() int {
	return len(s.elements)
}

func (s *optionSet) Elements() []string {
	keys := make([]string, 0, len(s.elements))
	for key := range s.elements {
		keys = append(keys, key)
	}
	return keys
}

// normalizePath join and normalize file path.
func normalizePath(path ...string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(path...)))
}

// alphaNum extract alpha and numbers from string [a-zA-Z0-9].
func alphaNum(s string, includes ...string) string {
	pattern := "[^a-zA-Z0-9" + strings.Join(includes, "") + "]"
	rx := regexp.MustCompile(pattern)
	return rx.ReplaceAllString(s, "")
}

// slugify make url friendly slug from strings.
// Only Alpha-Num characters compiled to final result.
func slugify(parts ...string) string {
	normalized := alphaNum(strings.Join(parts, " "), `\s\-`)
	rx := regexp.MustCompile(`[\s\-]+`)
	return rx.ReplaceAllString(strings.ToLower(normalized), "-")
}

// getFlag get flag from input command.
func getFlag(cmd *cobra.Command, name string) string {
	if v, err := cmd.Flags().GetString(name); err == nil {
		return v
	}
	return ""
}
