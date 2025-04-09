package query

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// parseVariadic returns the first value from the variadic parameter `vals` if it exists,
// otherwise it returns the default value `def`.
func parseVariadic[T any](fallback T, vals ...T) T {
	if len(vals) > 0 {
		return vals[0]
	}
	return fallback
}

// normalizePath normalizes the given path segments
// by joining them and converting to a slashed separator.
func normalizePath(path ...string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(path...)))
}

// toName converts a file path to
// a name by removing the root and extension.
func toName(path, root, ext string) string {
	if path == "" {
		return ""
	}

	path = strings.TrimPrefix(path, root)
	path = strings.TrimSuffix(path, ext)
	return normalizePath(path)
}

// parseQueries extracts named queries from the given SQL content.
// Query sections are defined using the format: "-- {query: name}"
func parseQueries(content string) (map[string]string, error) {
	var name, body string
	res := make(map[string]string)

	// Compile query regexp
	rx, err := regexp.Compile(`^\s*--\s*\{\s*(\w+):\s*([\w\s]+)\s*\}$`)
	if err != nil {
		return nil, err
	}

	// Parse query helper functions
	parseTag := func(content string) (string, string, bool) {
		matches := rx.FindStringSubmatch(content)
		if len(matches) == 3 {
			matches[2] = strings.TrimSpace(matches[2])
			if matches[2] != "" {
				return matches[1], matches[2], true
			}
		}
		return "", "", false
	}

	// Scan lines
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tag, query, isNew := parseTag(line)
		if isNew {
			// Save old
			if name != "" {
				res[name] = strings.TrimRight(body, "\n")
			}

			// Start new
			name = ""
			body = ""
			if tag == "query" {
				name = query
			}
		} else if line != "" && name != "" {
			body = body + line + "\n"
		}
	}

	if name != "" {
		res[name] = strings.TrimRight(body, "\n")
	}

	return res, nil
}
