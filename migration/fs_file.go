package migration

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type migrationFile struct {
	timestamp   int64
	name        string
	extension   string
	upScripts   map[string]string
	downScripts map[string]string
}

// newMigrationFile parses the migration file's path and content, extracting metadata and SQL scripts.
func newMigrationFile(path, content string) *migrationFile {
	// Extract file details and SQL scripts for "up" and "down" stages.
	timestamp, name, ext, ok := parseFileName(filepath.Base(path))
	if !ok {
		return nil
	}

	return &migrationFile{
		timestamp:   timestamp,
		name:        name,
		extension:   ext,
		upScripts:   parseFileSections(content, "up"),
		downScripts: parseFileSections(content, "down"),
	}
}

// UpScript retrieves the "up" script for a specific stage.
func (f migrationFile) UpScript(stage string) (string, bool) {
	v, ok := f.upScripts[stage]
	return v, ok
}

// DownScript retrieves the "down" script for a specific stage.
func (f migrationFile) DownScript(stage string) (string, bool) {
	v, ok := f.downScripts[stage]
	return v, ok
}

// parseFileName extracts the timestamp, name, and extension from a file name.
// Returns the extracted values and true if successful, or zero values and false on failure.
func parseFileName(name string) (int64, string, string, bool) {
	// Regex to match the expected file name format.
	rx, err := regexp.Compile(`^(\d+)-([a-zA-Z0-9-]+)\.([a-zA-Z0-9]+)$`)
	if err != nil {
		return 0, "", "", false
	}

	matches := rx.FindStringSubmatch(name)
	if len(matches) != 4 {
		return 0, "", "", false
	}

	// Parse timestamp and return the file name components.
	timestamp, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, "", "", false
	}

	return timestamp, strings.ReplaceAll(matches[2], "-", " "), matches[3], true
}

// parseFileSections extracts SQL sections defined by the format "-- {section: name}".
func parseFileSections(content, section string) map[string]string {
	var name, body string
	res := make(map[string]string)

	// Regex to match section tags.
	rx := regexp.MustCompile(`^\s*--\s*\{\s*(\w+):\s*([\w\s]+)\s*\}$`)
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

	// Scan and parse the content line by line.
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tag, query, isNew := parseTag(line)
		if isNew {
			// Save previous section if it exists.
			if name != "" {
				res[name] = strings.TrimRight(body, "\n")
			}

			// Start a new section.
			name = ""
			body = ""
			if tag == section {
				name = query
			}
		} else if line != "" && name != "" {
			body = body + line + "\n"
		}
	}

	// Save the last section.
	if name != "" {
		res[name] = strings.TrimRight(body, "\n")
	}

	return res
}
