package migration

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// CreateMigrationFile creates a migration file in the specified root directory with the given name, extension, and optional stages.
// It also accepts optional stages to include in the migration file.
func CreateMigrationFile(root, name, ext string, stages ...string) error {
	// Ensure required parameters are provided.
	if root == "" || name == "" || ext == "" {
		return errors.New("root, name, and extension parameters are required")
	}

	// Normalize the path and generate the file name.
	root = normalizePath(root)
	name = generateFileName(name, ext)

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(root, os.ModeDir|0755); err != nil {
		return err
	}

	// Prepare content with the provided stages, defaulting to "main".
	content := make([]string, 0)
	if len(stages) == 0 {
		stages = []string{"main"}
	}
	for _, stage := range stages {
		content = append(content, fmt.Sprintf("-- { up: %s }\n\n", stage))
		content = append(content, fmt.Sprintf("-- { down: %s }\n\n", stage))
	}

	// Write the generated content to the file.
	if err := os.WriteFile(
		normalizePath(root, name),
		[]byte(strings.Join(content, "")),
		0644,
	); err != nil {
		return err
	}
	return nil
}

// generateFileName generates a migration file name with a Unix timestamp, slugified name, and specified extension.
func generateFileName(name, ext string) string {
	return fmt.Sprintf(
		"%s-%s.%s",
		strconv.FormatInt(time.Now().Unix(), 10),
		slugify(name), ext,
	)
}
