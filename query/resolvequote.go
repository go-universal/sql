package query

import "fmt"

// QuoteResolver is a function to resolve quote identity in SQL queries.
type QuoteResolver func(identity string) string

// BacktickResolver wrap identity with backtick.
func BacktickResolver(i string) string {
	return fmt.Sprintf("`%s`", i)
}

// DoubleQuoteResolver wrap identity with double-quote.
func DoubleQuoteResolver(i string) string {
	return fmt.Sprintf(`"%s"`, i)
}
