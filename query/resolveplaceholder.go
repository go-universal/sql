package query

import "strconv"

// PlaceholderResolver is a function to resolve placeholders in SQL queries.
type PlaceholderResolver func(idx int) string

// NumbericResolver returns a numeric SQL placeholder in the form of "$1", "$2", etc.
func NumbericResolver(idx int) string {
	return `$` + strconv.Itoa(idx)
}
