package migration

import "slices"

type sortableFiles []migrationFile

func (fs sortableFiles) Len() int {
	return len(fs)
}

func (fs sortableFiles) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

func (fs sortableFiles) Less(i, j int) bool {
	return fs[i].timestamp < fs[j].timestamp
}

func (fs sortableFiles) Copy() sortableFiles {
	return append([]migrationFile{}, fs...)
}

func (fs sortableFiles) Reverse() sortableFiles {
	result := fs.Copy()
	slices.Reverse(result)
	return result
}

func (fs sortableFiles) Filter(only, exclude []string) sortableFiles {
	skip := func(name string) bool {
		if name == "" ||
			(len(only) > 0 && !slices.Contains(only, name)) ||
			(len(exclude) > 0 && slices.Contains(exclude, name)) {
			return true
		}

		return false
	}

	result := make(sortableFiles, 0)
	for _, file := range fs {
		if !skip(file.name) {
			result = append(result, file)
		}
	}
	return result
}
